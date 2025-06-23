package logger

import (
	"fmt"
	"os"
)

func Info(msg string) {
	fmt.Fprintf(os.Stdout, "[INFO] %s\n", msg)
}

func Error(msg string, err error) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s: %v\n", msg, err)
}

func Fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "[FATAL] %s: %v\n", msg, err)
	os.Exit(1)
}