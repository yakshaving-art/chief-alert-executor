package metrics

import (
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

	LastConfigReloadTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_configuration_reload_seconds",
			Help:      "unix timestamp of when the configuration was last reloaded",
		},
	)
	LastConfigReloadSuccessful = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_configuration_reload_successful",
			Help:      "wether or not the last configuration was successfully reloaded",
		},
	)
	SlackUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "slack_up",
		Help:      "Wether or not if slack notifications are enabled",
	})
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
		}, []string{"matcher"})

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
		}, []string{"matcher", "successful"})

	CommandExecutionSeconds = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  namespace,
		Subsystem:  "command",
		Name:       "execution_seconds",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		Help:       "command execution seconds summary",
	}, []string{"matcher", "successful"})
	SlackNotificationsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "slack",
		Name:      "notifications_total",
	}, []string{"kind", "status"})
)

func init() {
	bootTime.SetToCurrentTime()
	buildInfo.WithLabelValues(version.Version, version.Commit, version.Date).Set(1)
	SlackUp.Set(0)

	prometheus.MustRegister(bootTime,
		buildInfo,
		LastConfigReloadTime,
		LastConfigReloadSuccessful,
		SlackUp,
	)

	prometheus.MustRegister(
		AlertsReceivedTotal,
		AlertsMissed,
		AlertsMatchedToCommand,
		CommandsExecuted,
		CommandExecutionSeconds,
		InvalidWebhooksTotal,
		WebhooksReceivedTotal,
	)

}
