package messenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal"
)

type slackMessenger struct {
	url string
}

// Slack returns a new Slack Messenger
func Slack(url string) internal.Messenger {
	return slackMessenger{
		url: url,
	}
}

func (s slackMessenger) Send(message string) error {
	if strings.TrimSpace(message) == "" {
		logrus.Debugf("received empty message to send, ignoring")
		return nil
	}

	b, err := json.Marshal(slackMessage{message})
	if err != nil {
		return fmt.Errorf("failed to encode json with the message %s: %s", message, err)
	}

	resp, err := http.Post(s.url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("Failed to POST to slack: %s", err)
	}

	logrus.WithField("statusCode", resp.StatusCode).
		WithField("message", message).
		Debugf("posted message to slack")

	return nil
}

type slackMessage struct {
	Text string `json:"text"`
}
