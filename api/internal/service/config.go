package service

import (
	"github.com/sol-armada/sol-bot/config"
)

func SetupConfigService() error {
	if err := config.Setup(); err != nil {
		return err
	}
	return nil
}

func GetAvailableAttendanceNames() ([]string, error) {
	return config.GetAttendanceNames()
}

func CreateAttendanceName(name string) error {
	return config.NewAttendanceName(name)
}

func DeleteAttendanceName(name string) error {
	return config.RemoveAttendanceName(name)
}
