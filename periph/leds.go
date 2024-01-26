package periph

import (
	"time"

	"github.com/rosowskimik/sled"
	"github.com/rosowskimik/sled/trigger"
)

type StatusLeds struct {
	redLed   *sled.LED
	greenLed *sled.LED
}

func InitLeds(redName, greenName string) (*StatusLeds, error) {
	redLed, err := sled.New(redName)
	if err != nil {
		return nil, err
	}
	if err := redLed.SetBrightness(0); err != nil {
		redLed.Close()
		return nil, err
	}

	greenLed, err := sled.New(greenName)
	if err != nil {
		redLed.Close()
		return nil, err
	}
	if err := greenLed.SetBrightness(0); err != nil {
		greenLed.Close()
		redLed.Close()
		return nil, err
	}

	return &StatusLeds{
		redLed,
		greenLed,
	}, nil
}

func (s *StatusLeds) LockedMode() error {
	return s.setBrightness(1, 0)
}

func (s *StatusLeds) CheckMode() error {
	return s.setBrightness(0, 0)
}

func (s *StatusLeds) UnlockedMode() error {
	return s.setBrightness(0, 1)
}

func (s *StatusLeds) ErrorMode() error {
	if err := s.greenLed.SetBrightness(0); err != nil {
		return err
	}
	return s.redLed.SetTrigger(trigger.NewTimer(170*time.Millisecond, 120*time.Millisecond))
}

func (s *StatusLeds) Close() error {
	if err := s.greenLed.Close(); err != nil {
		return err
	}
	return s.redLed.Close()
}

func (s *StatusLeds) setBrightness(red, green uint) error {
	if err := s.greenLed.SetBrightness(green); err != nil {
		return err
	}
	return s.redLed.SetBrightness(red)
}
