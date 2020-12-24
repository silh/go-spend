package log

import (
	"log"
)

const (
	TraceLvl = iota
	DebugLvl
	InfoLvl
	WarnLvl
	ErrorLvl
)

var Level = InfoLvl

func Trace(msg string, args ...interface{}) {
	if Level <= TraceLvl {
		log.Printf(msg, args...)
	}
}

func Debug(msg string, args ...interface{}) {
	if Level <= DebugLvl {
		log.Printf(msg, args...)
	}
}

func Info(msg string, args ...interface{}) {
	if Level <= InfoLvl {
		log.Printf(msg, args...)
	}
}

func Warn(msg string, args ...interface{}) {
	if Level <= WarnLvl {
		log.Printf(msg, args...)
	}
}

func Error(msg string, args ...interface{}) {
	if Level <= ErrorLvl {
		log.Printf(msg, args...)
	}
}

func Fatal(msg string, args ...interface{}) {
	if Level <= ErrorLvl {
		log.Printf(msg, args...)
	}
}
