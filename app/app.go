package app

import (
	"github.com/rosowskimik/card_reader/api"
	"github.com/rosowskimik/card_reader/config"
	"github.com/rosowskimik/card_reader/periph"

	"log/slog"
	"time"
)

type App struct {
	api    *api.MockApi
	leds   *periph.StatusLeds
	sensor *periph.MovementSensor
	reader *periph.RFIDController
	timer  *time.Timer
	locked bool
}

func Init() (*App, error) {
	api, err := api.InitAPI(config.AppConfig.Network.Interface)
	if err != nil {
		return nil, err
	}

	slog.Debug("Setting up status leds")
	leds, err := periph.InitLeds(config.AppConfig.Periph.Leds.RedLedName, config.AppConfig.Periph.Leds.GreenLedName)
	if err != nil {
		slog.Error("Failed to initialize leds")
		return nil, err
	}

	slog.Debug("Setting up movement sensor")
	sensor, err := periph.InitMove("move sensor")
	if err != nil {
		slog.Error("Failed to initialize movement sensor")
		leds.Close()
		return nil, err
	}

	slog.Debug("Setting up card reader")
	reader, err := periph.InitRFID(int(config.AppConfig.Periph.Reader.AntennaStrength))
	if err != nil {
		slog.Error("Failed to initialize card reader")
		sensor.Close()
		leds.Close()
		return nil, err
	}

	return &App{
		api:    api,
		leds:   leds,
		sensor: sensor,
		reader: reader,
		locked: config.AppConfig.InitLocked,
	}, nil
}

func (a *App) Run() <-chan error {
	c := make(chan error)

	go a.moveLoop(config.AppConfig.Periph.Movement.MoveTimeout, c)
	go a.cardLoop(config.AppConfig.Periph.Reader.ReaderTimeout, c)

	return c
}

func (a *App) Stop() {
	a.sensor.Close()
	a.reader.Close()
	a.leds.Close()
}

func (a *App) moveLoop(debounceTime time.Duration, e chan<- error) {
	if err := a.sensor.Start(); err != nil {
		slog.Error("Failed to start movement detection")
		e <- err
		return
	}
	slog.Info("Starting movement detection")

	var timer *time.Timer
	for ev := range a.sensor.WatchEvent() {
		if ev == periph.EdgeRising {
			if timer == nil || !timer.Stop() {
				a.api.PostMove()
				slog.Info("Movement detected")
				timer = nil
			}
		} else {
			timer = time.AfterFunc(15*time.Second, func() {
				timer = nil
			})
		}
	}
}

func (a *App) cardLoop(debounceTime time.Duration, e chan<- error) {
	ledErr := func(c chan<- error, e error) {
		slog.Error("Failed to update status leds")
		c <- e
	}

	slog.Info("Starting lock status leds")
	if err := a.displayLock(); err != nil {
		ledErr(e, err)
		return
	}

	slog.Info("Starting card reader")
	var uid []byte
	for {
		uid = <-a.reader.RecvUID()
		if err := a.leds.CheckMode(); err != nil {
			ledErr(e, err)
			return
		}

		if a.api.CheckCard(uid) {
			if a.locked = !a.locked; a.locked {
				slog.Info("Door locked")
			} else {
				slog.Info("Door unlocked")
			}
			if err := a.displayLock(); err != nil {
				ledErr(e, err)
				return
			}
		} else {
			slog.Warn("Lock attempt with unaithorized card")
			if err := a.displayError(debounceTime); err != nil {
				ledErr(e, err)
				return
			}
		}
	}
}

func (a *App) displayLock() error {
	if a.locked {
		return a.leds.LockedMode()
	}
	return a.leds.UnlockedMode()
}

func (a *App) displayError(dur time.Duration) error {
	if err := a.leds.ErrorMode(); err != nil {
		return err
	}
	<-time.After(dur)
	return a.displayLock()
}
