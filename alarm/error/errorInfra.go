package error

import (
	log "github.com/sirupsen/logrus"
)

// ---- error handling ----

func GlobalErrorHandler() {
	if err := recover(); err != nil {

		switch errorMsg := err.(type) {

		// Note: add more error handlers based on type.

		case CouldNotConnectToServerError:
			log.Fatal("could not connect to nats server, message: ", errorMsg.Details())

		case CouldNotSubscribeToTopicError:
			log.Fatal("could not subscribe to topic, message: ", errorMsg.Details())

		case CouldNotUnsubscribeFromTopicError:
			log.Fatal("could not unsubscribe from topic, message: ", errorMsg.Details())

		default:
			log.Fatal("unknown error occurred, message: ", err)
		}

	}
}

// --- error types definition ---

type ApplicationError interface {
	Details() string
}

type CouldNotConnectToServerError struct {
	Msg string
}

type CouldNotSubscribeToTopicError struct {
	Msg string
}

type CouldNotUnsubscribeFromTopicError struct {
	Msg string
}

func (err *CouldNotConnectToServerError) Details() string {
	res := "could not connect to server, error: " + (*err).Msg
	return res
}

func (err *CouldNotSubscribeToTopicError) Details() string {
	res := "could not subscribe to topic, error: " + (*err).Msg
	return res
}

func (err *CouldNotUnsubscribeFromTopicError) Details() string {
	res := "could not unsubscribe from topic, error: " + (*err).Msg
	return res
}
