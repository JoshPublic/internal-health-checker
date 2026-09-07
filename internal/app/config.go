package app

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Target struct {
	Name          string    `json:"name"`
	URL           string    `json:"url"`
	Status        string    `json:"status"`
	HTTPStatus    int       `json:"httpStatus,omitempty"`
	LastCheckedAt time.Time `json:"lastCheckedAt,omitempty"`
}

type Config struct {
	ServiceName          string
	ServiceVersion       string
	ServiceStatus        string
	Port                 int
	LogLevel             string
	MonitoredTargets     []Target
	CheckIntervalSeconds int
	AlertMode            string
	AlertWebhookURL      string
}

func LoadConfig() Config {
	cfg := Config{
		ServiceName:          getEnv("SERVICE_NAME", "health-checker"),
		ServiceVersion:       getEnv("SERVICE_VERSION", "dev"),
		ServiceStatus:        getEnv("SERVICE_STATUS", "ok"),
		Port:                 getEnvInt("PORT", 8080),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		CheckIntervalSeconds: getEnvInt("CHECK_INTERVAL_SECONDS", 30),
		AlertMode:            strings.ToLower(getEnv("ALERT_MODE", "log")),
		AlertWebhookURL:      getEnv("ALERT_WEBHOOK_URL", ""), // Optional: URL for sending alerts to a webhook
	}

	if cfg.CheckIntervalSeconds <= 0 {
		cfg.CheckIntervalSeconds = 30
	}
	if cfg.AlertMode == "" {
		cfg.AlertMode = "log"
	}

	cfg.MonitoredTargets = loadMonitoredTargets()
	return cfg
}

func loadMonitoredTargets() []Target {
	// First, check if MONITORED_TARGETS environment variable is set
	if targets := parseMonitoredTargets(getEnv("MONITORED_TARGETS", "")); len(targets) > 0 {
		return targets
	}

	// If MONITORED_TARGETS is not set, check for individual target environment variables
	// Check for environment variables in the format TARGET_0_NAME, TARGET_0_URL, TARGET_1_NAME, TARGET_1_URL, etc.
	var items []Target
	for i := 0; i < 25; i++ {
		nameKey := "TARGET_" + strconv.Itoa(i) + "_NAME"
		urlKey := "TARGET_" + strconv.Itoa(i) + "_URL"
		name := strings.TrimSpace(getEnv(nameKey, ""))
		url := strings.TrimSpace(getEnv(urlKey, ""))
		if name == "" && url == "" {
			continue
		}
		if name == "" {
			name = "target-" + strconv.Itoa(i)
		}
		items = append(items, Target{Name: name, URL: url})
	}
	return items
}

func parseMonitoredTargets(value string) []Target {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	// Split the input string by commas to get individual target definitions
	parts := strings.Split(value, ",")
	result := make([]Target, 0, len(parts))
	for _, part := range parts {
		entry := strings.TrimSpace(part)
		if entry == "" {
			continue
		}
		// Split each entry by the first '=' to separate the name and URL
		kv := strings.SplitN(entry, "=", 2)
		if len(kv) != 2 {
			continue
		}
		name := strings.TrimSpace(kv[0])
		url := strings.TrimSpace(kv[1])
		if name == "" || url == "" {
			continue
		}
		// Append the parsed target to the result slice
		result = append(result, Target{Name: name, URL: url})
	}
	return result
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil {
			return parsed
		}
	}
	return fallback
}
