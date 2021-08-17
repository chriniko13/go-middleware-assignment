package message

import "time"

// --- consuming ---

type AlarmStatusChangedMessage struct {
	AlarmId string
	UserId string
	Status string
	ChangedAt time.Time
}


type SendAlarmDigestMessage struct {
	UserId string
}




// --- producing ---

type AlarmDigestMessage struct {
	UserId string
	ActiveAlarms []ActiveAlarm
}

type ActiveAlarm struct {
	AlarmId string
	Status string
	LatestChangedAt time.Time
}

