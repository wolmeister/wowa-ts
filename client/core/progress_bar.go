package core

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
)

func CreateProgressBar(max int, description string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions(
		max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "#",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	return bar
}
