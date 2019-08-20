package messenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal/metrics"

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

func (s slackMessenger) Send(event internal.Event, message string) error {
	if strings.TrimSpace(message) == "" {
		metrics.SlackNotificationsTotal.WithLabelValues(string(event), "empty-message").Inc()
		logrus.Debugf("received empty message to send, ignoring")
		return nil
	}

	b, err := json.Marshal(slackPayload{
		[]slackAttachment{
			slackAttachment{
				event.Color(),
				message,
			},
		},
	})
	if err != nil {
		metrics.SlackNotificationsTotal.WithLabelValues(string(event), "encoding-error").Inc()
		return fmt.Errorf("failed to encode json with the message %s: %s", message, err)
	}

	resp, err := http.Post(s.url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		metrics.SlackNotificationsTotal.WithLabelValues(string(event), "error").Inc()
		return fmt.Errorf("Failed to POST to slack: %s", err)
	}

	logrus.WithField("statusCode", resp.StatusCode).
		WithField("message", message).
		Debugf("posted message to slack")

	metrics.SlackNotificationsTotal.WithLabelValues(string(event), "sent").Inc()
	return nil
}

type slackPayload struct {
	Attachments []slackAttachment `json:"attachments"`
}

type slackAttachment struct {
	Color string `json:"color"`
	Text  string `json:"text"`
}
