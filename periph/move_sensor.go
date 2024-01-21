package periph

import (
	"errors"
	"log/slog"
	"sync/atomic"

	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
)

const (
	offsetPin = rpi.GPIO26
	chip      = "gpiochip0"
)

type EventType int

const (
	EdgeRising EventType = iota
	EdgeFalling
)

func (g EventType) String() string {
	if g == EdgeRising {
		return "rising"
	} else {
		return "falling"
	}
}

type MovementSensor struct {
	line    *gpiod.Line
	started *atomic.Bool
	evtC    chan EventType
}

func InitMove(label string) (*MovementSensor, error) {
	// Buffer up to 2 events (edge rising/falling pair)
	evtC := make(chan EventType, 2)
	started := &atomic.Bool{}

	eventHandler := func(evt gpiod.LineEvent) {
		if !started.Load() {
			return
		}

		evtType := EdgeRising
		if evt.Type == gpiod.LineEventFallingEdge {
			evtType = EdgeFalling
		}

		select {
		case evtC <- evtType:
			break
		default:
			g := slog.Group("event",
				slog.Uint64("seqNo", uint64(evt.Seqno)),
				slog.Uint64("lineSeqNo", uint64(evt.LineSeqno)),
				slog.Int("offset", evt.Offset),
				slog.String("edge", evtType.String()),
				slog.Int64("timestamp", int64(evt.Timestamp)),
			)
			slog.Warn("Event channel overflow. Dropping event", g)
		}
	}
	line, err := gpiod.RequestLine(
		chip,
		offsetPin,
		gpiod.WithConsumer(label),
		gpiod.WithPullDown,
		gpiod.WithBothEdges,
		gpiod.WithEventHandler(eventHandler),
	)
	if err != nil {
		return nil, err
	}

	return &MovementSensor{
		line,
		started,
		evtC,
	}, nil
}

func (m *MovementSensor) Start() error {
	if !m.started.CompareAndSwap(false, true) {
		return errors.New("Sensor already started")
	}

	return nil
}

func (m *MovementSensor) WatchEvent() <-chan EventType {
	return m.evtC
}

func (m *MovementSensor) Pause() {
	m.started.Store(false)
}

func (m *MovementSensor) Resume() {
	m.started.Store(true)
}

func (m *MovementSensor) Close() error {
	m.started.Store(false)
	return m.line.Close()
}
