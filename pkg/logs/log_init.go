package logs

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func NewLog(logLevel string) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// timeZone := os.Getenv("APP_TIMEZONE")
	// location, err := time.LoadLocation(timeZone)
	// if err != nil {
	// 	location = time.UTC
	// }

	consoleWriter := zerolog.ConsoleWriter{
		Out: os.Stderr,
		FormatTimestamp: func(i interface{}) string {
			// Disable default timestamp printing — we’ll add it manually
			return ""
		},
		FormatLevel: func(i interface{}) string {
			// Disable built-in level printing
			return ""
		},
		FormatMessage: func(i interface{}) string {
			// No extra timestamp here — message contains formatted block
			return fmt.Sprintf("\n%s", i)
		},
		FormatFieldName: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("\n%-13s:", i))
		},
		FormatFieldValue: func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		},
	}

	jsonFile, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("Could not open JSON log file: %v", err))
	}

	multiWriter := zerolog.MultiLevelWriter(consoleWriter, jsonFile)
	Logger = zerolog.New(multiWriter).With().Logger()
}
