package matcher

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

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
	Match(internal.AlertGroup) Executor
}

type oneAlertMatcher struct {
	matcherName string
	labels      map[string]*regexp.Regexp
	annotations map[string]*regexp.Regexp

	cmd  string
	args []string
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

func (m matcherMap) Match(ag internal.AlertGroup) Executor {
	for _, matcher := range m.matchers {
		if matcher.Match(ag) {

			metrics.AlertsMatchedToCommand.
				WithLabelValues(matcher.matcherName).Inc()

			log.WithFields(log.Fields{
				"alertgroup": ag,
				"matcher":    matcher}).
				Debugf("matched alergroup")

			return cmdExecutor{
				cmd:  matcher.cmd,
				args: matcher.args,
			}
		}
	}

	metrics.AlertsMissed.Inc()

	return nil
}

// Executor represents a unit of work
type Executor interface {
	Execute()
}

type cmdExecutor struct {
	matcherName string
	cmd         string
	args        []string
}

func (c cmdExecutor) Execute() {
	cmd := exec.Command(c.cmd, c.args...)
	b, err := cmd.CombinedOutput()

	logger := log.WithField("output", fmt.Sprintf("%s", b)).
		WithField("cmd", c.cmd).
		WithField("matcher", c.matcherName).
		WithField("args", strings.Join(c.args, ","))

	if err != nil {
		logger.WithField("error", err).
			Error("Command failed execution")

		metrics.CommandsExecuted.WithLabelValues(c.matcherName, "false").Inc()

	} else {
		logger.Debug("Command executed correctly")
		metrics.CommandsExecuted.WithLabelValues(c.matcherName, "true").Inc()
	}
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

	return &oneAlertMatcher{
		labels:      labels,
		annotations: annotations,

		matcherName: strings.TrimSpace(mc.Name),
		cmd:         mc.Command,
		args:        mc.Arguments,
	}, nil
}
