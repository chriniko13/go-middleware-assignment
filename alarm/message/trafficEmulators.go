package message

import (
	"alarm/domain"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"time"
)

/*
	The purpose of this file is to provide functions that will produce traffic to the following topics:

	* AlarmStatusChanged
	* SendAlarmDigest

	So we can use these functions during our end-2-end/integration tests, but also for fast manual testing during
	prototyping of the challenge.

	It is advised all methods start with emulate prefix.

*/

func EmulateTraffic(shots int, serverConnection *nats.Conn, alarmStatusChangedListeners int, sendAlarmDigestListeners int, bursts int, waitTimeInMs int64) {

	currentShot := 0

	for {

		if shots < 0 {
			// we run forever
		} else {
			if currentShot >= shots {
				break
			}
			currentShot++
		}

		for i:=0; i<bursts; i++ {
			go func() {

				alarmId, _ := uuid.NewUUID()
				userId, _ := uuid.NewUUID()

				_, _ = EmulateAlarmStatusChangedMessage(serverConnection, &domain.Alarm{
					Id:        domain.AlarmId(alarmId.String()),
					UserId:    domain.UserId(userId.String()),
					Status:    domain.CRITICAL,
					CreatedAt: time.Now(),
				}, time.Now(), alarmStatusChangedListeners)

				_, _ = EmulateSendAlarmDigestMessage(serverConnection, userId.String(), sendAlarmDigestListeners)

			}()
		}

		time.Sleep(time.Millisecond * time.Duration(waitTimeInMs))
	}
}


func EmulateAlarmStatusChangedMessage(conn *nats.Conn, alarm *domain.Alarm, changedAt time.Time, totalListeners int) ([]byte, error) {

	dataToSerialize := map[string]interface{}{}

	dataToSerialize["AlarmID"] = alarm.Id
	dataToSerialize["UserID"] = alarm.UserId
	dataToSerialize["Status"] = alarm.Status
	dataToSerialize["ChangedAt"] = changedAt.Format(time.RFC3339)

	serializedInfo, err := json.Marshal(dataToSerialize)
	if err != nil {
		log.Error("error during json marshalling, error: ", err)
		return nil, err
	}

	listenerId := rand.Intn(totalListeners)
	topicName := AlarmStatusChangedTopic + "." + strconv.Itoa(listenerId)

	pubErr := PublishMessage(conn, topicName, serializedInfo)
	if pubErr != nil {
		log.Error("error during publish message, error: ", pubErr)
		return nil, pubErr
	}


	pubErr2 := PublishMessage(conn, AlarmStatusChangedTopic, serializedInfo)
	if pubErr2 != nil {
		log.Error("error during publish message, error: ", pubErr)
		return nil, pubErr2
	}

	return serializedInfo, nil
}

func EmulateSendAlarmDigestMessage(conn *nats.Conn, userId string, totalListeners int) ([]byte, error) {

	dataToSerialize := map[string]interface{}{}
	dataToSerialize["UserID"] = userId

	serializedInfo, err := json.Marshal(dataToSerialize)
	if err != nil {
		log.Error("error during json marshalling, error: ", err)
		return nil, err
	}

	listenerId := rand.Intn(totalListeners)
	topicName := SendAlarmDigestTopic + "." + strconv.Itoa(listenerId)

	pubErr := PublishMessage(conn, topicName, serializedInfo)
	if pubErr != nil {
		log.Error("error during publish message, error: ", pubErr)
		return nil, pubErr
	}


	pubErr2 := PublishMessage(conn, SendAlarmDigestTopic, serializedInfo)
	if pubErr2 != nil {
		log.Error("error during publish message, error: ", pubErr)
		return nil, pubErr2
	}


	return serializedInfo, nil
}



