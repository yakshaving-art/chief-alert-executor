package messenger

import (
	"github.com/sirupsen/logrus"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal"
)

type noopMessenger struct{}

// Noop returns a null messenger that does nothing
func Noop() internal.Messenger {
	return noopMessenger{}
}

func (noopMessenger) Send(message string) error {
	logrus.WithField("message", message).Debugf("Noop message.")
	return nil
}
