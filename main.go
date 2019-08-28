package main

import (
	"flag"
	"os"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal/messenger"

	"github.com/sirupsen/logrus"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal/metrics"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/server"
)

func main() {

	address := flag.String("address", ":9099", "Address to listen to")
	metricsPath := flag.String("metrics", "/metrics", "path in which to listen for metrics")
	configFilename := flag.String("config", "config.yml", "configuration filename")
	debug := flag.Bool("debug", false, "enable debug mode")
	concurrency := flag.Int("concurrency", 10, "how many commands can be executed concurrently")

	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	m := messenger.Noop()

	slackURL := os.Getenv("SLACK_URL")
	if slackURL != "" {
		logrus.Info("Slack notifications enabled")
		metrics.SlackUp.Set(1)
		m = messenger.Slack(slackURL)
	}

	s := server.New(server.Args{
		Address:        *address,
		MetricsPath:    *metricsPath,
		ConfigFilename: *configFilename,
		Concurrency:    *concurrency,
		Messenger:      m,
	})

	s.Start()
}
