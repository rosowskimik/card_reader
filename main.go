package main

import (
	"io"
	"log"
	"log/slog"
	"log/syslog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rosowskimik/card_reader/app"
	"github.com/rosowskimik/card_reader/config"
)

func fatal(msg string, err error) {
	slog.Error(msg, slog.String("error", err.Error()))
	os.Exit(1)
}

func setupLog() error {
	writers := make([]io.Writer, 0, 2)
	writers = append(writers, os.Stdout)

	syslogger, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "")
	if err != nil {
		slog.Error("Failed to connect to system logger")
	} else {
		writers = append(writers, syslogger)
	}

	writer := io.MultiWriter(writers...)
	log.SetFlags(0)
	log.SetOutput(writer)
	return nil
}

func main() {
	setupLog()

	if err := config.InitConfig(); err != nil {
		fatal("Failed to parse config", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	slog.Info("Registered signal handler")

	app, err := app.Init()
	if err != nil {
		fatal("Error setting up app", err)
	}

	select {
	case err := <-app.Run():
		app.Stop()
		fatal("Stopping app due to fatal error", err)
		break
	case <-sigChan:
		slog.Info("Received stop signal - stopping app")
		app.Stop()
		break
	}
}
