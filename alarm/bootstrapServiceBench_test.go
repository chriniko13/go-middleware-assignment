package main

import (
	. "alarm/error"
	. "alarm/fileutil"
	. "alarm/message"
	log "github.com/sirupsen/logrus"
	"testing"
	"time"
)


func BenchmarkTestBootstrapAlarmService_1(t *testing.B) {



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
	props := ReadPropertiesFile("test-config-bench.properties")
	props.PrintInfo()

	go BootstrapAlarmService(serverConnection, &props, false)




	// when - then
	alarmStatusChangedListeners := props.FetchAsInt("alarmStatusChangedListeners")
	sendAlarmDigestListeners := props.FetchAsInt("sendAlarmDigestListeners")
	EmulateTraffic(200, serverConnection, alarmStatusChangedListeners, sendAlarmDigestListeners, 60, 15)



	time.Sleep(time.Millisecond * 1500)

}
