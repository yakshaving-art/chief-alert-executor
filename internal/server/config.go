package server

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"gitlab.com/yakshaving.art/chief-alert-executor/internal"
	"gitlab.com/yakshaving.art/chief-alert-executor/internal/metrics"
)

// Load loads the configuration from the provided file
func Load(filename string) (internal.Configuration, error) {
	c := internal.Configuration{}

	in, err := ioutil.ReadFile(filename)
	if err != nil {
		metrics.LastConfigReloadSuccessful.Set(0)
		return c, fmt.Errorf("failed to read configuration file %s: %s", filename, err)
	}

	err = yaml.UnmarshalStrict(in, &c)
	if err != nil {
		metrics.LastConfigReloadSuccessful.Set(0)
		return c, fmt.Errorf("failed to parse yaml configuration file %s: %s", filename, err)
	}

	metrics.LastConfigReloadSuccessful.Set(1)
	metrics.LastConfigReloadTime.SetToCurrentTime()
	return c, nil
}
