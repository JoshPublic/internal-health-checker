package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	cfg    Config
	router *http.ServeMux
}

func (s *Server) StartPeriodicChecks() {
	if s.cfg.CheckIntervalSeconds <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(time.Duration(s.cfg.CheckIntervalSeconds) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			for _, target := range s.cfg.MonitoredTargets {
				status := s.checkTarget(target)
				if status.Status == "down" || status.Status == "degraded" {
					s.alert(status)
				}
			}
		}
	}()
}

type HealthResponse struct {
	ServiceName    string    `json:"serviceName"`
	ServiceVersion string    `json:"serviceVersion"`
	ServiceStatus  string    `json:"serviceStatus"`
	Timestamp      time.Time `json:"timestamp"`
}

type TargetStatus struct {
	Name          string    `json:"name"`
	URL           string    `json:"url"`
	Status        string    `json:"status"`
	HTTPStatus    int       `json:"httpStatus,omitempty"`
	LastCheckedAt time.Time `json:"lastCheckedAt,omitempty"`
	Error         string    `json:"error,omitempty"`
}

type MonitoredTargetResponse struct {
	ServiceName string         `json:"serviceName"`
	Targets     []TargetStatus `json:"targets"`
}

func NewServer(cfg Config) *Server {
	s := &Server{cfg: cfg, router: http.NewServeMux()}
	s.router.HandleFunc("/health", s.handleHealth)
	s.router.HandleFunc("/ready", s.handleHealth)
	s.router.HandleFunc("/monitored-targets", s.handleMonitoredTargets)
	s.router.HandleFunc("/", s.handleRoot)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"status":"not_found"}`))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	payload := HealthResponse{
		ServiceName:    s.cfg.ServiceName,
		ServiceVersion: s.cfg.ServiceVersion,
		ServiceStatus:  s.cfg.ServiceStatus,
		Timestamp:      time.Now().UTC(),
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode health response: %v", err)
	}
}

func (s *Server) handleMonitoredTargets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	statuses := make([]TargetStatus, 0, len(s.cfg.MonitoredTargets))
	for _, target := range s.cfg.MonitoredTargets {
		statuses = append(statuses, s.checkTarget(target))
	}

	response := MonitoredTargetResponse{
		ServiceName: s.cfg.ServiceName,
		Targets:     statuses,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("encode monitored targets response: %v", err)
	}
}

func (s *Server) checkTarget(target Target) TargetStatus {
	status := TargetStatus{
		Name:          target.Name,
		URL:           target.URL,
		Status:        "unknown",
		LastCheckedAt: time.Now().UTC(),
	}

	if target.URL == "" {
		status.Status = "missing"
		status.Error = "empty URL"
		return status
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(target.URL)
	if err != nil {
		status.Status = "down"
		status.Error = err.Error()
		log.Printf("target check failed for %s (%s): %v", target.Name, target.URL, err)
		return status
	}
	defer resp.Body.Close()

	status.HTTPStatus = resp.StatusCode
	status.Status = "up"
	if resp.StatusCode >= 400 {
		status.Status = "degraded"
		status.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	log.Printf("target %s (%s) status=%s http=%d", target.Name, target.URL, status.Status, resp.StatusCode)
	return status
}

func (s *Server) alert(status TargetStatus) {
	switch strings.ToLower(s.cfg.AlertMode) {
	case "", "log":
		log.Printf("ALERT service=%s target=%s status=%s url=%s error=%s", s.cfg.ServiceName, status.Name, status.Status, status.URL, status.Error)
	case "none":
		return
	default:
		log.Printf("ALERT service=%s target=%s status=%s url=%s error=%s (alert mode=%s)", s.cfg.ServiceName, status.Name, status.Status, status.URL, status.Error, s.cfg.AlertMode)
	}
	// TODO: Implement webhook alerting if s.cfg.AlertMode is set to "webhook" and s.cfg.AlertWebhookURL is provided.
}

func (s *Server) LogContext() {
	//TODO: In the future, we can add more context logging here, such as environment variables, build info, current ip, etc.
	if len(s.cfg.MonitoredTargets) == 0 {
		log.Printf("service=%s version=%s status=%s monitoring=0 targets", s.cfg.ServiceName, s.cfg.ServiceVersion, s.cfg.ServiceStatus)
		return
	}
	values := make([]string, 0, len(s.cfg.MonitoredTargets))
	for _, target := range s.cfg.MonitoredTargets {
		values = append(values, target.Name+"="+target.URL)
	}
	log.Printf("service=%s version=%s status=%s monitoring=%d targets=%s", s.cfg.ServiceName, s.cfg.ServiceVersion, s.cfg.ServiceStatus, len(s.cfg.MonitoredTargets), strings.Join(values, ","))
}
