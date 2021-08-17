package main

import (
	. "alarm/error"
	. "alarm/fileutil"
	. "alarm/message"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)


const sendLogsToFile = false

func BootstrapAlarmService(serverConnection *nats.Conn, props *AppConfigProperties, produceTestTraffic bool) {

	// setup logging
	if sendLogsToFile {
		file, err := os.OpenFile("info.log", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			log.Fatal(err)
		}

		defer func(file *os.File) {
			_ = file.Close()
		}(file)

		log.SetOutput(file)
		log.SetFormatter(&log.JSONFormatter{})
	}
	log.SetLevel(log.DebugLevel)


	// setup error handling
	defer GlobalErrorHandler()


	// listen for incoming messages and process them for topics:  [AlarmStatusChanged, SendAlarmDigest]
	alarmStatusChangedMessagesChan := make(chan AlarmStatusChangedMessage,  props.FetchAsInt("alarmStatusChangedMessagesChan"))
	sendAlarmDigestMessagesChan := make(chan SendAlarmDigestMessage, props.FetchAsInt("sendAlarmDigestMessagesChan"))
	alarmDigestMessagesChan := make(chan AlarmDigestMessage, props.FetchAsInt("alarmDigestMessagesChan"))

	defer func() {
		log.Debugf("[DEFER-CLEAR-RESOURCES] will close alarmStatusChangedMessagesChan, sendAlarmDigestMessagesChan and alarmDigestMessagesChan channels now...")
		close(alarmStatusChangedMessagesChan)
		close(sendAlarmDigestMessagesChan)
		close(alarmDigestMessagesChan)
	}()


	// PRODUCERS
	registerProducers(props, serverConnection, alarmDigestMessagesChan)


	// WORKERS
	alarmMessageWorkersByDistributionId := registerWorkers(props, alarmDigestMessagesChan)
	defer func() {
		log.Debugf("[DEFER-CLEAR-RESOURCES] will close alarmMessageWorkersByDistributionId now...")

		for _, v := range alarmMessageWorkersByDistributionId {
			close(*v.AlarmStatusChangedMessages)
			close(*v.SendAlarmDigestMessages)
		}
	}()


	// CONSUMERS
	registerConsumers(props, alarmStatusChangedMessagesChan, alarmMessageWorkersByDistributionId, sendAlarmDigestMessagesChan)


	// LISTENERS
	alarmStatusChangedTopicSubscriptions,
	sendAlarmDigestTopicSubscriptions,
	alarmDigestTopicSubscription := registerListeners(props, serverConnection, alarmStatusChangedMessagesChan, sendAlarmDigestMessagesChan)
	defer func() {
		log.Debugf("[DEFER-CLEAR-RESOURCES] will close alarmStatusChangedTopicSubscriptions now...")

		for _, subscription := range alarmStatusChangedTopicSubscriptions {
			AsyncUnsubscribe(subscription, subscription.Subject)
		}
	}()
	defer func() {
		log.Debugf("[DEFER-CLEAR-RESOURCES] will close sendAlarmDigestTopicSubscriptions now...")

		for _, subscription := range sendAlarmDigestTopicSubscriptions {
			AsyncUnsubscribe(subscription, subscription.Subject)
		}
	}()
	defer func() {
		log.Debugf("[DEFER-CLEAR-RESOURCES] will close alarmDigestTopicSubscription now...")
		AsyncUnsubscribe(alarmDigestTopicSubscription, alarmDigestTopicSubscription.Subject)
	}()



	// --- produce some test traffic ---
	alarmStatusChangedListeners := props.FetchAsInt("alarmStatusChangedListeners")
	sendAlarmDigestListeners := props.FetchAsInt("sendAlarmDigestListeners")
	if produceTestTraffic {
		go EmulateTraffic(-1, serverConnection, alarmStatusChangedListeners, sendAlarmDigestListeners,
			500, 1000)
	}


	select {}
}

func registerListeners(props *AppConfigProperties, serverConnection *nats.Conn,
	alarmStatusChangedMessagesChan chan AlarmStatusChangedMessage, sendAlarmDigestMessagesChan chan SendAlarmDigestMessage) ([]*nats.Subscription, []*nats.Subscription, *nats.Subscription) {

	alarmStatusChangedListeners := props.FetchAsInt("alarmStatusChangedListeners")
	alarmStatusChangedTopicSubscriptions := RegisterAlarmStatusChangedTopicListeners("alarmStatusChangedTopicListener", alarmStatusChangedListeners, serverConnection,
		alarmStatusChangedMessagesChan, true, true)


	sendAlarmDigestListeners := props.FetchAsInt("sendAlarmDigestListeners")
	sendAlarmDigestTopicSubscriptions := RegisterSendAlarmDigestTopicListeners("sendAlarmDigestTopicListener", sendAlarmDigestListeners, serverConnection,
		sendAlarmDigestMessagesChan, true, true)


	// Note: this is used for debugging purposes and see if we produce correct message.
	alarmDigestTopicSubscription := RegisterAlarmDigestTopicListener("alarmDigestTopicListener", serverConnection, AlarmDigestTopic,
		true, nil)

	return alarmStatusChangedTopicSubscriptions, sendAlarmDigestTopicSubscriptions, alarmDigestTopicSubscription
}


func registerConsumers(props *AppConfigProperties, alarmStatusChangedMessagesChan chan AlarmStatusChangedMessage, alarmMessageWorkersByDistributionId map[DistributionId]*AlarmMessageWorker, sendAlarmDigestMessagesChan chan SendAlarmDigestMessage) {
	alarmStatusChangedMessagesConsumers := props.FetchAsInt("alarmStatusChangedMessagesConsumers")
	for i := 0; i < alarmStatusChangedMessagesConsumers; i++ {
		consumerId := "alarmStatusChangedConsumer#" + strconv.Itoa(i)
		go AlarmStatusChangedMessagesConsumer(alarmStatusChangedMessagesChan, consumerId, &alarmMessageWorkersByDistributionId)
	}

	sendAlarmDigestMessagesConsumers := props.FetchAsInt("sendAlarmDigestMessagesConsumers")
	for i := 0; i < sendAlarmDigestMessagesConsumers; i++ {
		consumerId := "sendAlarmDigestConsumer#" + strconv.Itoa(i)
		go SendAlarmDigestMessagesConsumer(sendAlarmDigestMessagesChan, consumerId, &alarmMessageWorkersByDistributionId)
	}
}


func registerWorkers(props *AppConfigProperties, alarmDigestMessagesChan chan AlarmDigestMessage) map[DistributionId]*AlarmMessageWorker {
	alarmStatusChangedMessagesTotalWorkers := props.FetchAsInt("alarmStatusChangedMessagesTotalWorkers")
	alarmMessageWorkersByDistributionId := make(map[DistributionId]*AlarmMessageWorker)
	for i := 0; i < alarmStatusChangedMessagesTotalWorkers; i++ {

		alarmStatusChangedMessages := make(chan AlarmStatusChangedMessage, props.FetchAsInt("alarmStatusChangedMessages"))
		sendAlarmDigestMessages := make(chan SendAlarmDigestMessage, props.FetchAsInt("sendAlarmDigestMessages"))

		workerName := "alarmMessagesWorker#" + strconv.Itoa(i)
		worker := AlarmMessageWorkerCreateNew(&workerName, &i, &alarmStatusChangedMessages, &sendAlarmDigestMessages, &alarmDigestMessagesChan)

		go worker.Consume()
		alarmMessageWorkersByDistributionId[DistributionId(i)] = worker

	}
	return alarmMessageWorkersByDistributionId
}


func registerProducers(props *AppConfigProperties, serverConnection *nats.Conn, alarmDigestMessagesChan chan AlarmDigestMessage) {
	alarmDigestMessageProducers := props.FetchAsInt("alarmDigestMessageProducers")
	for i := 0; i < alarmDigestMessageProducers; i++ {

		workerName := "alarmDigestMessageProducer#" + strconv.Itoa(i)
		topicName := AlarmDigestTopic
		worker := AlarmDigestMessageProducerCreateNew(&workerName, &i, &topicName, serverConnection, &alarmDigestMessagesChan)

		go worker.Produce()
	}
}



