package logging

import (
	"fmt"
	"os"

	te "github.com/muesli/termenv"
)

var color = te.ColorProfile().Color
var successColor = "#00ff00"
var errorColor = "#ff0000"
var warningColor = "#ff0000"

// Infof will log messages to os.Stdout with INFO level
func Infof(format string, args ...interface{}) {
	printLog("info", format, args...)
}

// Errorf will log messages to os.Stdout with ERROR level
func Errorf(format string, args ...interface{}) {
	printLog("error", format, args...)
}

// Fatalf will log messages to os.Stdout and will call os.Exit(1) afterwards
func Fatalf(format string, args ...interface{}) {
	printLog("fatal", format, args...)
	os.Exit(1)
}

func printLog(logType, format string, args ...interface{}) {
	var c te.Color
	var level string

	switch logType {
	case "info":
		c = color(successColor)
		level = "[INFO]"
	case "warning":
		c = color(warningColor)
		level = "[WARNING]"
	case "error":
		c = color(errorColor)
		level = "[ERROR]"
	case "fatal":
		c = color(errorColor)
		level = "[FATAL]"
	}

	levelMsg := te.String(level).Bold().Foreground(c).String()
	msg := te.String(fmt.Sprintf(format, args...)).Bold().Foreground(c).String()
	fmt.Println(levelMsg, msg)
}
