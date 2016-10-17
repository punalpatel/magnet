package magnet

import (
	"io"
	"os"

	"github.com/fatih/color"
	isatty "github.com/mattn/go-isatty"
)

var (
	// output is the writer that magnet writes its output to
	output io.Writer = os.Stdout

	redSprintf   = color.New(color.FgRed).SprintfFunc()
	greenSprintf = color.New(color.FgGreen).SprintfFunc()
)

// SetOutput can be used to customize where magnet writes its output.
// Magnet output is written to stdout by default.
func SetOutput(w io.Writer) {
	output = w

	// don't let ANSI escape codes ruin our output if
	// we're not writing to stdout
	if output == os.Stdout {
		color.NoColor = !isatty.IsTerminal(os.Stdout.Fd())
	} else {
		color.NoColor = true
	}
}
