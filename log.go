package main

import (
	// "log"
	"os"
	"runtime/debug"

	// "strings"
	log "gopkg.in/inconshreveable/log15.v2"
)

var LogStackHandler = log.CallerStackHandler("%+v", log.StdoutHandler)
var LogFuncHandler = log.CallerFuncHandler(log.StdoutHandler)

func ErrHandler(err error) {
	if err != nil {
		switch os.Getenv("RUNMODE") {
		case "production":
			log.Error("Error occured", "error", err)
		case "testing":
			log.Root().SetHandler(LogFuncHandler)
			log.Error("Error occured", "error", err)
		case "development":
			fallthrough
		default:
			debug.PrintStack()
			log.Root().SetHandler(LogStackHandler)
			log.Error("Error occured", "error", err)
		}
	}
}
