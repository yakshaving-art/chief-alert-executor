package server

import (
	"fmt"
	"io/ioutil"
	"sync"

	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal/matcher"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/metrics"
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
}

// Server represents a web server that processes webhooks
type Server struct {
	r *mux.Router

	configFile string
	address    string
	matcher    matcher.Matcher

	m *sync.Mutex
}

// New returns a new web server, or fails misserably
func New(args Args) *Server {
	r := mux.NewRouter()

	log.Debugf("Creating new server with args: %#v", args)

	s := &Server{
		r: r,

		configFile: args.ConfigFilename,
		address:    args.Address,

		m: &sync.Mutex{},
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
	log.Println("Starting listener on", s.address)
	log.Fatal(http.ListenAndServe(s.address, s.r))
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

	data, err := webhook.Parse(body)
	if err != nil {
		metrics.InvalidWebhooksTotal.Inc()

		log.Printf("Invalid payload: %s\n", err)
		http.Error(w, fmt.Sprintf("Invalid payload: %s", err), http.StatusBadRequest)
		return
	}

	if data.Version != SupportedWebhookVersion {
		metrics.InvalidWebhooksTotal.Inc()

		log.Printf("Invalid payload: webhook version %s is not supported\n", data.Version)
		http.Error(w, fmt.Sprintf("Invalid payload: webhook version %s is not supported", data.Version), http.StatusBadRequest)
		return
	}

	metrics.AlertsReceivedTotal.Inc()

	s.m.Lock()
	defer s.m.Unlock()

	ex := s.matcher.Match(*data)
	if ex != nil {
		ex.Execute()
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
	return nil
}
