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

// Server represents a web server that processes webhooks
type Server struct {
	r *mux.Router

	configFile string
	matcher    matcher.Matcher

	m *sync.Mutex
}

// New returns a new web server, or fails misserably
func New(cnfPath string) *Server {
	r := mux.NewRouter()

	s := Server{
		r: r,
		m: &sync.Mutex{},
	}

	if err := s.loadConfiguration(); err != nil {
		log.Fatalf("failed to load initial configuration: %s", err)
	}

	r.HandleFunc("/webhook", s.webhookPost).Methods("POST")
	r.HandleFunc("/-/health", s.healthyProbe).Methods("GET")
	r.HandleFunc("/-/reload", s.triggerReloadConfiguration).Methods("POST")
	r.Handle("/metrics", promhttp.Handler())

	return &s
}

// Start starts a new server on the given address
func (s Server) Start(address string) {
	log.Println("Starting listener on", address)
	log.Fatal(http.ListenAndServe(address, s.r))
}

func (s Server) webhookPost(w http.ResponseWriter, r *http.Request) {
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

func (s Server) healthyProbe(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) triggerReloadConfiguration(w http.ResponseWriter, r *http.Request) {
	if err := s.loadConfiguration(); err != nil {
		http.Error(w, fmt.Sprintf("failed to reload configuration: %s", err),
			http.StatusInternalServerError)
	}
}

func (s *Server) loadConfiguration() error {
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
