package main

import (
	"log/slog"
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

func main() {
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
