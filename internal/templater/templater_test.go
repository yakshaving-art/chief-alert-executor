package templater_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/templater"
)

func TestTemplater(t *testing.T) {
	tt := []struct {
		name      string
		templater templater.Templater
		specific  *internal.MessageTemplate
		event     internal.Event
		payload   interface{}
		expected  string
	}{
		{
			"onMatch",
			templater.Templater{
				DefaultTemplate: &internal.MessageTemplate{
					OnMatch: "matched {{ . }}",
				},
			},
			nil,
			internal.MatchEvent,
			"payload",
			"matched payload",
		},
		{
			"OnSuccess",
			templater.Templater{
				DefaultTemplate: &internal.MessageTemplate{
					OnSuccess: "successful {{ . }}",
				},
			},
			nil,
			internal.SuccessEvent,
			"payload",
			"successful payload",
		},
		{
			"OnFailure",
			templater.Templater{
				DefaultTemplate: &internal.MessageTemplate{
					OnFailure: "failure {{ . }}",
				},
			},
			nil,
			internal.FailureEvent,
			"payload",
			"failure payload",
		},
		{
			"OnSpecific",
			templater.Templater{
				DefaultTemplate: &internal.MessageTemplate{
					OnFailure: "failure {{ . }}",
				},
			},
			&internal.MessageTemplate{
				OnFailure: "specific failure {{ . }}",
			},
			internal.FailureEvent,
			"payload",
			"specific failure payload",
		},
		{
			"OnNoTemplate",
			templater.Templater{},
			nil,
			internal.MatchEvent,
			"payload",
			"",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			s, err := tc.templater.WithTemplate(tc.specific).Expand(tc.event, tc.payload)
			a.NoError(err)
			a.Equal(s, tc.expected)
		})
	}

}
