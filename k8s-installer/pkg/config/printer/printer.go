// Package printer carries a stdout fd. It helps caller to print message on console
// even if it has set the log fd redirect.
package printer

import (
	"fmt"
	"io"
	"os"
)

var (
	// Printer is global StdPrinter.
	Printer = &StdPrinter{Out: os.Stdout}
)

// StdPrinter outputs info to console directly.
type StdPrinter struct {
	Out io.Writer
}

// Print outputs info to console directly.
func (sp *StdPrinter) Print(msg string) {
	if sp.Out != nil {
		fmt.Fprint(sp.Out, msg)
	}
}

// Println outputs info to console directly.
func (sp *StdPrinter) Println(msg string) {
	if sp.Out != nil {
		fmt.Fprintln(sp.Out, msg)
	}
}

// Printf formats according to a format specifier.
func (sp *StdPrinter) Printf(format string, a ...interface{}) {
	if sp.Out != nil {
		fmt.Fprintf(sp.Out, format+"\n", a...)
	}
}

// Print outputs info to console directly.
func Print(msg string) {
	if Printer.Out != nil {
		fmt.Fprint(Printer.Out, msg)
	}
}

// Println outputs info to console directly.
func Println(msg string) {
	if Printer.Out != nil {
		fmt.Fprintln(Printer.Out, msg)
	}
}

// Printf formats according to a format specifier.
func Printf(format string, a ...interface{}) {
	if Printer.Out != nil {
		fmt.Fprintf(Printer.Out, format+"\n", a...)
	}
}
