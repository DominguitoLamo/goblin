package main

import (
	"log"
	"fmt"
)

const (
	INFO = 1
	DEBUG    = 2
	WARNING  = 3
	ERROR = 4
)

const _LOGLEVEL = 2

func DebugLog(format string, s ...interface{}) {
	if (_LOGLEVEL > DEBUG) {
		return
	}
	log.Printf("DEBUG: %s\n", fmt.Sprintf(format, s...))
}

func InfoLog(format string, s ...interface{}) {
	if (_LOGLEVEL > INFO) {
		return
	}
	log.Printf("INFO: %s\n", fmt.Sprintf(format, s...))
}

func ErrorLog(format string, s ...interface{}) {
	if (_LOGLEVEL > ERROR) {
		return
	}
	log.Printf("ERROR: %s\n", fmt.Sprintf(format, s...))
}