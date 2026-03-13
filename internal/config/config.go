package config

import (
	"flag"
	"fmt"
)

const (
	DefaultHost = "0.0.0.0"
	DefaultPort = 8631
)

type Config struct {
	Host string
	Port int
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func Parse(args []string) (Config, error) {
	cfg := Config{
		Host: DefaultHost,
		Port: DefaultPort,
	}

	fs := flag.NewFlagSet("mdpreview", flag.ContinueOnError)
	fs.StringVar(&cfg.Host, "host", DefaultHost, "server host")
	fs.IntVar(&cfg.Port, "port", DefaultPort, "server port")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if cfg.Host == "" {
		return Config{}, fmt.Errorf("host cannot be empty")
	}

	if cfg.Port <= 0 || cfg.Port > 65535 {
		return Config{}, fmt.Errorf("port must be between 1 and 65535")
	}

	return cfg, nil
}
