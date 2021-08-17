package message

import (
	. "alarm/error"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"time"
)


func ServerConnection() (*nats.Conn, error) {

	nc, err := nats.Connect(nats.DefaultURL,

		nats.Name("Alarm Digest Service"),
		nats.Timeout(4*time.Second),

		nats.MaxPingsOutstanding(5),
		nats.MaxReconnects(10),
		nats.ReconnectWait(7*time.Second),

		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
			if s != nil {
				log.Errorf("Async error in %q/%q: %v", s.Subject, s.Queue, err)
			} else {
				log.Errorf("Async error outside subscription: %v", err)
			}
		}),

		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			// handle disconnect error event
			log.Warnf("disconnected from the server id: %v --- error: %v\n", nc.ConnectedServerId(), err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			// handle reconnect event
			log.Warnf("reconnected to server id: %v\n", nc.ConnectedServerId())
		}))

	if err != nil {
		return nil, err
	}

	log.Infof("connected to server: %v --- addr: %v --- serverId: %v\n",
		nc.ConnectedServerName(),
		nc.ConnectedAddr(),
		nc.ConnectedServerId())

	mp := nc.MaxPayload()
	log.Infof("Maximum payload is %v bytes", mp)

	return nc, nil
}

func AsyncSubscribe(conn *nats.Conn, topicName string, consumer func(msg *nats.Msg)) *nats.Subscription {

	defer GlobalErrorHandler()

	sub, err := conn.Subscribe(topicName, func(m *nats.Msg) {
		consumer(m)
	})

	if err != nil {
		panic(CouldNotSubscribeToTopicError{Msg: "topic-name --> " + topicName})
	}

	return sub
}


func AsyncUnsubscribe(sub *nats.Subscription, topicName string) {
	defer GlobalErrorHandler()

	if err := sub.Unsubscribe(); err != nil {
		panic(CouldNotUnsubscribeFromTopicError{Msg: "topic-name --> " + topicName})
	}
}


func PublishMessage(conn *nats.Conn, topicName string, msg []byte) error {
	return conn.Publish(topicName, msg)
}

