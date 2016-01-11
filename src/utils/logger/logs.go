/*
Package logs defines logs with default flags.
*/
package logger

import (
	"log"
	"os"
)

var logger Logger = log.New(os.Stderr, "", log.LstdFlags)

type Logger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}

func SetLogger(l Logger) {
	logger = l
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
}

func Print(args ...interface{}) {
	logger.Print(args...)
}

func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

func Println(args ...interface{}) {
	logger.Println(args...)
}

func Info(args ...interface{}) {
	logger.Printf("[INFO] %+v\n", args...)
}

func Error(args ...interface{}) {
	logger.Fatalf("[ERROR] %+v\n", args...)
}

func Infof(foramt string, args ...interface{}) {
	logger.Printf("[INFO] "+foramt, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Fatalf("[ERROR] "+format, args...)
}
