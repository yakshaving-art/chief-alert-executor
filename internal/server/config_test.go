package server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/server"
)

func TestLoadingConfig(t *testing.T) {
	tt := []struct {
		name     string
		filename string
		err      string
		config   internal.Configuration
	}{
		{
			name:     "empty config works",
			filename: "fixtures/empty-config.yaml",
			err:      "",
			config:   internal.Configuration{},
		},
		{
			name:     "valid config works",
			filename: "fixtures/valid-config.yaml",
			err:      "",
			config: internal.Configuration{Matchers: []internal.MatcherConfiguration{
				internal.MatcherConfiguration{
					Labels: map[string]string{
						"exact":   "^exact-value$",
						"somekey": "somevalue.*"},
					Annotations: map[string]string{
						"alertname": "SomeAlert.*"},
					Command: "echo",
					Arguments: []string{
						"this", "alert", "is", "silly",
					},
				}}},
		},
		{
			name:     "non-existing file fails",
			filename: "fixtures/non-existing-config.yaml",
			err:      "failed to read configuration file fixtures/non-existing-config.yaml: open fixtures/non-existing-config.yaml: no such file or directory",
			config:   internal.Configuration{},
		},
		{
			name:     "invalid-config fails",
			filename: "fixtures/invalid-config.yaml",
			err: "failed to parse yaml configuration file fixtures/invalid-config.yaml: yaml: unmarshal errors:\n" +
				"  line 3: field commands not found in type internal.MatcherConfiguration",
			config: internal.Configuration{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)

			c, err := server.Load(tc.filename)
			if tc.err != "" {
				a.EqualError(err, tc.err)
			} else {
				a.NoError(err)
				a.Equal(tc.config, c)
			}
		})
	}
}
