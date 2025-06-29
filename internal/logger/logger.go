package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Info(source, message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(os.Stdout, "%s [%s] %s\n", timestamp(), source, msg)
}

func Error(source, message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(os.Stderr, "%s [%s] ❌ %s\n", timestamp(), source, msg)
}

func Debug(source, message string, args ...any) {
	debug := strings.ToLower(os.Getenv("DEBUG")) == "true"
	if debug {
		msg := fmt.Sprintf(message, args...)
		fmt.Fprintf(os.Stdout, "%s [%s] 🔍 %s\n", timestamp(), source, msg)
	}
}
