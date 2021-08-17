package main

import (
	. "alarm/error"
	. "alarm/fileutil"
	. "alarm/message"
	log "github.com/sirupsen/logrus"
)

func main() {


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
	props := ReadPropertiesFile("config.properties")
	props.PrintInfo()

	BootstrapAlarmService(serverConnection, &props, false)
}
