package log

import (
	"fmt"
)

type ConsoleLogger struct {
	TimeZone   string
	TimeLayout string
	BeforeTime string
	AfterTime  string
}

// log a message to the console
func (cl *ConsoleLogger) LogMsg(message string) {
	time, timeError := GetCurrentTime(cl.TimeZone, cl.TimeLayout)
	if timeError != nil {
		panic(timeError)
	}
	rsMsg := fmt.Sprintf("%s%s%s%s", cl.BeforeTime, time, cl.AfterTime, message)
	fmt.Println(rsMsg)
}

// set the struct field BeforeTime
// BeforeTime will always be printed before the timestamp
func (cl *ConsoleLogger) SetBeforeTime(beforeTime string) {
	cl.BeforeTime = beforeTime
}

// set the struct field AfterTime
// AfterTime will always be printed after the timestamp
func (cl *ConsoleLogger) SetAfterTime(afterTime string) {
	cl.AfterTime = afterTime
}

// set the layout for the timestamp to be put out as
func (cl *ConsoleLogger) SetTimeLayout(layout string) {
	cl.TimeLayout = layout
}

// set the timezone for the logger. used for printing timestamps
func (cl *ConsoleLogger) SetTimeZone(timeZone string, timeLayout string) {
	cl.TimeZone = timeZone
}

// create and return a logger instance pointer
func NewConsoleLogger(timeZone string, timeLayout string) *ConsoleLogger {
	return &ConsoleLogger{
		TimeZone:   timeZone,
		TimeLayout: timeLayout,
	}
}
