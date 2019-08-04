package internal

import "time"

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
	Matchers []MatcherConfiguration `yaml:"matchers,omitempty"`
}

// MatcherConfiguration provides configuration to match alerts and map them to a
// command with arguments
type MatcherConfiguration struct {
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
	Command     string            `yaml:"command"`
	Arguments   []string          `yaml:"args"`
}
