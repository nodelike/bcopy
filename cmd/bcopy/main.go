package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nodelike/bcopy/internal/analyzer"
	"github.com/nodelike/bcopy/internal/clipboard"
	"github.com/nodelike/bcopy/internal/collector"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	noGitignore    bool
	excludeTests   bool
	customExcludes []string
	allowedExts    []string
	path           string
	maxDepth       int
	thresholdMB    float64
	hardMaxMB      float64
	maxFileSizeMB  float64
	dryRun         bool
	outputFile     string
)

var rootCmd = &cobra.Command{
	Use:   "bcopy [path]",
	Short: "Bulk copy codebase files to clipboard",
	Long: `bcopy is a tool for copying multiple files from your codebase to clipboard,
with smart filtering and git repository support.`,
	Version: "1.0.2",
	Args:    cobra.MaximumNArgs(1),
	Run:     runBcopy,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .bcopy.yaml)")
	rootCmd.Flags().BoolVar(&noGitignore, "no-gitignore", false, "Ignore .gitignore patterns (always-excluded patterns still apply)")
	rootCmd.Flags().BoolVar(&excludeTests, "exclude-tests", false, "Exclude test files (_test.go, test/, tests/, *.test.*, *.spec.*)")
	rootCmd.Flags().StringArrayVar(&customExcludes, "exclude", []string{}, "Additional exclusion pattern (can be repeated)")
	rootCmd.Flags().StringArrayVar(&allowedExts, "ext", []string{}, "Override allowed file extensions (can be repeated)")
	rootCmd.Flags().IntVar(&maxDepth, "max-depth", 0, "Maximum directory traversal depth (0 = unlimited)")
	rootCmd.Flags().Float64Var(&thresholdMB, "threshold", 1.0, "Size warning threshold in MB")
	rootCmd.Flags().Float64Var(&hardMaxMB, "hard-max", 50.0, "Hard maximum total size in MB (aborts if exceeded)")
	rootCmd.Flags().Float64Var(&maxFileSizeMB, "max-file-size", 10.0, "Maximum individual file size in MB")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print output to stdout instead of copying to clipboard")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write output to file instead of clipboard")

	viper.BindPFlag("no-gitignore", rootCmd.Flags().Lookup("no-gitignore"))
	viper.BindPFlag("exclude-tests", rootCmd.Flags().Lookup("exclude-tests"))
	viper.BindPFlag("exclude", rootCmd.Flags().Lookup("exclude"))
	viper.BindPFlag("ext", rootCmd.Flags().Lookup("ext"))
	viper.BindPFlag("max-depth", rootCmd.Flags().Lookup("max-depth"))
	viper.BindPFlag("threshold", rootCmd.Flags().Lookup("threshold"))
	viper.BindPFlag("hard-max", rootCmd.Flags().Lookup("hard-max"))
	viper.BindPFlag("max-file-size", rootCmd.Flags().Lookup("max-file-size"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".bcopy")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func runBcopy(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		path = args[0]
	} else {
		path = "."
	}

	if path == "." {
		var err error
		path, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get current directory: %v\n", err)
			os.Exit(1)
		}
	}

	if err := analyzer.ValidatePath(path); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check if it's a git repo and prompt if not
	isGitRepo := analyzer.IsGitRepo(path)
	if !isGitRepo {
		fmt.Fprintf(os.Stderr, "\033[33m⚠️  Warning: %s is not in a git repository\033[0m\n", path)
		fmt.Fprintln(os.Stderr, "bcopy works best in git repos but can run anywhere.")
		fmt.Fprint(os.Stderr, "\033[33mPress Enter to continue or Ctrl+C to cancel...\033[0m ")
		
		reader := bufio.NewReader(os.Stdin)
		_, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nCanceled by user\n")
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "")
	}

	if shouldWarn, warning := analyzer.ShouldWarnLargeDirectory(path); shouldWarn {
		fmt.Fprintln(os.Stderr, warning)
	}

	if !cmd.Flags().Changed("no-gitignore") {
		noGitignore = viper.GetBool("no-gitignore")
	}

	if !cmd.Flags().Changed("exclude-tests") {
		excludeTests = viper.GetBool("exclude-tests")
	}

	if len(customExcludes) == 0 {
		customExcludes = viper.GetStringSlice("exclude")
	}

	if len(allowedExts) == 0 {
		allowedExts = viper.GetStringSlice("ext")
	}

	if maxDepth == 0 {
		maxDepth = viper.GetInt("max-depth")
	}

	if !cmd.Flags().Changed("threshold") {
		if viper.IsSet("threshold") {
			thresholdMB = viper.GetFloat64("threshold")
		}
	}

	if !cmd.Flags().Changed("hard-max") {
		if viper.IsSet("hard-max") {
			hardMaxMB = viper.GetFloat64("hard-max")
		}
	}

	if !cmd.Flags().Changed("max-file-size") {
		if viper.IsSet("max-file-size") {
			maxFileSizeMB = viper.GetFloat64("max-file-size")
		}
	}

	filter := analyzer.NewFilter(allowedExts, customExcludes, !noGitignore, excludeTests)

	if !noGitignore && isGitRepo {
		repoRoot, err := analyzer.GetRepoRoot(path)
		if err == nil {
			filter.LoadGitignore(repoRoot)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nReceived interrupt signal, canceling...")
		cancel()
	}()

	result, err := collector.Collect(ctx, path, filter, maxDepth, maxFileSizeMB)
	if err != nil {
		if err == context.Canceled {
			fmt.Fprintln(os.Stderr, "\nCollection canceled by user")
			os.Exit(130)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if result.FileCount == 0 {
		fmt.Fprintln(os.Stderr, "\n\033[31m❌ No files found matching the criteria\033[0m")
		os.Exit(0)
	}

	sizeMB := float64(result.TotalSize) / (1024 * 1024)
	fmt.Fprintf(os.Stderr, "\n\033[35m✨ Found \033[1m%d files\033[0m\033[35m (\033[1m%.2f MB\033[0m\033[35m)\033[0m\n", result.FileCount, sizeMB)

	// Check hard maximum
	if sizeMB > hardMaxMB {
		fmt.Fprintf(os.Stderr, "\n\033[31m❌ Error: Total size (%.2f MB) exceeds hard maximum (%.2f MB)\033[0m\n", sizeMB, hardMaxMB)
		fmt.Fprintln(os.Stderr, "This is a safety limit to prevent clipboard overflow.")
		fmt.Fprintf(os.Stderr, "Use --hard-max to increase or --output to write to a file instead.\n")
		os.Exit(1)
	}

	if sizeMB > thresholdMB {
		fmt.Fprintf(os.Stderr, "\n\033[33m⚠️  Warning: Total size (%.2f MB) exceeds threshold (%.2f MB)\033[0m\n", sizeMB, thresholdMB)
		fmt.Fprint(os.Stderr, "\033[33mContinue copying to clipboard? (y/N): \033[0m")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(os.Stderr, "Canceled by user")
			os.Exit(0)
		}
	}

	markdown := collector.FormatAsMarkdown(result)

	// Handle different output modes
	if dryRun {
		fmt.Println(markdown)
		return
	}

	if outputFile != "" {
		fmt.Fprintf(os.Stderr, "\033[36m📝 Writing to file...\033[0m ")
		if err := os.WriteFile(outputFile, []byte(markdown), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "\n\033[31m❌ Error writing to file: %v\033[0m\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "\033[32m✓\033[0m")
		fmt.Fprintf(os.Stderr, "\033[1m\033[32m✅ Successfully written to %s!\033[0m\n", outputFile)
		return
	}

	fmt.Fprint(os.Stderr, "\033[36m📋 Copying to clipboard...\033[0m ")

	if err := clipboard.Copy(markdown); err != nil {
		fmt.Fprintf(os.Stderr, "\n\033[31m❌ Error copying to clipboard: %v\033[0m\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "\033[32m✓\033[0m")
	fmt.Fprintln(os.Stderr, "\033[1m\033[32m✅ Successfully copied to clipboard!\033[0m")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
