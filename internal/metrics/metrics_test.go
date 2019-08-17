package metrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal/metrics"
)

func TestMetricsAreRegistered(t *testing.T) {
	a := assert.New(t)
	a.True(prometheus.DefaultRegisterer.Unregister(metrics.AlertsReceivedTotal),
		"alerts received total")
	a.True(prometheus.DefaultRegisterer.Unregister(metrics.AlertsMissed),
		"alerts missed")
	a.True(prometheus.DefaultRegisterer.Unregister(metrics.AlertsMatchedToCommand),
		"alerts matched to a command")
	a.True(prometheus.DefaultRegisterer.Unregister(metrics.CommandsExecuted),
		"commands executed")
	a.True(prometheus.DefaultRegisterer.Unregister(metrics.InvalidWebhooksTotal),
		"invalid webhooks total")
	a.True(prometheus.DefaultRegisterer.Unregister(metrics.WebhooksReceivedTotal),
		"webhooks received total")
}
