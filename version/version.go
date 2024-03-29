package version

import "fmt"

// Name is the application name
const Name = "alertsnitch"

// Version is the application Version
var Version = "unspecified"

// Date is the built date and time
var Date = "unspecified"

// Commit is the commit in which the package is based
var Commit = "unspecified"

// GetVersion returns the version as a string
func GetVersion() string {
	return fmt.Sprintf("%s Version: %s Commit: %s Date: %s", Name, Version, Commit, Date)
}
