package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"time"
)

type MockApi struct {
	client *http.Client
	mac    string
}

func InitAPI(iface string) (*MockApi, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var mac net.HardwareAddr
	for _, ifa := range ifaces {
		if ifa.Name == iface {
			mac = ifa.HardwareAddr
		}
	}
	if len(mac) == 0 {
		return nil, errors.New(fmt.Sprintf("Could not find MAC address for interface '%s'", iface))
	}

	return &MockApi{
		client: http.DefaultClient,
		mac:    mac.String(),
	}, nil
}

func (a *MockApi) CheckCard(uid []byte) bool {
	data, err := json.Marshal(CardEvent{
		CardID: hex.EncodeToString(uid),
		commonFields: commonFields{
			Mac:       a.mac,
			Timestamp: time.Now(),
		},
	})
	if err != nil {
		slog.Error("Failed to check card uid")
		return false
	}

	slog.Info("Check card", "data", data)

	time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
	return rand.Intn(2) < 1
}

func (a *MockApi) PostMove() {
	data, err := json.Marshal(MoveEvent{
		commonFields: commonFields{
			Mac:       a.mac,
			Timestamp: time.Now(),
		},
	})
	if err != nil {
		slog.Error("Failed to post movement event")
	}
	slog.Info("Move post", "data", data)
	time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
}
