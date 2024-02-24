package main

import (
	"fmt"
	"time"
)

type logMsg struct {
	msg   string
	level int
}

var logFormat = "[%s] [%s] %s\n"

func levelToText(level int) string {
	switch level {
	case 1:
		return "INFO"
	case 2:
		return "ERROR"
	default:
		return "INFO"

	}
}

func logger(logs <-chan logMsg) {
	for msg := range logs {
		if verbose {
			fmt.Printf(logFormat, levelToText(msg.level), time.Now().Format("2006-01-02 15:04:05"), msg.msg)
		}
	}
}

func forceLogger(l logMsg) {
	fmt.Printf(logFormat, levelToText(l.level), time.Now().Format("2006-01-02 15:04:05"), l.msg)
}
