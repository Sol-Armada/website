package service

import (
	"log/slog"

	"github.com/sol-armada/sol-bot/config"
)

type ConfigService struct {
	logger *slog.Logger
}

func NewConfigService(logger *slog.Logger) *ConfigService {
	if err := config.Setup(); err != nil {
		logger.Error("Failed to setup config", "error", err)
	}
	return &ConfigService{logger: logger}
}

func (s *ConfigService) GetAvailableAttendanceNames() ([]string, error) {
	return config.GetAttendanceNames()
}
