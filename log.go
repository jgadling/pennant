package main

import (
	"github.com/op/go-logging"
	"os"
)

var logger = logging.MustGetLogger("pennant")

// Configure the logger module.
func setLogLevel(level logging.Level) {
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{color:reset} %{message}`,
	)
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	leveledBackend := logging.AddModuleLevel(backendFormatter)
	leveledBackend.SetLevel(level, "")
	logging.SetBackend(leveledBackend)
}

// Default is INFO and above
func initLogger() {
	setLogLevel(logging.INFO)
}

// Bump to debug for more info
func enableDebugLogs() {
	setLogLevel(logging.DEBUG)
}
