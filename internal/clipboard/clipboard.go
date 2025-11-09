package clipboard

import (
	"github.com/atotto/clipboard"
)

func Copy(content string) error {
	return clipboard.WriteAll(content)
}
