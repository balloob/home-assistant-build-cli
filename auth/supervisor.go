package auth

import "os"

const (
	// SupervisorTokenEnv is the environment variable set by the Home Assistant Supervisor
	// when running inside an add-on (app).
	SupervisorTokenEnv = "SUPERVISOR_TOKEN"

	// SupervisorURL is the base URL for accessing the Home Assistant Core API
	// through the Supervisor proxy when running inside an add-on.
	SupervisorURL = "http://supervisor/core"
)

// IsSupervisorEnvironment returns true if running inside a Home Assistant add-on
// by checking for the SUPERVISOR_TOKEN environment variable.
func IsSupervisorEnvironment() bool {
	return os.Getenv(SupervisorTokenEnv) != ""
}

// GetSupervisorToken returns the supervisor token if available.
func GetSupervisorToken() string {
	return os.Getenv(SupervisorTokenEnv)
}
