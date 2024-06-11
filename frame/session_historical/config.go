package session_historical

import (
	"github.com/LimeChain/gosemble/frame/session"
)

type Config struct {
	SessionModule session.Module
}

func NewConfig(sessionModule session.Module) *Config {
	return &Config{
		SessionModule: sessionModule,
	}
}
