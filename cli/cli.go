package cli

import (
	"fmt"
	"strings"
)

const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
)

// map of color ansii codes
var colors = map[string]string{
	"black":  "\033[30m",
	"red":    "\033[31m",
	"green":  "\033[32m",
	"yellow": "\033[33m",
	"blue":   "\033[34m",
	"purple": "\033[35m",
	"cyan":   "\033[36m",
	"white":  "\033[37m",
}

// PrintColor prints a given text with a specified color
// the following colors are supported: red, green, yellow, blue, purple, cyan and white
func PrintColor(text string, color string, newLine bool) {
	color = strings.ToLower(color)
	newLineString := ""

	if newLine {
		newLineString = "\n"
	}
	fmt.Printf("%s%s%s%s", colors[color], text, Reset, newLineString)
}

// print a text as bold
func PrintBold(text string, newLine bool) {
	newLineString := ""

	if newLine {
		newLineString = "\n"
	}
	fmt.Printf("%s%s%s%s", Bold, text, Reset, newLineString)
}

// print a text as bold and colored
func PrintBoldAndColor(text string, color string, newLine bool) {
	color = strings.ToLower(color)
	newLineString := ""

	if newLine {
		newLineString = "\n"
	}
	fmt.Printf("%s%s%s%s%s", colors[color], Bold, text, Reset, newLineString)
}

// returns a text colored and bold
func GetBoldAndColor(text string, color string, newLine bool) string {
	color = strings.ToLower(color)
	newLineString := ""

	if newLine {
		newLineString = "\n"
	}
	return fmt.Sprintf("%s%s%s%s%s", colors[color], Bold, text, Reset, newLineString)
}
