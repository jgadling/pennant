package main

import (
	"github.com/op/go-logging"
	"os"
)

var logger = logging.MustGetLogger("pennant")

func initLogger() {
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	stderr := logging.NewLogBackend(os.Stdout, "", 0)
	stderrFormatter := logging.NewBackendFormatter(stderr, format)
	logging.SetBackend(stderrFormatter)
}
