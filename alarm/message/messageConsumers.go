package message

import (
	. "alarm/hashing"
	"errors"
	log "github.com/sirupsen/logrus"
	"strconv"
)

// ------------------------------------------------------------------------------------------


type DistributionId uint32



func calculateDistributionId(consumerId *string, userId *string, totalWorkers uint32) DistributionId {
	hashedUserId := Hash(*userId)
	channelIdToSend := DistributionId(hashedUserId % totalWorkers)

	log.Infof("[%v] hashedUserId: %v --- channelIdToSend: %v\n", *consumerId, hashedUserId, channelIdToSend)

	return channelIdToSend
}


// ------------------------------------------------------------------------------------------




func AlarmStatusChangedMessagesConsumer(msgs <-chan AlarmStatusChangedMessage,
										consumerId string,
										alarmStatusChangedMessagesWorkersByDistributionId *map[DistributionId]*AlarmMessageWorker) {

	var totalWorkers = uint32(len(*alarmStatusChangedMessagesWorkersByDistributionId))

	for msg := range msgs {

		log.Infof("[%v] AlarmStatusChangedMessage received: %v\n", consumerId, msg)

		channelIdToSend := calculateDistributionId(&consumerId, &msg.UserId, totalWorkers)

		worker, found := (*alarmStatusChangedMessagesWorkersByDistributionId)[channelIdToSend]
		if !found {
			panic(errors.New("could not found a worker for distribution id: " + strconv.Itoa(int(channelIdToSend))))
		}

		*worker.AlarmStatusChangedMessages <- msg
	}

	log.Printf("alarmStatusChangedConsumer with id: %v exiting...\n", consumerId)

}


func SendAlarmDigestMessagesConsumer(msgs <-chan SendAlarmDigestMessage,
	                                 consumerId string,
	                                 alarmStatusChangedMessagesWorkersByDistributionId *map[DistributionId]*AlarmMessageWorker) {

	var totalWorkers = uint32(len(*alarmStatusChangedMessagesWorkersByDistributionId))

	for msg := range msgs {

		log.Infof("[%v] SendAlarmDigestMessage received: %v\n", consumerId, msg)

		channelIdToSend := calculateDistributionId(&consumerId, &msg.UserId, totalWorkers)

		worker, found := (*alarmStatusChangedMessagesWorkersByDistributionId)[channelIdToSend]
		if !found {
			panic(errors.New("could not found a worker for distribution id: " + strconv.Itoa(int(channelIdToSend))))
		}

		*worker.SendAlarmDigestMessages <- msg

	}

	log.Infof("sendAlarmDigestMessagesConsumer with id: %v exiting...\n", consumerId)
}


