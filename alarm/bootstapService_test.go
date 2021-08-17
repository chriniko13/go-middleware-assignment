package main

import (
	"alarm/domain"
	. "alarm/error"
	. "alarm/fileutil"
	. "alarm/message"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
	"testing"
	"time"
)


/*
	Scenario:

	* Produce a AlarmStatusChanged message for user X with critical status
	* Produce a AlarmStatusChanged message for user Y with warning status
	* Produce a AlarmStatusChanged message for user Z with critical status
	* Produce a AlarmStatusChanged message for user Z for previous with cleared status

	* Produce a SendAlarmDigest for user X, Y and Z, so from the above only 2 alarms
      should have produced to AlarmDigest topic

	* Make sure that necessary AlarmDigest messages have been produced

 */
func TestBootstrapAlarmService_integrationTest1(t *testing.T) {



	// given


	// get the connection to nats server
	serverConnection, err := ServerConnection()
	if err != nil {
		panic(CouldNotConnectToServerError{Msg: err.Error()})
	}
	defer func() {
		log.Debugf("[DEFER-CLEAR-RESOURCES] will close nats server connection now...")
		serverConnection.Close()
	}()

	// get properties
	props := ReadPropertiesFile("test-config.properties")
	props.PrintInfo()

	go BootstrapAlarmService(serverConnection, &props, false)


	alarmStatusChangedListeners := props.FetchAsInt("alarmStatusChangedListeners")
	sendAlarmDigestListeners := props.FetchAsInt("sendAlarmDigestListeners")

	// when

	// user X
	alarmIdX, _ := uuid.NewUUID()
	userIdX, _ := uuid.NewUUID()
	_, _ = EmulateAlarmStatusChangedMessage(serverConnection, &domain.Alarm{
		Id:        domain.AlarmId(alarmIdX.String()),
		UserId:    domain.UserId(userIdX.String()),
		Status:    domain.CRITICAL,
		CreatedAt: time.Now(),
	}, time.Now(), alarmStatusChangedListeners)

	// user Y
	alarmIdY, _ := uuid.NewUUID()
	userIdY, _ := uuid.NewUUID()
	_, _ = EmulateAlarmStatusChangedMessage(serverConnection, &domain.Alarm{
		Id:        domain.AlarmId(alarmIdY.String()),
		UserId:    domain.UserId(userIdY.String()),
		Status:    domain.WARNING,
		CreatedAt: time.Now(),
	}, time.Now(), alarmStatusChangedListeners)

	// user Z
	alarmIdZ, _ := uuid.NewUUID()
	userIdZ, _ := uuid.NewUUID()
	_, _ = EmulateAlarmStatusChangedMessage(serverConnection, &domain.Alarm{
		Id:        domain.AlarmId(alarmIdZ.String()),
		UserId:    domain.UserId(userIdZ.String()),
		Status:    domain.CRITICAL,
		CreatedAt: time.Now(),
	}, time.Now(), alarmStatusChangedListeners)

	_, _ = EmulateAlarmStatusChangedMessage(serverConnection, &domain.Alarm{
		Id:        domain.AlarmId(alarmIdZ.String()),
		UserId:    domain.UserId(userIdZ.String()),
		Status:    domain.CLEARED,
		CreatedAt: time.Now(),
	}, time.Now(), alarmStatusChangedListeners)

	time.Sleep(time.Millisecond * 400)

	_, _ = EmulateSendAlarmDigestMessage(serverConnection, userIdX.String(), sendAlarmDigestListeners)

	_, _ = EmulateSendAlarmDigestMessage(serverConnection, userIdY.String(), sendAlarmDigestListeners)

	_, _ = EmulateSendAlarmDigestMessage(serverConnection, userIdZ.String(), sendAlarmDigestListeners)





	// then
	var totalMessagesReceived uint64 = 0

	sub := RegisterAlarmDigestTopicListener("testRegisterAlarmDigestListener", serverConnection,
		AlarmDigestTopic, true, func(message AlarmDigestMessage) {

			fmt.Printf("~~~received alarm digest message %v~~~\n", message)
			atomic.AddUint64(&totalMessagesReceived, 1)
		})
	defer func() {
		AsyncUnsubscribe(sub, sub.Subject)
	}()

	tries := 15
	currentTry := 0
	for {

		received := atomic.LoadUint64(&totalMessagesReceived)
		fmt.Printf("received: %v --- current try: %v\n", received, currentTry)

		if received == 2 {
			break
		}

		if currentTry > tries {
			t.Fatal("alarm service does not fulfil specification")
		}

		currentTry++


		time.Sleep(time.Millisecond * 750)
	}





}
