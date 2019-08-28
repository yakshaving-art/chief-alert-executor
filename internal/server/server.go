package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/matcher"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/metrics"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/templater"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/webhook"
)

// SupportedWebhookVersion is the alert webhook data version that is supported
// by this app
const SupportedWebhookVersion = "4"

// Args are the arguments for building a new server
type Args struct {
	MetricsPath    string
	Address        string
	ConfigFilename string

	Messenger   internal.Messenger
	Concurrency int
}

// Server represents a web server that processes webhooks
type Server struct {
	r *mux.Router

	configFile string
	address    string
	matcher    matcher.Matcher
	templater  templater.Templater

	messenger internal.Messenger

	m       *sync.Mutex
	matches chan matchPayload
}

// New returns a new web server, or fails misserably
func New(args Args) *Server {
	r := mux.NewRouter()

	log.Debugf("Creating new server with args: %#v", args)

	s := &Server{
		r: r,

		configFile: args.ConfigFilename,
		address:    args.Address,

		messenger: args.Messenger,

		m: &sync.Mutex{},

		matches: make(chan matchPayload, args.Concurrency),
	}

	if err := s.LoadConfiguration(); err != nil {
		log.Fatalf("failed to load initial configuration: %s", err)
	}

	r.Handle(args.MetricsPath, promhttp.Handler())
	r.HandleFunc("/webhook", s.webhookPost).Methods("POST")
	r.HandleFunc("/-/health", s.healthyProbe).Methods("GET")
	r.HandleFunc("/-/reload", s.triggerReloadConfiguration).Methods("POST")

	return s
}

// Start starts a new server on the given address
func (s *Server) Start() {
	go s.processMatches()
	log.Println("Starting listener on", s.address)
	log.Fatal(http.ListenAndServe(s.address, s.r))
}

func (s *Server) processMatches() {
	log.Println("starting matches processor")
	for m := range s.matches {
		templater := s.templater.WithTemplate(m.match.Template())
		logger := log.WithField("templater", templater).
			WithField("payload", m.alertGroup).
			WithField("match", m.match)

		event := internal.MatchEvent
		message, err := templater.Expand(internal.MatchEvent, m)

		if err != nil {
			logger.WithField("event", "match").
				Warnf("failed to expand template: %s", err)
		} else {
			err = s.messenger.Send(event, message)
			if err != nil {
				logger.WithField("event", "match").
					WithField("message", message).
					Warnf("failed to send message: %s", err)
			}
		}

		output, err := m.match.Execute()
		payload := struct {
			AlertGroup internal.AlertGroup
			Match      matcher.Match
			Output     string
			Err        error
		}{
			m.alertGroup,
			m.match,
			output,
			err,
		}

		logger = logger.WithField("payload", payload)
		if err == nil {
			event = internal.SuccessEvent
			message, err = templater.Expand(internal.SuccessEvent, payload)
			if err != nil {
				logger.WithField("event", "success").
					Warnf("failed to expand message: %s", err)
			}
		} else {
			event = internal.FailureEvent
			message, err = templater.Expand(internal.FailureEvent, payload)
			if err != nil {
				logger.WithField("event", "failure").
					Warnf("failed to expand message: %s", err)
			}
		}

		if err = s.messenger.Send(event, message); err != nil {
			logger.WithField("message", message).
				Errorf("failed to send message: %s", err)
		}

	}
}

func (s *Server) webhookPost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	metrics.WebhooksReceivedTotal.Inc()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		metrics.InvalidWebhooksTotal.Inc()
		log.Printf("Failed to read payload: %s\n", err)
		http.Error(w, fmt.Sprintf("Failed to read payload: %s", err), http.StatusBadRequest)
		return
	}

	log.Debugln("Received webhook payload", string(body))

	alertGroup, err := webhook.Parse(body)
	if err != nil {
		metrics.InvalidWebhooksTotal.Inc()

		log.Printf("Invalid payload: %s\n", err)
		http.Error(w, fmt.Sprintf("Invalid payload: %s", err), http.StatusBadRequest)
		return
	}

	if alertGroup.Version != SupportedWebhookVersion {
		metrics.InvalidWebhooksTotal.Inc()

		log.Printf("Invalid payload: webhook version %s is not supported\n", alertGroup.Version)
		http.Error(w, fmt.Sprintf("Invalid payload: webhook version %s is not supported",
			alertGroup.Version), http.StatusBadRequest)
		return
	}

	metrics.AlertsReceivedTotal.Inc()

	s.m.Lock()
	defer s.m.Unlock()

	match := s.matcher.Match(*alertGroup)
	if match == nil {
		return
	}

	s.matches <- matchPayload{
		*alertGroup,
		match,
	}

}

func (s *Server) healthyProbe(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) triggerReloadConfiguration(w http.ResponseWriter, r *http.Request) {
	log.Infoln("reloading configuration...")
	if err := s.LoadConfiguration(); err != nil {
		http.Error(w, fmt.Sprintf("failed to reload configuration: %s", err),
			http.StatusInternalServerError)
	} else {
		log.Infoln("configuration reloaded correctly")
	}
}

// LoadConfiguration reloads the configuration file
func (s *Server) LoadConfiguration() error {
	c, err := Load(s.configFile)
	if err != nil {
		return err
	}
	m, err := matcher.New(c)
	if err != nil {
		return err
	}

	s.m.Lock()
	defer s.m.Unlock()

	s.matcher = m
	s.templater = templater.Templater{
		DefaultTemplate: c.DefaultTemplate,
	}

	return nil
}

type matchPayload struct {
	alertGroup internal.AlertGroup
	match      matcher.Match
}
