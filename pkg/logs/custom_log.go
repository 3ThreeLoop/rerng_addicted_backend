package logs

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type CustomLog struct {
	MessageID string
	Reason    string
	Function  string
	File      string
	Line      int
}

func (e *CustomLog) LogToString() string {
	return fmt.Sprintf(
		"[MessageID: %s] %s | %s (%s:%d)",
		e.MessageID, e.Reason, e.Function, filepath.Base(e.File), e.Line,
	)
}

func NewCustomLog(messageID, reason string, levelOpt ...string) *CustomLog {
	pc, file, line, ok := runtime.Caller(1)
	function := "unknown"
	if ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			function = fn.Name()
		}
	}

	entry := &CustomLog{
		MessageID: messageID,
		Reason:    reason,
		Function:  function,
		File:      file,
		Line:      line,
	}

	levelStr := "info"
	if len(levelOpt) > 0 && levelOpt[0] != "" {
		levelStr = strings.ToLower(levelOpt[0])
	}

	level, icon := resolveLogLevel(levelStr)
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	Logger.WithLevel(level).
		Str("🆔 MESSAGEID", entry.MessageID).
		Str("📁 FILE", filepath.Base(entry.File)).
		Str("📝 REASON", entry.Reason).
		Int("🔢 LINE", entry.Line).
		Str("🔧 FUNCTION", entry.Function).
		Msg(fmt.Sprintf(
			"%s %s (%s)\n──────────────────────────────────────────────────────────────",
			icon, strings.ToUpper(levelStr), currentTime,
		))

	return entry
}

func resolveLogLevel(level string) (zerolog.Level, string) {
	switch level {
	case "fatal":
		return zerolog.FatalLevel, "☠️"
	case "error":
		return zerolog.ErrorLevel, "🛑"
	case "warn", "warning":
		return zerolog.WarnLevel, "⚠️"
	case "debug":
		return zerolog.DebugLevel, "🐞"
	case "trace":
		return zerolog.TraceLevel, "🔍"
	default:
		return zerolog.InfoLevel, "ℹ️"
	}
}
