package util

import (
	"fmt"
	"log"
)

type logLevel uint8

const (
	NONE logLevel = iota
	ERROR
	INFO
	WARN
	DEBUG
)

type Logger interface {
	Error(s string, a ...any)
	Info(s string, a ...any)
	Warn(s string, a ...any)
	Debug(s string, a ...any)
}

type logger struct {
	*log.Logger
	namespace string
	logLevel  logLevel
}

func NewLogger(namespace, strLogLevel string) *logger {
	levelAssigner := func(s string) logLevel {
		switch s {
		case "NONE":
			return NONE
		case "ERROR":
			return ERROR
		case "INFO":
			return INFO
		case "WARN":
			return WARN
		case "DEBUG":
			return DEBUG
		default:
			return WARN
		}
	}

	return &logger{log.Default(), namespace, levelAssigner(strLogLevel)}
}

func (l *logger) Error(s string, a ...any) {
	if l.logLevel < ERROR {
		return
	}
	l.Printf("ERROR: %s - %s\n", l.namespace, fmt.Sprintf(s, a...))
}

func (l *logger) Info(s string, a ...any) {
	if l.logLevel < INFO {
		return
	}
	l.Printf("INFO: %s - %s\n", l.namespace, fmt.Sprintf(s, a...))
}

func (l *logger) Warn(s string, a ...any) {
	if l.logLevel < WARN {
		return
	}
	l.Printf("WARN: %s - %s\n", l.namespace, fmt.Sprintf(s, a...))
}

func (l *logger) Debug(s string, a ...any) {
	if l.logLevel < DEBUG {
		return
	}
	l.Printf("DEBUG: %s - %s\n", l.namespace, fmt.Sprintf(s, a...))
}
