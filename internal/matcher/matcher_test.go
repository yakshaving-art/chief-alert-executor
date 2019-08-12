package matcher_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/matcher"
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
			m, err := matcher.New(tt.cnf)
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
		matches    bool
	}{
		{
			"matcher for alertname label being present matches",
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
			true,
		},
		{
			"matcher for alertname label starting with a matches",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Labels: map[string]string{
							"alertname": "^a.*$",
						},
						Command:   "echo",
						Arguments: []string{"alert!"},
					},
				},
			},
			internal.AlertGroup{
				CommonLabels: map[string]string{
					"alertname": "alert!",
				},
			},
			true,
		},
		{
			"matcher for alertname label starting with a fails to match",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Labels: map[string]string{
							"alertname": "^a.*$",
						},
						Command:   "echo",
						Arguments: []string{"alert!"},
					},
				},
			},
			internal.AlertGroup{
				CommonLabels: map[string]string{
					"alertname": "myalert!",
				},
			},
			false,
		},
		{
			"matcher for alertname label being present fails to match without it",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Labels: map[string]string{
							"name": ".+",
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
			false,
		},
		// Annotations
		{
			"matcher for annotation being present matches",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Annotations: map[string]string{
							"hostname": ".+",
						},
						Command:   "echo",
						Arguments: []string{"alert!"},
					},
				},
			},
			internal.AlertGroup{
				CommonAnnotations: map[string]string{
					"hostname": "myhostname",
				},
			},
			true,
		},
		{
			"matcher for hostname annotation starting with a matches",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Annotations: map[string]string{
							"hostname": "^a.*$",
						},
						Command:   "echo",
						Arguments: []string{"alert!"},
					},
				},
			},
			internal.AlertGroup{
				CommonAnnotations: map[string]string{
					"hostname": "ahostname",
				},
			},
			true,
		},
		{
			"matcher for hostname annotation starting with a fails to match",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Annotations: map[string]string{
							"hostname": "^a.*$",
						},
						Command:   "echo",
						Arguments: []string{"alert!"},
					},
				},
			},
			internal.AlertGroup{
				CommonAnnotations: map[string]string{
					"hostname": "myhostname",
				},
			},
			false,
		},
		{
			"matcher for hostname annotation being present fails to match without it",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Annotations: map[string]string{
							"hostname": "^a.*$",
						},
						Command:   "echo",
						Arguments: []string{"alert!"},
					},
				},
			},
			internal.AlertGroup{
				CommonAnnotations: map[string]string{
					"myhostname": "myhostname",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			m, err := matcher.New(tt.cnf)
			a.NoError(err)
			a.NotNil(m)

			ex := m.Match(tt.alertGroup)

			if tt.matches {
				a.NotNil(ex)
				ex.Execute()
			} else {
				a.Nil(ex)
			}
		})
	}
}

func TestCommandExecutionFails(t *testing.T) {
	tests := []struct {
		name       string
		cnf        internal.Configuration
		alertGroup internal.AlertGroup
	}{
		{
			"empty matcher works, but fails command",
			internal.Configuration{
				Matchers: []internal.MatcherConfiguration{
					{
						Command:   "/bin/false",
						Arguments: []string{},
					},
				},
			},
			internal.AlertGroup{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			m, err := matcher.New(tt.cnf)
			a.NoError(err)
			a.NotNil(m)

			ex := m.Match(tt.alertGroup)

			a.NotNil(ex)
			ex.Execute()
		})
	}
}
