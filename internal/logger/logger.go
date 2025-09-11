package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Logger struct {
	std *log.Logger
}

func New() *Logger {
	return &Logger{std: log.New(os.Stdout, "", 0)}
}

func (l *Logger) Info(msg string, keyvals ...interface{}) {
	fields := make([]string, 0, len(keyvals)/2+2)
	fields = append(fields, "ts="+time.Now().UTC().Format(time.RFC3339))
	fields = append(fields, "level=INFO")
	fields = append(fields, fmt.Sprintf("msg=\"%s\"", escapeQuotes(msg)))
	for i := 0; i+1 < len(keyvals); i += 2 {
		k, v := keyvals[i], keyvals[i+1]
		fields = append(fields, fmt.Sprintf("%v=%v", k, v))
	}
	l.std.Println(strings.Join(fields, " "))
}

func (l *Logger) Error(msg string, keyvals ...interface{}) {
	fields := make([]string, 0, len(keyvals)/2+2)
	fields = append(fields, "ts="+time.Now().UTC().Format(time.RFC3339))
	fields = append(fields, "level=ERROR")
	fields = append(fields, fmt.Sprintf("msg=\"%s\"", escapeQuotes(msg)))
	for i := 0; i+1 < len(keyvals); i += 2 {
		k, v := keyvals[i], keyvals[i+1]
		fields = append(fields, fmt.Sprintf("%v=%v", k, v))
	}
	l.std.Println(strings.Join(fields, " "))
}

func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}
