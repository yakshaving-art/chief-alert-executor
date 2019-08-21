package internal

import (
	"time"

	"github.com/sirupsen/logrus"
)

// AlertGroup is the data read from a webhook call
type AlertGroup struct {
	Version  string `json:"version"`
	GroupKey string `json:"groupKey"`

	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   Alerts `json:"alerts"`

	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`

	ExternalURL string `json:"externalURL"`
}

// Alerts is a slice of Alert
type Alerts []Alert

// Alert holds one alert for notification templates.
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
}

// Configuration represents the configuration of the alert matchers
type Configuration struct {
	Matchers        []MatcherConfiguration `yaml:"matchers,omitempty"`
	DefaultTemplate *MessageTemplate       `yaml:"default_template,omitempty"`
}

// MatcherConfiguration provides configuration to match alerts and map them to a
// command with arguments
type MatcherConfiguration struct {
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
	Command     string            `yaml:"command"`
	Arguments   []string          `yaml:"args"`
	Template    *MessageTemplate  `yaml:"template,omitempty"`
}

// Messenger represents an object capable of sending a message to somewhere
type Messenger interface {
	Send(Event, string) error
}

// MessageTemplate is the message to send when the match is successful
type MessageTemplate struct {
	OnMatch   string `yaml:"on_match"`
	OnSuccess string `yaml:"on_success"`
	OnFailure string `yaml:"on_failure"`
}

// GetMessage returns the template according to the event type
func (m MessageTemplate) GetMessage(event Event) string {
	switch event {
	case MatchEvent:
		return m.OnMatch

	case SuccessEvent:
		return m.OnSuccess

	case FailureEvent:
		return m.OnFailure

	}
	logrus.Panicf("Invalid event %s", event)
	return ""
}

// Constants used to signal the different kind of events
const (
	MatchEvent   = Event("match")
	SuccessEvent = Event("success")
	FailureEvent = Event("failure")
)

// Event is an extension of a string used to map the different colors of the events
type Event string

// Color returns the color given the kind of event it is
func (e Event) Color() string {
	switch e {
	case SuccessEvent:
		return "good" // Green
	case FailureEvent:
		return "danger" // Red
	}
	return "warning" // Matchevent will be yellow
}
