package message

import (
	. "alarm/domain"
	"encoding/json"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"sort"
	"time"
)


// -------------------

type Worker struct {

	workerId int
	workerName string
}


// -------------------


type AlarmDigestMessageProducer struct {
	Worker

	topicName *string

	serverConnection * nats.Conn

	alarmDigestMessages *chan AlarmDigestMessage
}


func AlarmDigestMessageProducerCreateNew(workerName *string,
	workerId *int,
	topicName *string,
	serverConnection * nats.Conn,
	alarmDigestMessages *chan AlarmDigestMessage) *AlarmDigestMessageProducer {


	w := & AlarmDigestMessageProducer{
		Worker:       Worker{
			workerId:   *workerId,
			workerName: *workerName,
		},
		topicName: topicName,
		serverConnection: serverConnection,
		alarmDigestMessages: alarmDigestMessages,
	}

	return w
}

func (w *AlarmDigestMessageProducer) Produce() {

	for {

		select {
		case alarmDigestMessage := <-*w.alarmDigestMessages:
			log.Infof("[%v] worker received AlarmDigestMessage message: %v\n", w.workerName, alarmDigestMessage)

			serializedInfo, serializationError := json.Marshal(alarmDigestMessage)
			if serializationError != nil {
				log.Infoln("error during json marshalling, error: ", serializationError)
				panic(serializationError)
			}

			err := PublishMessage(w.serverConnection, *w.topicName, serializedInfo)
			if err != nil {
				log.Errorf("could not publish alarm digest message, err: %v\n", err)
			}
		}

	}
}

// -------------------

const printDebugMessageOfWorkerState = false

type AlarmMessageWorker struct {
	Worker

	AlarmStatusChangedMessages *chan AlarmStatusChangedMessage
	SendAlarmDigestMessages *chan SendAlarmDigestMessage

	alarmDigestMessages *chan AlarmDigestMessage

	alarmsState map[UserId]map[AlarmId]*Alarm
	// Note: we hold here the active alarms so during SendAlarmDigest message to calculate it more efficient in terms of time complexity.
	activeAlarms map[UserId]map[AlarmId]*Alarm
}

func AlarmMessageWorkerCreateNew(workerName *string,
								  workerId *int,
								  alarmStatusChangedMessages *chan AlarmStatusChangedMessage,
								  sendAlarmDigestMessages *chan SendAlarmDigestMessage,
								  alarmDigestMessages *chan AlarmDigestMessage) *AlarmMessageWorker {

	w := &AlarmMessageWorker{}

	w.workerName = *workerName
	w.workerId = *workerId

	w.AlarmStatusChangedMessages = alarmStatusChangedMessages
	w.SendAlarmDigestMessages = sendAlarmDigestMessages
	w.alarmDigestMessages = alarmDigestMessages

	w.alarmsState = make(map[UserId]map[AlarmId]*Alarm)
	w.activeAlarms = make(map[UserId]map[AlarmId]*Alarm)

	return w
}

func (w *AlarmMessageWorker) Consume() {

	for {

		select {
		case alarmStatusChangeMessage := <-*w.AlarmStatusChangedMessages:
			log.Infof("[%v] WORKER RECEIVED AlarmStatusChangedMessages message: %v\n", w.workerName, alarmStatusChangeMessage)

			w.handleAlarmStatusChangeMessage(&alarmStatusChangeMessage)

			if printDebugMessageOfWorkerState {
				log.Debugf("[%v] alarmsState: %v\n", w.alarmsState, alarmStatusChangeMessage)
				log.Debugf("[%v] activeAlarms: %v\n", w.activeAlarms, alarmStatusChangeMessage)
			}

		case sendAlarmDigestMessage := <-*w.SendAlarmDigestMessages:
			log.Infof("[%v] WORKER RECEIVED SendAlarmDigestMessage message: %v\n", w.workerName, sendAlarmDigestMessage)

			w.handleSendAlarmDigestMessage(&sendAlarmDigestMessage)
		}
	}
}

func (w *AlarmMessageWorker) handleSendAlarmDigestMessage(msg *SendAlarmDigestMessage) {

	userId := UserId(msg.UserId)

	activeAlarms, activeAlarmsExistForUser := w.activeAlarms[userId]
	if activeAlarmsExistForUser && len(activeAlarms) != 0 {

		// Note: collect them.
		alarms := make([]ActiveAlarm, 0, len(activeAlarms))

		for _, alarm := range activeAlarms {

			var timestamp time.Time
			if alarm.UpdatedAt.IsZero() {
				timestamp = alarm.CreatedAt
			} else {
				timestamp = alarm.UpdatedAt
			}

			activeAlarm := ActiveAlarm{
				AlarmId:         string(alarm.Id),
				Status:          AlarmStatusAsString(alarm.Status),
				LatestChangedAt: timestamp,
			}

			alarms = append(alarms, activeAlarm)
		}


		// Note: sort them (oldest to newest)
		sort.Slice(alarms, func(i, j int) bool {
			return alarms[i].LatestChangedAt.Before(alarms[j].LatestChangedAt)
		})


		// Note: send them
		alarmDigestMessage := AlarmDigestMessage{
			UserId:       msg.UserId,
			ActiveAlarms: alarms,
		}

		log.Infof("[%v] worker will send AlarmDigestMessage message: %v\n", w.workerName, alarmDigestMessage)
		*w.alarmDigestMessages <- alarmDigestMessage


		// Note: delete sent active alarms for user so that to not received them twice.
		for k, _ := range activeAlarms {
			delete(activeAlarms, k)
		}

	} else {
		log.Infof("[%v] worker will NOT send AlarmDigestMessage message - NO CRITICAL ALARMS\n", w.workerName)
	}

}


func (w *AlarmMessageWorker) handleAlarmStatusChangeMessage(msg *AlarmStatusChangedMessage) {

	userId := UserId(msg.UserId)
	alarmId := AlarmId(msg.AlarmId)
	newAlarmStatus := ExtractAlarmStatus(msg.Status)
	changedAt := msg.ChangedAt

	alarms, alarmsExistsForSelectedUser := w.alarmsState[userId]
	if alarmsExistsForSelectedUser {

		selectedAlarmState, selectedAlarmExist := alarms[alarmId]
		if selectedAlarmExist {

			if newAlarmStatus != UNKNOWN {
				selectedAlarmState.Status = newAlarmStatus
			}
			selectedAlarmState.UpdatedAt = changedAt

			w.updateActiveAlarms(selectedAlarmState, userId, alarmId)

		} else {
			alarm := w.constructNewAlarm(alarmId, userId, newAlarmStatus, changedAt)
			alarms[alarmId] = alarm

			w.updateActiveAlarms(alarm, userId, alarmId)
		}

	} else {

		alarmsForUser := make(map[AlarmId]*Alarm)
		alarm := w.constructNewAlarm(alarmId, userId, newAlarmStatus, changedAt)
		alarmsForUser[alarmId] = alarm

		w.alarmsState[userId] = alarmsForUser

		w.updateActiveAlarms(alarm, userId, alarmId)
	}
}

func (w *AlarmMessageWorker) updateActiveAlarms(alarm *Alarm, userId UserId, alarmId AlarmId) {
	if IsActiveAlarm(alarm) {

		activeAlarmsForSelectedUsers, activeAlarmsExistForSelectedUser := w.activeAlarms[userId]

		if activeAlarmsExistForSelectedUser {
			// Note: just update it - upsert.
			activeAlarmsForSelectedUsers[alarmId] = alarm
		} else {
			// Note: otherwise create new structure and store it.
			activeAlarmsForUser := make(map[AlarmId]*Alarm)
			activeAlarmsForUser[alarmId] = alarm

			w.activeAlarms[userId] = activeAlarmsForUser
		}

	} else {
		// Note: not active alarm - check if an active record exists and clear it.

		activeAlarmsForSelectedUsers, activeAlarmsExistForSelectedUser := w.activeAlarms[userId]
		if activeAlarmsExistForSelectedUser {

			_, selectedActiveAlarmExist := activeAlarmsForSelectedUsers[alarmId]
			if selectedActiveAlarmExist {
				// Note: clear status
				delete(activeAlarmsForSelectedUsers, alarmId)
			}
		}

	}
}


func (w *AlarmMessageWorker) constructNewAlarm(alarmId AlarmId, userId UserId, newAlarmStatus AlarmStatus, timestamp time.Time) *Alarm {
	alarm := &Alarm{
		Id:        alarmId,
		UserId:    userId,
		Status:    newAlarmStatus,
		CreatedAt: timestamp,
	}
	return alarm
}

// -------------------


