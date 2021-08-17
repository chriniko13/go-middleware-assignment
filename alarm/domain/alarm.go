package domain

import (
	"time"
)

type Alarm struct {
	Id     AlarmId
	UserId UserId

	Status AlarmStatus

	CreatedAt time.Time
	UpdatedAt time.Time
}

type AlarmId string
type UserId string

type AlarmStatus string

const (
	CLEARED  AlarmStatus = "CLEARED"
	WARNING  AlarmStatus = "WARNING"
	CRITICAL AlarmStatus = "CRITICAL"

	UNKNOWN AlarmStatus = "UNKNOWN"
)

func ExtractAlarmStatus(status string) AlarmStatus {
	switch status {
	case "CRITICAL":
		return CRITICAL
	case "WARNING":
		return WARNING
	case "CLEARED":
		return CLEARED

	default:
		return UNKNOWN
	}
}

func AlarmStatusAsString(status AlarmStatus) string {
	switch status {
	case CRITICAL:
		return "CRITICAL"
	case WARNING:
		return "WARNING"
	case CLEARED:
		return "CLEARED"

	default:
		return "UNKNOWN"
	}
}

func IsActiveAlarm(alarm *Alarm) bool {
	return alarm.Status == CRITICAL || alarm.Status == WARNING
}
