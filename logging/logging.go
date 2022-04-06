package logging

import (
	"fmt"
	"io"
	"log"
	"os"

	te "github.com/muesli/termenv"
)

var color = te.ColorProfile().Color
var infoColor = "#d3d3d3"
var successColor = "#00ff00"
var errorColor = "#ff0000"
var warningColor = "#ff0000"

var f *os.File

func init() {
	f, err := os.OpenFile("lateralus.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
}

func WriteFile(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(f, msg)
}

// Infof will log messages to os.Stdout with INFO level
func Infof(format string, args ...interface{}) {
	printLog("info", format, args...)
}

// Errorf will log messages to os.Stdout with ERROR level
func Errorf(format string, args ...interface{}) {
	printLog("error", format, args...)
}

func Warningf(format string, args ...interface{}) {
	printLog("warning", format, args...)
}

// Fatalf will log messages to os.Stdout and will call os.Exit(1) afterwards
func Fatalf(format string, args ...interface{}) {
	printLog("fatal", format, args...)
	os.Exit(1)
}

func Successf(format string, args ...interface{}) {
	printLog("success", format, args...)
}

func printLog(logType, format string, args ...interface{}) {
	var c te.Color
	var level string

	switch logType {
	case "info":
		c = color(infoColor)
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
	case "success":
		c = color(successColor)
		level = "[SUCCESS]"
	}

	levelMsg := te.String(level).Bold().Foreground(c).String()
	msg := te.String(fmt.Sprintf(format, args...)).Bold().Foreground(c).String()
	log.Println(levelMsg, msg)
}
