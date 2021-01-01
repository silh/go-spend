// Simple logger wrapper that prepends log level
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
		log.Printf("[TRACE] "+msg, args...)
	}
}

func Debug(msg string, args ...interface{}) {
	if Level <= DebugLvl {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func Info(msg string, args ...interface{}) {
	if Level <= InfoLvl {
		log.Printf("[INFO] "+msg, args...)
	}
}

func Warn(msg string, args ...interface{}) {
	if Level <= WarnLvl {
		log.Printf("[WARN] "+msg, args...)
	}
}

func Error(msg string, args ...interface{}) {
	if Level <= ErrorLvl {
		log.Printf("[ERROR] "+msg, args...)
	}
}

func Fatal(msg string, args ...interface{}) {
	if Level <= ErrorLvl {
		log.Fatalf("[FATAL] "+msg, args...)
	}
}

func FatalErr(err error) {
	Fatal(err.Error())
}
