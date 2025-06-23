package logger

import (
	"fmt"
	"os"
	"time"
)

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Info(msg string) {
	fmt.Fprintf(os.Stdout, "%s [INFO] %s\n", timestamp(), msg)
}

func Error(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s [ERROR] %s: %v\n", timestamp(), msg, err)
}

func Fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s [FATAL] %s: %v\n", timestamp(), msg, err)
	os.Exit(1)
}