package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

const CONFIG_PATH = "/etc/xdg/card_reader/config.yml"

var AppConfig struct {
	Network struct {
		Interface string `envconfig:"HAL_NET_INTERFACE"`
		Hostname  string `envconfig:"HAL_API_HOST"`
	}

	Periph struct {
		Leds struct {
			RedLedName   string `yaml:"red_name" envconfig:"HAL_RED_LED_NAME"`
			GreenLedName string `yaml:"green_name" envconfig:"HAL_GREEN_LED_NAME"`
		}
		Reader struct {
			AntennaStrength int8          `yaml:"strength" envconfig:"HAL_ANTENNA_STRENGTH" default:"5"`
			ReaderTimeout   time.Duration `yaml:"timeout" envconfig:"HAL_READER_TIMEOUT" default:"7s"`
		}

		Movement struct {
			MoveTimeout time.Duration `yaml:"timeout" envconfig:"HAL_MOVEMENT_TIMEOUT" default:"20s"`
		}
	}
}

func InitConfig() error {
	fh, err := os.Open(CONFIG_PATH)
	if err != nil {
		slog.Error("Failed to open config file")
		return err
	}
	defer fh.Close()

	decoder := yaml.NewDecoder(fh)
	if err := decoder.Decode(&AppConfig); err != nil {
		slog.Error("Failed to parse config file")
		return err
	}

	if envconfig.Process("", &AppConfig) != nil {
		slog.Warn("Failed to parse environment")
	}

	slog.Info("Got app configuration", slog.Any("config", AppConfig))

	return nil
}
