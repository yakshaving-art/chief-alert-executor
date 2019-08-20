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

func (noopMessenger) Send(event internal.Event, message string) error {
	logrus.WithField("message", message).WithField("event", event).Debugf("Noop message.")
	return nil
}
