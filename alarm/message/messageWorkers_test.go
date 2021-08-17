package message

import (
	. "alarm/domain"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"testing"
	"time"
)

func TestAlarmMessageWorker_handleAlarmStatusChangeMessage_noAlarmExists(t *testing.T) {

	// given
	i := 0

	alarmStatusChangedMessages := make(chan AlarmStatusChangedMessage, 75)
	sendAlarmDigestMessages := make(chan SendAlarmDigestMessage, 75)
	alarmDigestMessagesChan := make(chan AlarmDigestMessage, 500)

	defer func() {
		close(alarmStatusChangedMessages)
		close(sendAlarmDigestMessages)
		close(alarmDigestMessagesChan)
	}()

	workerName := "alarmMessagesWorker#" + strconv.Itoa(i)
	worker := AlarmMessageWorkerCreateNew(&workerName, &i, &alarmStatusChangedMessages, &sendAlarmDigestMessages, &alarmDigestMessagesChan)

	alarmId, _ := uuid.NewUUID()
	userId, _ := uuid.NewUUID()


	msg := &AlarmStatusChangedMessage{
		AlarmId:   alarmId.String(),
		UserId:    userId.String(),
		Status:    AlarmStatusAsString(CRITICAL),
		ChangedAt: time.Now(),
	}

	// when
	worker.handleAlarmStatusChangeMessage(msg)


	// then
	domainAlarmId := AlarmId(alarmId.String())
	domainUserId := UserId(userId.String())
	alarm := worker.alarmsState[domainUserId][domainAlarmId]

	if alarm.Status != CRITICAL {
		t.Errorf("alarm status is invalid")
	}
	if alarm.Id != domainAlarmId {
		t.Errorf("alarm id is invalid")
	}
	if alarm.UserId != domainUserId {
		t.Errorf("alarm user id is invalid")
	}
	if alarm.CreatedAt.IsZero() {
		t.Errorf("alarm created at is invalid")
	}
	if !alarm.UpdatedAt.IsZero() {
		t.Errorf("alarm updated at is invalid")
	}
	if !IsActiveAlarm(alarm) {
		t.Errorf("alarm is active state is invalid")
	}

	activeAlarm := worker.activeAlarms[domainUserId][domainAlarmId]

	if activeAlarm != alarm {
		t.Errorf("not correct state maintenance between alarmsState and activeAlarms")
	}


}


func TestAlarmMessageWorker_handleAlarmStatusChangeMessage_alarmExistsUpdateState(t *testing.T) {

	// given
	i := 0

	alarmStatusChangedMessages := make(chan AlarmStatusChangedMessage, 75)
	sendAlarmDigestMessages := make(chan SendAlarmDigestMessage, 75)
	alarmDigestMessagesChan := make(chan AlarmDigestMessage, 500)

	defer func() {
		close(alarmStatusChangedMessages)
		close(sendAlarmDigestMessages)
		close(alarmDigestMessagesChan)
	}()

	workerName := "alarmMessagesWorker#" + strconv.Itoa(i)
	worker := AlarmMessageWorkerCreateNew(&workerName, &i, &alarmStatusChangedMessages, &sendAlarmDigestMessages, &alarmDigestMessagesChan)

	alarmId, _ := uuid.NewUUID()
	userId, _ := uuid.NewUUID()


	msg := &AlarmStatusChangedMessage{
		AlarmId:   alarmId.String(),
		UserId:    userId.String(),
		Status:    AlarmStatusAsString(CRITICAL),
		ChangedAt: time.Now(),
	}
	worker.handleAlarmStatusChangeMessage(msg)


	// when
	newMsg := &AlarmStatusChangedMessage{
		AlarmId:   alarmId.String(),
		UserId:    userId.String(),
		Status:    AlarmStatusAsString(CRITICAL),
		ChangedAt: time.Now(),
	}
	worker.handleAlarmStatusChangeMessage(newMsg)


	// then
	domainAlarmId := AlarmId(alarmId.String())
	domainUserId := UserId(userId.String())
	alarm := worker.alarmsState[domainUserId][domainAlarmId]

	if alarm.Status != CRITICAL {
		t.Errorf("alarm status is invalid")
	}
	if alarm.Id != domainAlarmId {
		t.Errorf("alarm id is invalid")
	}
	if alarm.UserId != domainUserId {
		t.Errorf("alarm user id is invalid")
	}
	if alarm.CreatedAt.IsZero() {
		t.Errorf("alarm created at is invalid")
	}
	if alarm.UpdatedAt.IsZero() {
		t.Errorf("alarm updated at is invalid")
	}
	if !IsActiveAlarm(alarm) {
		t.Errorf("alarm is active state is invalid")
	}

	activeAlarm := worker.activeAlarms[domainUserId][domainAlarmId]

	if activeAlarm != alarm {
		t.Errorf("not correct state maintenance between alarmsState and activeAlarms")
	}


}



func TestAlarmMessageWorker_handleAlarmStatusChangeMessage_alarmExistsClearingState(t *testing.T) {

	// given
	i := 0

	alarmStatusChangedMessages := make(chan AlarmStatusChangedMessage, 75)
	sendAlarmDigestMessages := make(chan SendAlarmDigestMessage, 75)
	alarmDigestMessagesChan := make(chan AlarmDigestMessage, 500)

	defer func() {
		close(alarmStatusChangedMessages)
		close(sendAlarmDigestMessages)
		close(alarmDigestMessagesChan)
	}()

	workerName := "alarmMessagesWorker#" + strconv.Itoa(i)
	worker := AlarmMessageWorkerCreateNew(&workerName, &i, &alarmStatusChangedMessages, &sendAlarmDigestMessages, &alarmDigestMessagesChan)

	alarmId, _ := uuid.NewUUID()
	userId, _ := uuid.NewUUID()


	msg := &AlarmStatusChangedMessage{
		AlarmId:   alarmId.String(),
		UserId:    userId.String(),
		Status:    AlarmStatusAsString(CRITICAL),
		ChangedAt: time.Now(),
	}
	worker.handleAlarmStatusChangeMessage(msg)


	// when
	newMsg := &AlarmStatusChangedMessage{
		AlarmId:   alarmId.String(),
		UserId:    userId.String(),
		Status:    AlarmStatusAsString(CLEARED),
		ChangedAt: time.Now(),
	}
	worker.handleAlarmStatusChangeMessage(newMsg)




	// then
	domainAlarmId := AlarmId(alarmId.String())
	domainUserId := UserId(userId.String())
	alarm := worker.alarmsState[domainUserId][domainAlarmId]

	if alarm.Status == CRITICAL {
		t.Errorf("alarm status is invalid")
	}
	if alarm.Id != domainAlarmId {
		t.Errorf("alarm id is invalid")
	}
	if alarm.UserId != domainUserId {
		t.Errorf("alarm user id is invalid")
	}
	if alarm.CreatedAt.IsZero() {
		t.Errorf("alarm created at is invalid")
	}
	if alarm.UpdatedAt.IsZero() {
		t.Errorf("alarm updated at is invalid")
	}
	if IsActiveAlarm(alarm) {
		t.Errorf("alarm is active state is invalid")
	}

	_, exists := worker.activeAlarms[domainUserId][domainAlarmId]
	if exists {
		t.Errorf("not correct state maintenance between alarmsState and activeAlarms")
	}

}

func TestAlarmMessageWorker_handleSendAlarmDigestMessage_criticalAlarmNotExists(t *testing.T) {

	// given
	i := 0

	alarmStatusChangedMessages := make(chan AlarmStatusChangedMessage, 75)
	sendAlarmDigestMessages := make(chan SendAlarmDigestMessage, 75)
	alarmDigestMessagesChan := make(chan AlarmDigestMessage, 500)

	defer func() {
		close(alarmStatusChangedMessages)
		close(sendAlarmDigestMessages)
		close(alarmDigestMessagesChan)
	}()

	workerName := "alarmMessagesWorker#" + strconv.Itoa(i)

	worker := AlarmMessageWorkerCreateNew(&workerName, &i, &alarmStatusChangedMessages, &sendAlarmDigestMessages, &alarmDigestMessagesChan)

	userId, _ := uuid.NewUUID()

	msg := &SendAlarmDigestMessage{UserId: userId.String()}


	// when
	worker.handleSendAlarmDigestMessage(msg)


	// then
	select {
	case msg := <-*worker.alarmDigestMessages:
		t.Errorf("alarm digest message should have not produced, msg: %v\n", msg)

	default:
	}

}


func TestAlarmMessageWorker_handleSendAlarmDigestMessage_criticalAlarmExists(t *testing.T) {

	// given
	i := 0

	alarmStatusChangedMessages := make(chan AlarmStatusChangedMessage, 75)
	sendAlarmDigestMessages := make(chan SendAlarmDigestMessage, 75)
	alarmDigestMessagesChan := make(chan AlarmDigestMessage, 500)

	defer func() {
		close(alarmStatusChangedMessages)
		close(sendAlarmDigestMessages)
		close(alarmDigestMessagesChan)
	}()

	workerName := "alarmMessagesWorker#" + strconv.Itoa(i)

	worker := AlarmMessageWorkerCreateNew(&workerName, &i, &alarmStatusChangedMessages, &sendAlarmDigestMessages, &alarmDigestMessagesChan)

	alarmId, _ := uuid.NewUUID()
	userId, _ := uuid.NewUUID()


	msg := &AlarmStatusChangedMessage{
		AlarmId:   alarmId.String(),
		UserId:    userId.String(),
		Status:    AlarmStatusAsString(CRITICAL),
		ChangedAt: time.Now(),
	}
	worker.handleAlarmStatusChangeMessage(msg)


	sendAlarmDigestMsg := &SendAlarmDigestMessage{UserId: userId.String()}


	// when
	worker.handleSendAlarmDigestMessage(sendAlarmDigestMsg)


	// then
	select {
	case msg := <-*worker.alarmDigestMessages:
		fmt.Println("produced alarm digest message: ", msg)


	default:
		t.Errorf("alarm digest message should have produced\n")
	}

}


func TestAlarmMessageWorker_handleSendAlarmDigestMessage_criticalAlarmExistsSendAlarmDigestThenNoCriticalAlarmToSend(t *testing.T) {

	// given
	i := 0

	alarmStatusChangedMessages := make(chan AlarmStatusChangedMessage, 75)
	sendAlarmDigestMessages := make(chan SendAlarmDigestMessage, 75)
	alarmDigestMessagesChan := make(chan AlarmDigestMessage, 500)

	defer func() {
		close(alarmStatusChangedMessages)
		close(sendAlarmDigestMessages)
		close(alarmDigestMessagesChan)
	}()

	workerName := "alarmMessagesWorker#" + strconv.Itoa(i)

	worker := AlarmMessageWorkerCreateNew(&workerName, &i, &alarmStatusChangedMessages, &sendAlarmDigestMessages, &alarmDigestMessagesChan)

	alarmId, _ := uuid.NewUUID()
	userId, _ := uuid.NewUUID()


	msg := &AlarmStatusChangedMessage{
		AlarmId:   alarmId.String(),
		UserId:    userId.String(),
		Status:    AlarmStatusAsString(CRITICAL),
		ChangedAt: time.Now(),
	}
	worker.handleAlarmStatusChangeMessage(msg)


	sendAlarmDigestMsg := &SendAlarmDigestMessage{UserId: userId.String()}


	// when
	worker.handleSendAlarmDigestMessage(sendAlarmDigestMsg)


	// then
	select {
	case msg := <-*worker.alarmDigestMessages:
		fmt.Println("produced alarm digest message: ", msg)


	default:
		t.Errorf("alarm digest message should have produced\n")
	}


	// when
	worker.handleSendAlarmDigestMessage(sendAlarmDigestMsg)


	// then
	select {
	case msg := <-*worker.alarmDigestMessages:
		t.Errorf("alarm digest message should not have produced, msg: %v\n", msg)


	default:
	}
}

func TestAlarmMessageWorker_handleSendAlarmDigestMessage_criticalAlarmExistsButClearedStatusSoNoSendAlarmDigest(t *testing.T) {

	// given
	i := 0

	alarmStatusChangedMessages := make(chan AlarmStatusChangedMessage, 75)
	sendAlarmDigestMessages := make(chan SendAlarmDigestMessage, 75)
	alarmDigestMessagesChan := make(chan AlarmDigestMessage, 500)

	defer func() {
		close(alarmStatusChangedMessages)
		close(sendAlarmDigestMessages)
		close(alarmDigestMessagesChan)
	}()

	workerName := "alarmMessagesWorker#" + strconv.Itoa(i)

	worker := AlarmMessageWorkerCreateNew(&workerName, &i, &alarmStatusChangedMessages, &sendAlarmDigestMessages, &alarmDigestMessagesChan)

	alarmId, _ := uuid.NewUUID()
	userId, _ := uuid.NewUUID()


	msg := &AlarmStatusChangedMessage{
		AlarmId:   alarmId.String(),
		UserId:    userId.String(),
		Status:    AlarmStatusAsString(CRITICAL),
		ChangedAt: time.Now(),
	}
	worker.handleAlarmStatusChangeMessage(msg)

	msg2 := &AlarmStatusChangedMessage{
		AlarmId:   alarmId.String(),
		UserId:    userId.String(),
		Status:    AlarmStatusAsString(CLEARED),
		ChangedAt: time.Now(),
	}
	worker.handleAlarmStatusChangeMessage(msg2)


	sendAlarmDigestMsg := &SendAlarmDigestMessage{UserId: userId.String()}


	// when
	worker.handleSendAlarmDigestMessage(sendAlarmDigestMsg)


	// then
	select {
	case msg := <-*worker.alarmDigestMessages:
		t.Errorf("alarm digest message should nothave produced, msg: %v\n", msg)

	default:
	}

}