package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

const CONFIG_PATH = "/etc/xdg/card_reader/config.yml"

type AppConfig struct {
	InitLocked bool `envconfig:"HAL_INIT_LOCKED"`

	Network struct {
		Interface string `envconfig:"HAL_NET_INTERFACE"`
		Hostname  string `envconfig:"HAL_API_HOST"`
		Id        string `envconfig:"HAL_SYSTEM_ID"`
	}

	Periph struct {
		Leds struct {
			RedLedName   string `yaml:"red_name" envconfig:"HAL_RED_LED_NAME"`
			GreenLedName string `yaml:"green_name" envconfig:"HAL_GREEN_LED_NAME"`
		}
		Reader struct {
			AntennaStrength int8          `yaml:"strength" envconfig:"HAL_ANTENNA_STRENGTH"`
			ReaderTimeout   time.Duration `yaml:"timeout" envconfig:"HAL_READER_TIMEOUT"`
		}

		Movement struct {
			MoveTimeout time.Duration `yaml:"timeout" envconfig:"HAL_MOVEMENT_TIMEOUT"`
		}
	}
}

var Config AppConfig

func InitConfig() error {
	fh, err := os.Open(CONFIG_PATH)
	if err != nil {
		slog.Error("Failed to open config file")
		return err
	}
	defer fh.Close()

	decoder := yaml.NewDecoder(fh)
	if err := decoder.Decode(&Config); err != nil {
		slog.Error("Failed to parse config file")
		return err
	}

	if envconfig.Process("", &Config) != nil {
		slog.Warn("Failed to parse environment")
	}

	slog.Info("App starting", slog.Any("config", Config))

	return nil
}
