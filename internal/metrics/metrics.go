package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"gitlab.com/yakshaving.art/alertsnitch/version"
)

var (
	namespace = "chief_alert_executor"

	bootTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "boot_time_seconds",
		Help:      "unix timestamp of when the service was started",
	})

	buildInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "build_info",
		Help:      "Build information",
	}, []string{"version", "commit", "date"})
)

// Exported metrics
var (
	AlertsReceivedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "alerts",
		Name:      "received_total",
		Help:      "total number of valid alerts received",
	})
	WebhooksReceivedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "webhooks",
		Name:      "received_total",
		Help:      "total number of webhooks posts received",
	})
	InvalidWebhooksTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "webhooks",
		Name:      "invalid_total",
		Help:      "total number of invalid webhooks received",
	})

	AlertsMatchedToCommand = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "alert",
			Name:      "match_total",
			Help:      "total number of alerts matched to a command",
		}, []string{"command"})

	AlertsMissed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "alert",
			Name:      "miss_total",
			Help:      "total number of alerts that did not match to a command",
		})

	CommandsExecuted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "command",
			Name:      "execution_total",
			Help:      "total number of command executions",
		}, []string{"command", "successful"})
)

func init() {
	bootTime.Set(float64(time.Now().Unix()))

	buildInfo.WithLabelValues(version.Version, version.Commit, version.Date).Set(1)

	prometheus.MustRegister(bootTime)
	prometheus.MustRegister(buildInfo)

	prometheus.MustRegister(
		AlertsReceivedTotal,
		AlertsMissed,
		AlertsMatchedToCommand,
		CommandsExecuted,
		InvalidWebhooksTotal,
		WebhooksReceivedTotal,
	)

}
