package message

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"math"
	"strconv"
	"sync/atomic"
)

func RegisterAlarmDigestTopicListener(listenerName string, serverConnection *nats.Conn,
									topicName string, ignoreBadFormattedMessages bool,
									messageConsumer func(message AlarmDigestMessage)) *nats.Subscription {

	topicSubscription := AsyncSubscribe(serverConnection, topicName, func(msg *nats.Msg) {

		var dat AlarmDigestMessage
		if err := json.Unmarshal(msg.Data, &dat); err != nil {

			if ignoreBadFormattedMessages {
				log.Warnf("[%v] bad formatted message received from topic: %v, data: %v\n", listenerName, topicName, string(msg.Data))
			} else {
				panic(err)
			}
		}

		log.Debugf("[%v] !!!!!!!!!!!!!!!!!!!!!!!! SendAlarmDigestMessage received from topic: %v, data: %v\n", listenerName, topicName, dat)


		if messageConsumer != nil {
			messageConsumer(dat)
		}
	})

	return topicSubscription
}


func RegisterSendAlarmDigestTopicListeners(listenerNamePrefix string, sendAlarmDigestListeners int,
											serverConnection *nats.Conn, sendAlarmDigestMessagesChan chan SendAlarmDigestMessage,
											registerDelegator bool, ignoreBadFormattedMessages bool) []*nats.Subscription {

	var sendAlarmDigestTopicSubscriptions = make([]*nats.Subscription, sendAlarmDigestListeners)
	for i := 0; i < sendAlarmDigestListeners; i++ {

		topicName := SendAlarmDigestTopic + "." + strconv.Itoa(i)

		listenerName := listenerNamePrefix + "#" + strconv.Itoa(i)

		sendAlarmDigestTopicSubscription := AsyncSubscribe(serverConnection, topicName, func(msg *nats.Msg) {

			var dat SendAlarmDigestMessage
			if err := json.Unmarshal(msg.Data, &dat); err != nil {
				if ignoreBadFormattedMessages {
					log.Warnf("[%v] bad formatted message received from topic: %v, data: %v\n", listenerName, topicName, string(msg.Data))
				} else {
					panic(err)
				}
			}

			log.Infof("[%v] message received from topic: %v, data: %v\n", listenerName, topicName, dat)
			sendAlarmDigestMessagesChan <- dat

		})

		sendAlarmDigestTopicSubscriptions = append(sendAlarmDigestTopicSubscriptions, sendAlarmDigestTopicSubscription)
	}


	/*
		Note: because the provided test emits message to SendAlarmDigest topic, and I have implemented a solution
			  to scale listeners, so each one listens to .<id> suffix, so: SendAlarmDigest.<id> we will need to have
	          a listener which will listen on SendAlarmDigest topic as is, and delegate the messages to other listeners,
	          in our case with round-robin fashion.
	 */
	if registerDelegator {

		var currentCounter uint64 = 0

		subscription := AsyncSubscribe(serverConnection, SendAlarmDigestTopic, func(msg *nats.Msg) {
			var dat SendAlarmDigestMessage
			if err := json.Unmarshal(msg.Data, &dat); err != nil {
				panic(err)
			}

			topicName := SendAlarmDigestTopic
			listenerName := SendAlarmDigestTopic + "#Delegator"
			log.Infof("[%v] message received from topic: %v, data: %v\n", listenerName, topicName, dat)


			listenerIdToDelegate := atomic.LoadUint64(&currentCounter) % uint64(sendAlarmDigestListeners)
			if listenerIdToDelegate < 0 || listenerIdToDelegate >= uint64(sendAlarmDigestListeners) {
				panic("error occurred during calculation of listener id to delegate")
			}
			topicNameOfListenerToDelegate := SendAlarmDigestTopic + "." + strconv.FormatUint(listenerIdToDelegate, 10)


			log.Infof("[%v] will delegate now to topic: %v\n", listenerName, topicNameOfListenerToDelegate)


			err := PublishMessage(serverConnection, topicNameOfListenerToDelegate, msg.Data)
			if err != nil {
				if ignoreBadFormattedMessages {
					log.Warnf("[%v] bad formatted message received from topic: %v, data: %v\n", listenerName, topicName, string(msg.Data))
				} else {
					panic(err)
				}
			}

			if atomic.LoadUint64(&currentCounter) >= math.MaxUint32 - 10 {
				atomic.CompareAndSwapUint64(&currentCounter, currentCounter, 0)
			} else {
				atomic.AddUint64(&currentCounter, 1)
			}


		})

		sendAlarmDigestTopicSubscriptions = append(sendAlarmDigestTopicSubscriptions, subscription)
	}


	return sendAlarmDigestTopicSubscriptions
}

func RegisterAlarmStatusChangedTopicListeners(listenerNamePrefix string, alarmStatusChangedListeners int,
												serverConnection *nats.Conn, alarmStatusChangedMessagesChan chan AlarmStatusChangedMessage,
												registerDelegator bool, ignoreBadFormattedMessages bool) []*nats.Subscription {


	var alarmStatusChangedTopicSubscriptions = make([]*nats.Subscription, alarmStatusChangedListeners)
	for i := 0; i < alarmStatusChangedListeners; i++ {

		topicName := AlarmStatusChangedTopic + "." + strconv.Itoa(i)

		listenerName := listenerNamePrefix + "#" + strconv.Itoa(i)

		alarmStatusChangedTopicSubscription := AsyncSubscribe(serverConnection, topicName, func(msg *nats.Msg) {

			var dat AlarmStatusChangedMessage
			if err := json.Unmarshal(msg.Data, &dat); err != nil {
				if ignoreBadFormattedMessages {
					log.Warnf("[%v] bad formatted message received from topic: %v, data: %v\n", listenerName, topicName, string(msg.Data))
				} else {
					panic(err)
				}
			}

			log.Infof("[%v] message received from topic: %v, data: %v\n", listenerName, topicName, dat)
			alarmStatusChangedMessagesChan <- dat

		})

		alarmStatusChangedTopicSubscriptions = append(alarmStatusChangedTopicSubscriptions, alarmStatusChangedTopicSubscription)
	}


	/*
		Note: because the provided test emits message to AlarmStatusChangedTopic topic, and I have implemented a solution
			  to scale listeners, so each one listens to .<id> suffix, so: AlarmStatusChangedTopic.<id> we will need to have
			  a listener which will listen on SendAlarmDigest topic as is, and delegate the messages to other listeners,
			  in our case with round-robin fashion.
	*/
	if registerDelegator {


		var currentCounter uint64 = 0

		subscription := AsyncSubscribe(serverConnection, AlarmStatusChangedTopic, func(msg *nats.Msg) {
			var dat AlarmStatusChangedMessage
			if err := json.Unmarshal(msg.Data, &dat); err != nil {
				panic(err)
			}

			topicName := AlarmStatusChangedTopic
			listenerName := AlarmStatusChangedTopic + "#Delegator"
			log.Infof("[%v] message received from topic: %v, data: %v\n", listenerName, topicName, dat)


			listenerIdToDelegate := atomic.LoadUint64(&currentCounter) % uint64(alarmStatusChangedListeners)
			if listenerIdToDelegate < 0 || listenerIdToDelegate >= uint64(alarmStatusChangedListeners) {
				panic("error occurred during calculation of listener id to delegate")
			}
			topicNameOfListenerToDelegate := AlarmStatusChangedTopic + "." + strconv.FormatUint(listenerIdToDelegate, 10)


			log.Infof("[%v] will delegate now to topic: %v\n", listenerName, topicNameOfListenerToDelegate)


			err := PublishMessage(serverConnection, topicNameOfListenerToDelegate, msg.Data)
			if err != nil {
				if ignoreBadFormattedMessages {
					log.Warnf("[%v] bad formatted message received from topic: %v, data: %v\n", listenerName, topicName, string(msg.Data))
				} else {
					panic(err)
				}
			}

			if atomic.LoadUint64(&currentCounter) >= math.MaxUint32 - 10 {
				atomic.CompareAndSwapUint64(&currentCounter, currentCounter, 0)
			} else {
				atomic.AddUint64(&currentCounter, 1)
			}

		})

		alarmStatusChangedTopicSubscriptions = append(alarmStatusChangedTopicSubscriptions, subscription)
	}

	return alarmStatusChangedTopicSubscriptions
}


