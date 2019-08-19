package main

import (
	"flag"
	"os"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal/messenger"

	"github.com/sirupsen/logrus"

	_ "gitlab.com/yakshaving.art/chief-alert-executor/internal/metrics"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/server"
)

func main() {

	address := flag.String("address", ":9099", "Address to listen to")
	metricsPath := flag.String("metrics", "/metrics", "path in which to listen for metrics")
	configFilename := flag.String("config", "config.yml", "configuration filename")
	debug := flag.Bool("debug", false, "enable debug mode")

	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	m := messenger.Noop()

	slackURL := os.Getenv("SLACK_URL")
	if slackURL != "" {
		m = messenger.Slack(slackURL)
	}

	s := server.New(server.Args{
		Address:        *address,
		MetricsPath:    *metricsPath,
		ConfigFilename: *configFilename,
		Messenger:      m,
	})

	s.Start()
}
