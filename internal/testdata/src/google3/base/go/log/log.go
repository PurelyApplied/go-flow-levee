// Package log fakes google3/base/go/log.
package log

import (
	"fmt"
)

// Info fakes log.Info
func Info(args ...interface{}) {
	fmt.Print(args...)
}

// InfoDepth fakes log.InfoDepth
func InfoDepth(depth int, args ...interface{}) {
	fmt.Print(append(args, depth)...)
}

// InfoDepthf fakes log.InfoDepthf
func InfoDepthf(depth int, format string, args ...interface{}) {
	fmt.Printf(format, append(args, depth)...)
}

// Infoln fakes log.Infoln
func Infoln(args ...interface{}) {
	fmt.Println(args...)
}

// Infof fakes log.Infof
func Infof(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Warning fakes log.Warning
func Warning(args ...interface{}) {
	fmt.Print(args...)
}

// WarningDepth fakes log.WarningDepth
func WarningDepth(depth int, args ...interface{}) {
	fmt.Print(append(args, depth)...)
}

// WarningDepthf fakes log.WarningDepthf
func WarningDepthf(depth int, format string, args ...interface{}) {
	fmt.Printf(format, append(args, depth)...)
}

// Warningln fakes log.Warningln
func Warningln(args ...interface{}) {
	fmt.Println(args...)
}

// Warningf fakes log.Warningf
func Warningf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Error fakes log.Error
func Error(args ...interface{}) {
	fmt.Print(args...)
}

// ErrorDepth fakes log.ErrorDepth
func ErrorDepth(depth int, args ...interface{}) {
	fmt.Print(append(args, depth)...)
}

// ErrorDepthf fakes log.ErrorDepthf
func ErrorDepthf(depth int, format string, args ...interface{}) {
	fmt.Printf(format, append(args, depth)...)
}

// Errorln fakes log.Errorln
func Errorln(args ...interface{}) {
	fmt.Println(args...)
}

// Errorf fakes log.Errorf
func Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Fatal fakes log.Fatal
func Fatal(args ...interface{}) {
	fmt.Print(args...)
}

// FatalDepth fakes log.FatalDepth
func FatalDepth(depth int, args ...interface{}) {
	fmt.Print(append(args, depth)...)
}

// FatalDepthf fakes log.FatalDepthf
func FatalDepthf(depth int, format string, args ...interface{}) {
	fmt.Printf(format, append(args, depth)...)
}

// Fatalln fakes log.Fatalln
func Fatalln(args ...interface{}) {
	fmt.Println(args...)
}

// Fatalf fakes log.Fatalf
func Fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Exit fakes log.Exit
func Exit(args ...interface{}) {
	fmt.Print(args...)
}

// ExitDepth fakes log.ExitDepth
func ExitDepth(depth int, args ...interface{}) {
	fmt.Print(append(args, depth)...)
}

// ExitDepthf fakes log.ExitDepthf
func ExitDepthf(depth int, format string, args ...interface{}) {
	fmt.Printf(format, append(args, depth)...)
}

// Exitln fakes log.Exitln
func Exitln(args ...interface{}) {
	fmt.Println(args...)
}

// Exitf fakes log.Exitf
func Exitf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
