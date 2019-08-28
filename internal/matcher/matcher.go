package matcher

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/metrics"
)

// New creates a new Matcher with the provided configuration.
//
// May return an error if we fail to load the configuration
func New(cnf internal.Configuration) (Matcher, error) {
	am := make([]*oneAlertMatcher, 0)
	for _, m := range cnf.Matchers {
		matcher, err := newAlertMatcher(m)
		if err != nil {
			return nil, err
		}
		am = append(am, matcher)
	}

	return matcherMap{
		matchers: am,
	}, nil
}

// Matcher is the interface of the whatever loads the configuration and then is
// used to match an alert to an executor
type Matcher interface {
	Match(internal.AlertGroup) Match
}

type oneAlertMatcher struct {
	matcherName string
	labels      map[string]*regexp.Regexp
	annotations map[string]*regexp.Regexp

<<<<<<< Updated upstream
	template *internal.MessageTemplate
	cmd      string
	args     []string
=======
	cmd  string
	args []string

	timeout int
>>>>>>> Stashed changes
}

func (m oneAlertMatcher) Match(ag internal.AlertGroup) bool {
	for name, regex := range m.annotations {
		value, ok := ag.CommonAnnotations[name]
		if !ok {
			log.WithFields(log.Fields{
				"alertgroup": ag,
				"annotation": name,
				"matcher":    m.matcherName,
			}).Debugf("alert does not contain expected annotation")
			return false
		}
		if !regex.MatchString(value) {
			log.WithField("alertgroup", ag).
				WithField("annotation", name).
				WithField("value", value).
				WithField("matcher", m.matcherName).
				Debugf("alert does not match expected regex for annotation")
			return false
		}
	}

	for name, regex := range m.labels {
		value, ok := ag.CommonLabels[name]
		if !ok {
			log.WithFields(log.Fields{
				"alertgroup": ag,
				"label":      name,
				"matcher":    m.matcherName,
			}).Debugf("alert does not contain expected label")
			return false
		}
		if !regex.MatchString(value) {
			log.WithField("alertgroup", ag).
				WithField("label", name).
				WithField("value", value).
				WithField("matcher", m.matcherName).
				Debugf("alert does not match expected regex for label")
			return false
		}
	}

	log.WithField("alertgroup", ag).
		WithField("matcher", m).
		Debugf("alert matched")

	return true
}

type matcherMap struct {
	matchers []*oneAlertMatcher
}

func (m matcherMap) Match(ag internal.AlertGroup) Match {
	for _, matcher := range m.matchers {
		if matcher.Match(ag) {

			metrics.AlertsMatchedToCommand.
				WithLabelValues(matcher.matcherName).Inc()

			log.WithFields(log.Fields{
				"alertgroup": ag,
				"matcher":    matcher}).
				Debugf("matched alergroup")

			return cmdExecutor{
				template:    matcher.template,
				matcherName: matcher.matcherName,
				cmd:         matcher.cmd,
				args:        matcher.args,
				timeout:     time.Duration(matcher.timeout) * time.Second,
			}
		}
	}

	metrics.AlertsMissed.Inc()

	return nil
}

// Match represents a unit of work
type Match interface {
	Name() string
	Template() *internal.MessageTemplate
	Execute() (string, error)
}

type cmdExecutor struct {
	template    *internal.MessageTemplate
	matcherName string
	cmd         string
	args        []string
	timeout     time.Duration
}

<<<<<<< Updated upstream
func (c cmdExecutor) Name() string {
	return c.matcherName
}

func (c cmdExecutor) Template() *internal.MessageTemplate {
	return c.template
}

func (c cmdExecutor) Execute() (string, error) {
=======
func (c cmdExecutor) Execute() {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

>>>>>>> Stashed changes
	startTime := time.Now()
	cmd := exec.CommandContext(ctx, c.cmd, c.args...)
	b, err := cmd.CombinedOutput()
	executionTime := time.Now().Sub(startTime)

	output := fmt.Sprintf("%s", b)
	logger := log.WithField("output", output).
		WithField("cmd", c.cmd).
		WithField("matcher", c.matcherName).
		WithField("args", strings.Join(c.args, ","))

	if err != nil {
		logger.WithField("error", err).
			Error("Command failed execution")

		metrics.CommandsExecuted.WithLabelValues(c.matcherName, "false").Inc()
		metrics.CommandExecutionSeconds.WithLabelValues(c.matcherName, "false").Observe(executionTime.Seconds())
		return output, err
	}

	logger.Debug("Command executed correctly")
	metrics.CommandsExecuted.WithLabelValues(c.matcherName, "true").Inc()
	metrics.CommandExecutionSeconds.WithLabelValues(c.matcherName, "true").Observe(executionTime.Seconds())
	return output, nil
}

func newAlertMatcher(mc internal.MatcherConfiguration) (*oneAlertMatcher, error) {

	if strings.TrimSpace(mc.Name) == "" {
		return nil, fmt.Errorf("Metric name can't be empty in %#v", mc)
	}
	if strings.TrimSpace(mc.Command) == "" {
		return nil, fmt.Errorf("Command can't be empty in %#v", mc)
	}

	labels := make(map[string]*regexp.Regexp)
	for l, r := range mc.Labels {
		reg, err := regexp.Compile(r)
		if err != nil {
			return nil, fmt.Errorf("Failed to compile regex for label %s (%s): %s", l, r, err)
		}
		labels[l] = reg
	}

	annotations := make(map[string]*regexp.Regexp)
	for a, r := range mc.Annotations {
		reg, err := regexp.Compile(r)
		if err != nil {
			return nil, fmt.Errorf("Failed to compile regex for annotation %s (%s): %s", a, r, err)
		}
		annotations[a] = reg
	}
	timeout := mc.Timeout
	if timeout == 0 {
		timeout = 30 // By default, 30 seconds of command execution timeout
	}

	return &oneAlertMatcher{
		labels:      labels,
		annotations: annotations,

		matcherName: strings.TrimSpace(mc.Name),
		template:    mc.Template,
		cmd:         mc.Command,
		args:        mc.Arguments,
		timeout:     timeout,
	}, nil
}
