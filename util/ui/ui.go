// Package ui will provide hooks into STDOUT, STDERR and STDIN. It will also
// handle translation as necessary.
//
// This package is explicitly designed for the CF CLI and is *not* to be used
// by any package outside of the commands package.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/cli/util/configv3"
	"github.com/fatih/color"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/vito/go-interact/interact"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	red   color.Attribute = color.FgRed
	green                 = color.FgGreen
	// yellow                         = color.FgYellow
	// magenta                        = color.FgMagenta
	cyan           = color.FgCyan
	white          = color.FgWhite
	defaultFgColor = 38
)

//go:generate counterfeiter . Config

// Config is the UI configuration
type Config interface {
	// ColorEnabled enables or disabled color
	ColorEnabled() configv3.ColorSetting
	// Locale is the language to translate the output to
	Locale() string
}

//go:generate counterfeiter . TranslatableError

// TranslatableError it wraps the error interface adding a way to set the
// translation function on the error
type TranslatableError interface {
	// Returns back the untranslated error string
	Error() string
	Translate(func(string, ...interface{}) string) string
}

//go:generate counterfeiter . LogMessage

// LogMessage is a log response representing one to many joined lines of a log
// message.
type LogMessage interface {
	Message() string
	Type() string
	Timestamp() time.Time
	SourceType() string
	SourceInstance() string
}

// UI is interface to interact with the user
type UI struct {
	// In is the input buffer
	In io.Reader
	// Out is the output buffer
	Out io.Writer
	// Err is the error buffer
	Err io.Writer

	colorEnabled configv3.ColorSetting
	translate    i18n.TranslateFunc

	terminalLock *sync.Mutex
	fileLock     *sync.Mutex

	TimezoneLocation *time.Location
}

// NewUI will return a UI object where Out is set to STDOUT, In is set to
// STDIN, and Err is set to STDERR
func NewUI(c Config) (*UI, error) {
	translateFunc, err := GetTranslationFunc(c)
	if err != nil {
		return nil, err
	}

	location := time.Now().Location()

	return &UI{
		In:               os.Stdin,
		Out:              color.Output,
		Err:              os.Stderr,
		colorEnabled:     c.ColorEnabled(),
		translate:        translateFunc,
		terminalLock:     &sync.Mutex{},
		fileLock:         &sync.Mutex{},
		TimezoneLocation: location,
	}, nil
}

// NewTestUI will return a UI object where Out, In, and Err are customizable,
// and colors are disabled
func NewTestUI(in io.Reader, out io.Writer, err io.Writer) *UI {
	return &UI{
		In:               in,
		Out:              out,
		Err:              err,
		colorEnabled:     configv3.ColorDisabled,
		translate:        translationWrapper(i18n.IdentityTfunc()),
		terminalLock:     &sync.Mutex{},
		fileLock:         &sync.Mutex{},
		TimezoneLocation: time.UTC,
	}
}

// TranslateText passes the template through an internationalization function
// to translate it to a pre-configured language, and returns the template with
// templateValues substituted in. Only the first map in templateValues is used.
func (ui *UI) TranslateText(template string, templateValues ...map[string]interface{}) string {
	return ui.translate(template, getFirstSet(templateValues))
}

// UserFriendlyDate converts the time to UTC and then formats it to ISO8601.
func (ui *UI) UserFriendlyDate(input time.Time) string {
	return input.Local().Format("Mon 02 Jan 15:04:05 MST 2006")
}

// RequestLoggerTerminalDisplay returns a RequestLoggerTerminalDisplay that
// cannot overwrite another RequestLoggerTerminalDisplay or the current
// display.
func (ui *UI) RequestLoggerTerminalDisplay() *RequestLoggerTerminalDisplay {
	return newRequestLoggerTerminalDisplay(ui, ui.terminalLock)
}

// RequestLoggerFileWriter returns a RequestLoggerFileWriter that cannot
// overwrite another RequestLoggerFileWriter.
func (ui *UI) RequestLoggerFileWriter(filePaths []string) *RequestLoggerFileWriter {
	return newRequestLoggerFileWriter(ui, ui.fileLock, filePaths)
}

// DisplayOK outputs a bold green translated "OK" to UI.Out.
func (ui *UI) DisplayOK() {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Out, "%s\n", ui.addFlavor(ui.TranslateText("OK"), green, true))
}

// DisplayNewline outputs a newline to UI.Out.
func (ui *UI) DisplayNewline() {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Out, "\n")
}

// DisplayBoolPrompt outputs the prompt and waits for user input. It only
// allows for a boolean response. A default boolean response can be set with
// defaultResponse.
func (ui *UI) DisplayBoolPrompt(defaultResponse bool, template string, templateValues ...map[string]interface{}) (bool, error) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	response := defaultResponse
	interactivePrompt := interact.NewInteraction(ui.TranslateText(template, templateValues...) + ui.addFlavor(">>", cyan, true))
	interactivePrompt.Input = ui.In
	interactivePrompt.Output = ui.Out
	err := interactivePrompt.Resolve(&response)
	return response, err
}

// DisplayNonWrappingTable outputs a matrix of strings as a table to UI.Out. Prefix will
// be prepended to each row and padding adds the specified number of spaces
// between columns.
func (ui *UI) DisplayNonWrappingTable(prefix string, table [][]string, padding int) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	if len(table) == 0 {
		return
	}

	var columnPadding []int

	rows := len(table)
	columns := len(table[0])
	for col := 0; col < columns; col++ {
		var max int
		for row := 0; row < rows; row++ {
			if strLen := runewidth.StringWidth(table[row][col]); max < strLen {
				max = strLen
			}
		}
		columnPadding = append(columnPadding, max+padding)
	}

	for row := 0; row < rows; row++ {
		fmt.Fprintf(ui.Out, prefix)
		for col := 0; col < columns; col++ {
			var addedPadding int
			if col+1 != columns {
				addedPadding = columnPadding[col] - runewidth.StringWidth(table[row][col])
			}
			fmt.Fprintf(ui.Out, "%s%s", table[row][col], strings.Repeat(" ", addedPadding))
		}
		fmt.Fprintf(ui.Out, "\n")
	}
}

// DisplayTable outputs a matrix of strings as a table to UI.Out. Prefix will
// be prepended to each row and padding adds the specified number of spaces
// between columns.  The final columns may wrap to multiple lines but will
// still be confined to the last column.  Wrapping will occur on word boundaries.
func (ui *UI) DisplayTable(prefix string, table [][]string, padding int) {
	rows := len(table)
	if rows == 0 {
		return
	}

	columns := len(table[0])

	// if we don't want to wrap the table columns
	if columns < 2 || !terminal.IsTerminal(int(os.Stdout.Fd())) {
		ui.DisplayNonWrappingTable(prefix, table, padding)
		return
	}

	terminalWidth, _, err := terminal.GetSize(int(os.Stdout.Fd()))

	if err != nil {
		ui.DisplayNonWrappingTable(prefix, table, padding)
		return
	}

	ui.DisplayWrappingTableWithWidth(prefix, table, padding, terminalWidth)
}

func (ui *UI) DisplayWrappingTableWithWidth(prefix string, table [][]string, padding int, terminalWidth int) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	var columnPadding []int

	rows := len(table)
	columns := len(table[0])

	for col := 0; col < columns-1; col++ {
		var max int
		for row := 0; row < rows; row++ {
			if strLen := runewidth.StringWidth(table[row][col]); max < strLen {
				max = strLen
			}
		}
		columnPadding = append(columnPadding, max+padding)
	}

	spilloverPadding := len(prefix) + sum(columnPadding)
	lastColumnWidth := terminalWidth - spilloverPadding

	for row := 0; row < rows; row++ {
		fmt.Fprintf(ui.Out, prefix)

		// for all columns except last, add cell value and padding
		for col := 0; col < columns-1; col++ {
			var addedPadding int
			if col+1 != columns {
				addedPadding = columnPadding[col] - runewidth.StringWidth(table[row][col])
			}
			fmt.Fprintf(ui.Out, "%s%s", table[row][col], strings.Repeat(" ", addedPadding))
		}

		// for last column, add each word individually. If the added word would make the column exceed terminal width, create a new line and add padding
		words := strings.Split(table[row][columns-1], " ")
		currentWidth := 0

		for _, word := range words {
			wordWidth := runewidth.StringWidth(word)
			if currentWidth == 0 {
				currentWidth = wordWidth
				fmt.Fprintf(ui.Out, "%s", word)
			} else if wordWidth+1+currentWidth > lastColumnWidth {
				fmt.Fprintf(ui.Out, "\n%s%s", strings.Repeat(" ", spilloverPadding), word)
				currentWidth = wordWidth
			} else {
				fmt.Fprintf(ui.Out, " %s", word)
				currentWidth += wordWidth + 1
			}
		}

		fmt.Fprintf(ui.Out, "\n")
	}
}

func sum(intSlice []int) int {
	sum := 0

	for _, i := range intSlice {
		sum += i
	}

	return sum
}

// DisplayText translates the template, substitutes in templateValues, and
// outputs the result to ui.Out. Only the first map in templateValues is used.
func (ui *UI) DisplayText(template string, templateValues ...map[string]interface{}) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Out, "%s\n", ui.TranslateText(template, templateValues...))
}

// DisplayHeader translates the header, bolds and adds the default color to the
// header, and outputs the result to ui.Out.
func (ui *UI) DisplayHeader(text string) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Out, "%s\n", ui.addFlavor(ui.TranslateText(text), defaultFgColor, true))
}

// DisplayTextWithFlavor translates the template, bolds and adds cyan color to
// templateValues, substitutes templateValues into the template, and outputs
// the result to ui.Out. Only the first map in templateValues is used.
func (ui *UI) DisplayTextWithFlavor(template string, templateValues ...map[string]interface{}) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	firstTemplateValues := getFirstSet(templateValues)
	for key, value := range firstTemplateValues {
		firstTemplateValues[key] = ui.addFlavor(fmt.Sprint(value), cyan, true)
	}
	fmt.Fprintf(ui.Out, "%s\n", ui.TranslateText(template, firstTemplateValues))
}

// DisplayWarning translates the warning, substitutes in templateValues, and
// outputs to ui.Err. Only the first map in templateValues is used.
func (ui *UI) DisplayWarning(template string, templateValues ...map[string]interface{}) {
	fmt.Fprintf(ui.Err, "%s\n", ui.TranslateText(template, templateValues...))
}

// DisplayWarnings translates the warnings and outputs to ui.Err.
func (ui *UI) DisplayWarnings(warnings []string) {
	for _, warning := range warnings {
		fmt.Fprintf(ui.Err, "%s\n", ui.TranslateText(warning))
	}
}

// DisplayError outputs the translated error message to ui.Err if the error
// satisfies TranslatableError, otherwise it outputs the original error message
// to ui.Err. It also outputs "FAILED" in bold red to ui.Out.
func (ui *UI) DisplayError(err error) {
	var errMsg string
	if translatableError, ok := err.(TranslatableError); ok {
		errMsg = translatableError.Translate(ui.translate)
	} else {
		errMsg = err.Error()
	}
	fmt.Fprintf(ui.Err, "%s\n", errMsg)

	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	fmt.Fprintf(ui.Out, "%s\n", ui.addFlavor(ui.TranslateText("FAILED"), red, true))
}

const LogTimestampFormat = "2006-01-02T15:04:05.00-0700"

// DisplayLogMessage formats and outputs a given log message.
func (ui *UI) DisplayLogMessage(message LogMessage, displayHeader bool) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	var header string
	if displayHeader {
		time := message.Timestamp().In(ui.TimezoneLocation).Format(LogTimestampFormat)

		header = fmt.Sprintf("%s [%s/%s] %s ",
			time,
			message.SourceType(),
			message.SourceInstance(),
			message.Type(),
		)
	}

	for _, line := range strings.Split(message.Message(), "\n") {
		logLine := fmt.Sprintf("%s%s", header, strings.TrimRight(line, "\r\n"))
		if message.Type() == "ERR" {
			logLine = ui.addFlavor(logLine, red, false)
		}
		fmt.Fprintf(ui.Out, "%s\n", logLine)
	}
}

// addFlavor adds the provided text color and bold style to the text.
func (ui *UI) addFlavor(text string, textColor color.Attribute, isBold bool) string {
	if len(text) == 0 {
		return text
	}

	colorPrinter := color.New(textColor)

	switch ui.colorEnabled {
	case configv3.ColorEnabled:
		colorPrinter.EnableColor()
	case configv3.ColorDisabled:
		colorPrinter.DisableColor()
	}

	if isBold {
		colorPrinter = colorPrinter.Add(color.Bold)
	}

	return colorPrinter.SprintFunc()(text)
}

// getFirstSet returns the first map if 1 or more maps are provided. Otherwise
// it returns the empty map.
func getFirstSet(list []map[string]interface{}) map[string]interface{} {
	if list == nil || len(list) == 0 {
		return map[string]interface{}{}
	}
	return list[0]
}
