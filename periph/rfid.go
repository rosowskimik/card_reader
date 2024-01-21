package periph

import (
	"log/slog"

	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/devices/v3/mfrc522"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
)

const (
	spiPort = "SPI0.0"
)

var (
	resetPin = rpi.P1_18
	irqPin   = rpi.P1_22
)

type RFIDController struct {
	port   spi.PortCloser
	dev    *mfrc522.Dev
	eventC chan []byte
	doneC  chan interface{}
}

func InitRFID(strength int) (*RFIDController, error) {
	if _, err := host.Init(); err != nil {
		return nil, err
	}

	port, err := spireg.Open(spiPort)
	if err != nil {
		return nil, err
	}

	dev, err := mfrc522.NewSPI(port, resetPin, irqPin)
	if err != nil {
		port.Close()
		return nil, err
	}

	gain := max(0, min(strength, 7))
	if err := dev.SetAntennaGain(gain); err != nil {
		dev.Halt()
		port.Close()
		return nil, err
	}
	slog.Debug("Antenna power set", slog.Int("power", gain))

	eventC := make(chan []byte)
	doneC := make(chan interface{})

	return &RFIDController{
		port,
		dev,
		eventC,
		doneC,
	}, nil
}

func (r *RFIDController) RecvUID() <-chan []byte {
	go func() {
		slog.Debug("Started receiver thread", slog.String("device", r.dev.String()))

		dataC := make(chan []byte)
		closeC := make(chan interface{})
		var data []byte
		go func() {
			for {
				data, err := r.dev.ReadUID(-1)
				if err != nil {
					select {
					case <-closeC:
						return
					default:
						continue
					}
				}
				dataC <- data
				return
			}
		}()

		select {
		case data = <-dataC:
			break
		case <-r.doneC:
			closeC <- nil
			close(r.eventC)
			return
		}

		select {
		case r.eventC <- data:
			break
		case <-r.doneC:
			closeC <- nil
			close(r.eventC)
			return
		}
	}()

	return r.eventC
}

func (r *RFIDController) Close() error {
	slog.Debug("Stopping RFID receiver")
	r.doneC <- nil
	if err := r.dev.LowLevel.SetAntenna(false); err != nil {
		return err
	}
	if err := r.dev.Halt(); err != nil {
		return err
	}
	return r.port.Close()
}
