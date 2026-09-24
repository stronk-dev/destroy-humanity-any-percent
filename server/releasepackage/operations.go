package releasepackage

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v2"
)

var blockingAlerts = []string{
	"CloudClickerPublicEndpointOrReadinessUnavailable",
	"CloudClickerBackupMissingOrFailed",
	"CloudClickerPostgresUnreachable",
	"CloudClickerStoragePressure",
	"CloudClickerGameserverRestartLoop",
	"CloudClickerCleanupJobFailed",
	"CloudClickerDeadLetterGrowth",
}

func ValidateOperationsProfile(root string) error {
	files := map[string][]byte{}
	for _, name := range []string{"prometheus.yml", "cloud-clicker-alerts.yml", "cloud-clicker-alerts.test.yml", "alertmanager.example.yml", "journald.template.conf", "cloud-clicker-observe.service", "cloud-clicker-observe.timer", "operations.env.example"} {
		data, err := os.ReadFile(filepath.Join(root, "operations", name))
		if err != nil || len(bytes.TrimSpace(data)) == 0 {
			return fmt.Errorf("%w: operations file %s", ErrInvalidContent, name)
		}
		files[name] = data
	}
	for _, name := range []string{"prometheus.yml", "cloud-clicker-alerts.yml", "cloud-clicker-alerts.test.yml", "alertmanager.example.yml"} {
		var value any
		if yaml.Unmarshal(files[name], &value) != nil {
			return fmt.Errorf("%w: operations YAML %s", ErrInvalidContent, name)
		}
	}
	prometheus := string(files["prometheus.yml"])
	for target, count := range map[string]int{"gameserver:8080": 1, "caddy:2020": 1, "node-exporter:9100": 1, "alertmanager:9093": 2, "localhost:9090": 1} {
		if strings.Count(prometheus, target) != count {
			return fmt.Errorf("%w: Prometheus target %s", ErrInvalidContent, target)
		}
	}
	if strings.Contains(prometheus, "0.0.0.0") || strings.Contains(prometheus, "host.docker.internal") {
		return fmt.Errorf("%w: non-private Prometheus target", ErrInvalidContent)
	}
	rules, tests := string(files["cloud-clicker-alerts.yml"]), string(files["cloud-clicker-alerts.test.yml"])
	var testProfile struct {
		Tests []struct {
			AlertRules []struct {
				AlertName string           `yaml:"alertname"`
				Expected  []map[string]any `yaml:"exp_alerts"`
			} `yaml:"alert_rule_test"`
		} `yaml:"tests"`
	}
	if yaml.Unmarshal(files["cloud-clicker-alerts.test.yml"], &testProfile) != nil {
		return fmt.Errorf("%w: alert test profile", ErrInvalidContent)
	}
	for _, alert := range blockingAlerts {
		if strings.Count(rules, "alert: "+alert) != 1 || strings.Count(tests, "alertname: "+alert) < 2 {
			return fmt.Errorf("%w: incomplete alert evidence %s", ErrInvalidContent, alert)
		}
		fired, resolved := false, false
		for _, test := range testProfile.Tests {
			for _, check := range test.AlertRules {
				if check.AlertName != alert {
					continue
				}
				if len(check.Expected) == 0 {
					resolved = true
				} else {
					fired = true
				}
			}
		}
		if !fired || !resolved {
			return fmt.Errorf("%w: missing fired/resolved alert evidence %s", ErrInvalidContent, alert)
		}
	}
	for _, forbidden := range []string{"account_id", "founder_id", "stream_id", "recovery_code", "request_payload", "client_ip", "remote_ip"} {
		if strings.Contains(strings.ToLower(rules), forbidden) || strings.Contains(strings.ToLower(prometheus), forbidden) {
			return fmt.Errorf("%w: private/unbounded operations label %s", ErrInvalidContent, forbidden)
		}
	}
	journal := string(files["journald.template.conf"])
	if strings.Count(journal, "MaxRetentionSec=@@MAX_RETENTION_SEC@@") != 1 || strings.Count(journal, "SystemMaxUse=@@SYSTEM_MAX_USE@@") != 1 || strings.Count(journal, "Storage=persistent") != 1 {
		return fmt.Errorf("%w: journald policy template", ErrInvalidContent)
	}
	service, timer := string(files["cloud-clicker-observe.service"]), string(files["cloud-clicker-observe.timer"])
	if !strings.Contains(service, "deployment-operations host-observe") || strings.Count(service, "EnvironmentFile=/etc/cloud-clicker/operations.env") != 1 ||
		!strings.Contains(service, "--journal-budget-bytes-file=${CLOUD_CLICKER_JOURNAL_BUDGET_FILE}") ||
		strings.Count(timer, "OnUnitActiveSec=1m") != 1 || strings.Count(timer, "Persistent=true") != 1 {
		return fmt.Errorf("%w: host observation schedule", ErrInvalidContent)
	}
	environment := string(files["operations.env.example"])
	for _, name := range []string{"CLOUD_CLICKER_OPERATIONS_METRICS", "CLOUD_CLICKER_POSTGRES_DATA", "CLOUD_CLICKER_BACKUP_TARGET", "CLOUD_CLICKER_JOURNAL_PATH", "CLOUD_CLICKER_JOURNAL_BUDGET_FILE"} {
		if strings.Count(environment, name+"=") != 1 || strings.Count(service, "${"+name+"}") != 1 {
			return fmt.Errorf("%w: host observation variable %s", ErrInvalidContent, name)
		}
	}
	alertmanager := string(files["alertmanager.example.yml"])
	if strings.Count(alertmanager, "send_resolved: true") != 1 || strings.Contains(strings.ToLower(alertmanager), "password:") || strings.Contains(strings.ToLower(alertmanager), "token:") {
		return fmt.Errorf("%w: Alertmanager receiver template", ErrInvalidContent)
	}
	return nil
}
