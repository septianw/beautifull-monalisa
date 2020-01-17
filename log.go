package main

import (
	"os"
	"runtime/debug"

	log "gopkg.in/inconshreveable/log15.v2"
)

var LogStackHandler = log.CallerStackHandler("%+v", log.StdoutHandler)
var LogFuncHandler = log.CallerFuncHandler(log.StdoutHandler)

// Every error shall dump here.
func ErrHandler(err error) {
	if err != nil {
		switch os.Getenv("RUNMODE") {
		case "production":
			L.Error("Error occured", "error", err)
		case "testing":
			log.Root().SetHandler(LogFuncHandler)
			L.Error("Error occured", "error", err)
		case "development":
			fallthrough
		default:
			debug.PrintStack()
			log.Root().SetHandler(LogStackHandler)
			L.Error("Error occured", "error", err)
		}
	}
}
