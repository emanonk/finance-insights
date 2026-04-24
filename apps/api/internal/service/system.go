package service

// System carries application-wide operational logic (health, version, etc.).
type System struct{}

// NewSystem constructs a System service.
func NewSystem() *System {
	return &System{}
}

// HealthStatus is the operational health payload for clients.
type HealthStatus struct {
	Status string `json:"status"`
}

// Health reports whether the process is live.
func (s *System) Health() HealthStatus {
	return HealthStatus{Status: "ok"}
}
