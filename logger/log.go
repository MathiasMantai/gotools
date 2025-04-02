package logger

import (
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"strings"
)

type Logger struct {
	Options LoggerOptions
}

type LoggerOptions struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	Filename string `json:"filename" yaml:"filename"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize" yaml:"maxsize"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `json:"maxage" yaml:"maxage"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"maxbackups" yaml:"maxbackups"`

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool `json:"localtime" yaml:"localtime"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `json:"compress" yaml:"compress"`

	//LogToConsole determines whether console messages will be printed out
	//default is true
	LogToConsole bool `json:"log_to_console" yaml:"log_to_console"`

	//LogToFile determines whether messages will be printed to a log file
	//default is false
	LogToFile bool `json:"log_to_file" yaml:"log_to_file"`

	//LogWithTime determines whether timestamps will be printed before messages
	LogWithTimestamp bool `json:"log_with_timestamp" yaml:"log_with_timestamp"`

	//determines if console messages will be printed colored
	//default is true
	LogWithColor bool

	//determines the color for the console strings of a success message
	//default is white
	MessageColor string

	//determines the color for the console strings of a success message
	//default is red
	ErrorColor string

	//default is yellow
	WarningColor string

	//determines the color for the console strings of a success message
	//default is green
	SuccessColor string
}

func Create(options *LoggerOptions) Logger {

	if strings.TrimSpace(options.MessageColor) == "" {
		options.MessageColor = "white"
	}

	if strings.TrimSpace(options.SuccessColor) == "" {
		options.SuccessColor = "green"
	}

	if strings.TrimSpace(options.ErrorColor) == "" {
		options.ErrorColor = "red"
	}

	if strings.TrimSpace(options.WarningColor) == "" {
		options.WarningColor = "yellow"
	}

	// log.SetOutput(&lumberjack.Logger{
	// 	Filename: options.Filename,
	// 	MaxSize: options.MaxSize,
	// 	MaxAge: options.MaxAge,
	// 	MaxBackups: options.MaxBackups,
	// 	LocalTime: options.LocalTime,
	// })




	if options.LogToFile {
		lLogger := &lumberjack.Logger{}

		if strings.TrimSpace(options.Filename) != "" {
			lLogger.Filename = options.Filename
		}

		if options.MaxSize != 0 {
			lLogger.MaxSize = options.MaxSize
		}

		if options.MaxAge != 0 {
			lLogger.MaxAge = options.MaxAge
		}

		if options.MaxBackups != 0 {
			lLogger.MaxBackups = options.MaxBackups
		}

		lLogger.LocalTime = options.LocalTime
		lLogger.Compress = options.Compress

		log.SetOutput(lLogger)
	}

	return Logger{
		Options: *options,
	}
}

func (l *Logger) PrintMessage(messageType string, message string) {

	var color string
	switch messageType {
	case "error":
		color = l.Options.ErrorColor
	case "warning":
		color = l.Options.WarningColor
	case "success":
		color = l.Options.SuccessColor
	default:
		color = l.Options.MessageColor
	}

	//print to console
	if l.Options.LogWithTimestamp {
		if l.Options.LogWithColor {
			cli.PrintWithTimeAndColor(message, color, true)
		} else {
			cli.PrintWithTime(message, true)
		}
	} else {
		if l.Options.LogWithColor {
			cli.PrintColor(message, color, true)
		} else {
			fmt.Println(message)
		}
	}

	//print to file
	if l.Options.LogToFile {
		log.Println(message)
	}
}

// log without any kind of prev
func (l *Logger) LogMessage(message string) {
	l.PrintMessage("message", message)
}

func (l *Logger) LogError(message string) {
	l.PrintMessage("error", message)
}

func (l *Logger) LogSuccess(message string) {
	l.PrintMessage("success", message)
}

func (l *Logger) LogWarning(message string) {
	l.PrintMessage("warning", message)
}
