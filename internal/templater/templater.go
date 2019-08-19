package templater

import (
	"bytes"
	"fmt"
	"text/template"

	log "github.com/sirupsen/logrus"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal"
)

// Templater is the object used to handle default and different templates
// depending on the specific matcher configuration
type Templater struct {
	DefaultTemplate *internal.MessageTemplate
	template        *internal.MessageTemplate
}

// WithTemplate sets a specific template and returns a new matcher
func (m Templater) WithTemplate(template *internal.MessageTemplate) Templater {
	m.template = template
	return m
}

// Expand send the message by expanding the template to then send the
// actual message
func (m Templater) Expand(forEvent string, payload interface{}) (string, error) {
	var t *internal.MessageTemplate
	// DefaultTemplate has preference over the specific template
	if m.DefaultTemplate != nil {
		t = m.DefaultTemplate
	}
	if m.template != nil {
		t = m.template
	}

	// If there's no template defined, then there's no message to send
	if t == nil {
		log.Debugf("no template defined for event %s and payload %#v", forEvent, payload)
		return "", nil
	}

	tmpl, err := template.New(forEvent).Parse(t.GetMessage(forEvent))
	if err != nil {
		return "", fmt.Errorf("failed to parse %s template %#v: %s", forEvent, t, err)
	}

	b := bytes.NewBufferString("")
	err = tmpl.Execute(b, payload)
	if err != nil {
		return "", fmt.Errorf("failed to expand %s template %s: %s", forEvent, t, err)
	}

	return b.String(), nil
}
