package main

import (
	"log"
	"os"
)

var debugLogger *log.Logger = NewDebugLogger()

func NewDebugLogger() *log.Logger {
	prefix := "[DEBUG]"
	path := "debug.log"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	l := log.New(f, prefix, log.LstdFlags)
	return l
}

func Debug(args ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		debugLogger.Println(args...)
	}
}
