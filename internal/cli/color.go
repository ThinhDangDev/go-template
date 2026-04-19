package cli

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
)

var colorEnabled = true

// DisableColor turns off ANSI colors for CLI output.
func DisableColor() {
	colorEnabled = false
}

// IsColorEnabled reports whether ANSI colors should be used.
func IsColorEnabled() bool {
	if !colorEnabled {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func colorize(color, text string) string {
	if !IsColorEnabled() {
		return text
	}
	return color + text + colorReset
}

// Header prints a section header.
func Header(format string, args ...interface{}) {
	fmt.Println(colorize(colorBold, fmt.Sprintf(format, args...)))
}

// Step prints a numbered step message.
func Step(step, total int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	prefix := fmt.Sprintf("[%d/%d]", step, total)
	fmt.Println(colorize(colorBlue, prefix) + " " + msg)
}

// Success prints a success message.
func Success(format string, args ...interface{}) {
	fmt.Println(colorize(colorGreen, "OK") + " " + fmt.Sprintf(format, args...))
}

// Warning prints a warning message.
func Warning(format string, args ...interface{}) {
	fmt.Println(colorize(colorYellow, "WARN") + " " + fmt.Sprintf(format, args...))
}

// Errorf prints an error message to stderr.
func Errorf(format string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, colorize(colorRed, "ERR")+" "+fmt.Sprintf(format, args...))
}

// Info prints an informational message.
func Info(format string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(format, args...))
}
