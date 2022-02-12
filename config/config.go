package config

import (
	"fmt"
	"net"
)

type Config struct {
	InvocationSemantic
	Host string
	Port string
}

func (cfg *Config) Validate() error {
	if err := cfg.InvocationSemantic.Validate(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	_, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return err
	}

	return nil
}
