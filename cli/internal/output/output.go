// Package output provides consistent, colored terminal output for all commands.
package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	green = color.New(color.FgGreen)
	red   = color.New(color.FgRed)
	cyan  = color.New(color.FgCyan)
	dim   = color.New(color.Faint)
	bold  = color.New(color.Bold)
)

// Success prints "✓ <msg>" in green.
func Success(format string, a ...any) {
	green.Printf("✓ "+format+"\n", a...)
}

// Fail prints "✗ <msg>" in red to stderr.
func Fail(format string, a ...any) {
	red.Fprintf(os.Stderr, "✗ "+format+"\n", a...)
}

// Hint prints an indented hint line in dim text.
func Hint(format string, a ...any) {
	dim.Printf("  hint: "+format+"\n", a...)
}

// Step prints an in-progress action line (before a blocking call).
func Step(format string, a ...any) {
	fmt.Printf(format+"...\n", a...)
}

// Info prints a plain informational line.
func Info(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
}

// Header prints a bold section title.
func Header(title string) {
	bold.Println(title)
}

// Field prints a key-value pair indented under a header.
func Field(key, value string) {
	fmt.Printf("  %-26s %s\n", key+":", dim.Sprint(value))
}

// Address prints a key-address pair, highlighting the address in cyan.
func Address(key, addr string) {
	fmt.Printf("  %-26s %s\n", key+":", cyan.Sprint(addr))
}
