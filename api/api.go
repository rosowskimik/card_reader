package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/rosowskimik/card_reader/config"
)

type Api struct {
	client *http.Client
	id     string
	mac    string
	base   *url.URL
}

func InitAPI(cfg config.AppConfig) (*Api, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var mac net.HardwareAddr
	for _, ifa := range ifaces {
		if ifa.Name == cfg.Network.Interface {
			mac = ifa.HardwareAddr
		}
	}
	if len(mac) == 0 {
		return nil, errors.New(fmt.Sprintf("Could not find MAC address for interface '%s'", cfg.Network.Interface))
	}

	base, err := url.Parse(cfg.Network.Hostname)
	if err != nil {
		return nil, err
	}

	return &Api{
		client: &http.Client{},
		id:     cfg.Network.Id,
		mac:    mac.String(),
		base:   base,
	}, nil
}

func (a *Api) CheckCard(uid []byte) bool {
	data, err := json.Marshal(CardEvent{
		CardID: hex.EncodeToString(uid),
		commonFields: commonFields{
			Id: a.id,
		},
	})

	if err != nil {
		slog.Error("Json parsing failed")
		return false
	}

	slog.Debug("Check card", "data", string(data))

	resp, err := a.request(http.MethodPost, "/api/system/authorize", data)
	if err != nil {
		slog.Error("Failed to send card check request", "error", err)
		return false
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 202:
		return true
	case 401:
		return false
	default:
		slog.Error("Unknown response status code", "code", resp.Status)
		return false
	}
}

func (a *Api) PostMove() {
	data, err := json.Marshal(MoveEvent{
		commonFields: commonFields{
			Id: a.id,
		},
		Timestamp: time.Now().Unix(),
	})

	if err != nil {
		slog.Error("Json parsing failed")
		return
	}

	slog.Debug("Move post", "data", string(data))

	resp, err := a.request(http.MethodPost, "/api/event", data)
	if err != nil {
		slog.Error("Failed to send move detect event", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		slog.Error("Unknown response status code", "code", resp.Status)
	}
}

func (a *Api) request(method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, a.base.JoinPath(url).String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")

	return a.client.Do(req)
}
