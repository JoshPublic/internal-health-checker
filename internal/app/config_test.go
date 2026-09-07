package app

import "testing"

func TestLoadConfigFromEnvironment(t *testing.T) {
	t.Setenv("SERVICE_NAME", "health-alpha")
	t.Setenv("SERVICE_VERSION", "1.2.3")
	t.Setenv("SERVICE_STATUS", "ok")
	t.Setenv("CHECK_INTERVAL_SECONDS", "15")
	t.Setenv("ALERT_MODE", "log")
	t.Setenv("MONITORED_TARGETS", "api=http://localhost:8081/healthz,cache=http://localhost:6379")

	cfg := LoadConfig()

	if cfg.ServiceName != "health-alpha" {
		t.Fatalf("expected service name health-alpha, got %q", cfg.ServiceName)
	}
	if cfg.ServiceVersion != "1.2.3" {
		t.Fatalf("expected version 1.2.3, got %q", cfg.ServiceVersion)
	}
	if cfg.ServiceStatus != "ok" {
		t.Fatalf("expected status ok, got %q", cfg.ServiceStatus)
	}
	if cfg.CheckIntervalSeconds != 15 {
		t.Fatalf("expected check interval 15, got %d", cfg.CheckIntervalSeconds)
	}
	if cfg.AlertMode != "log" {
		t.Fatalf("expected alert mode log, got %q", cfg.AlertMode)
	}
	if len(cfg.MonitoredTargets) != 2 {
		t.Fatalf("expected 2 monitored targets, got %d", len(cfg.MonitoredTargets))
	}
	if cfg.MonitoredTargets[0].Name != "api" {
		t.Fatalf("expected first target name api, got %q", cfg.MonitoredTargets[0].Name)
	}
	if cfg.MonitoredTargets[0].URL != "http://localhost:8081/healthz" {
		t.Fatalf("expected first target URL to match, got %q", cfg.MonitoredTargets[0].URL)
	}
}

func TestParseMonitoredTargets(t *testing.T) {
	targets := parseMonitoredTargets("alpha=http://localhost:8080/healthz,beta=http://localhost:8081/healthz")
	if len(targets) != 2 {
		t.Fatalf("expected 2 parsed targets, got %d", len(targets))
	}
	if targets[0].Name != "alpha" {
		t.Fatalf("expected first target name alpha, got %q", targets[0].Name)
	}
	if targets[1].URL != "http://localhost:8081/healthz" {
		t.Fatalf("expected second target URL to match, got %q", targets[1].URL)
	}
}
