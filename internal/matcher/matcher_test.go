package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal"
)

func TestNewMatchers(t *testing.T) {
	tests := []struct {
		name  string
		cnf   internal.Configuration
		orErr string
	}{
		{
			"empty works",
			internal.Configuration{},
			"",
		},
		{
			"one label and one configuration",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Labels: map[string]string{
							"alertname": "bla",
						},
						Annotations: map[string]string{
							"myannotation": "scratch",
						},
					},
				},
			},
			"",
		},
		{
			"invalid regex for labels fails",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Labels: map[string]string{
							"invalid_label_regex": "[",
						},
					},
				},
			},
			"Failed to compile regex for label invalid_label_regex ([): " +
				"error parsing regexp: missing closing ]: `[`",
		},
		{
			"invalid regex for annotations fails",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Annotations: map[string]string{
							"invalid_annotation_regex": "[",
						},
					},
				},
			},
			"Failed to compile regex for label invalid_annotation_regex ([): " +
				"error parsing regexp: missing closing ]: `[`",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			m, err := New(tt.cnf)
			if tt.orErr == "" {
				a.NoError(err)
				a.NotNil(m)
			} else {
				a.Errorf(err, tt.orErr)
			}
		})
	}
}

func TestMatching(t *testing.T) {
	tests := []struct {
		name       string
		cnf        internal.Configuration
		alertGroup internal.AlertGroup
		orErr      string
	}{
		{
			"matcher for alertname being present",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Labels: map[string]string{
							"alertname": ".+",
						},
						Command:   "echo",
						Arguments: []string{"alert!"},
					},
				},
			},
			internal.AlertGroup{
				CommonLabels: map[string]string{
					"alertname": "myalert",
				},
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			m, err := New(tt.cnf)
			a.NoError(err)
			a.NotNil(m)

			ex := m.Match(tt.alertGroup)

			a.NotNil(ex)

			a.True(ex.Execute())
		})
	}
}
