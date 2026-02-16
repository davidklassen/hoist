package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGetStatusAllServices(t *testing.T) {
	cfg := testConfig()
	deploys := map[string]deploy{
		"backend:staging":     {Service: "backend", Env: "staging", Tag: "main-abc1234-20250101000000", Uptime: 3 * time.Hour},
		"backend:production":  {Service: "backend", Env: "production", Tag: "main-def5678-20241231000000", Uptime: 48 * time.Hour},
		"frontend:staging":    {Service: "frontend", Env: "staging", Tag: "main-abc1234-20250101000000", Uptime: 1 * time.Hour},
		"frontend:production": {Service: "frontend", Env: "production", Tag: "main-def5678-20241231000000", Uptime: 24 * time.Hour},
	}
	p, _ := testProviders(nil, deploys)

	rows, err := getStatus(context.Background(), cfg, p, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}

	// Rows should be sorted by service name, then env
	expected := []struct {
		service, env string
	}{
		{"backend", "production"},
		{"backend", "staging"},
		{"frontend", "production"},
		{"frontend", "staging"},
	}
	for i, e := range expected {
		if rows[i].Service != e.service || rows[i].Env != e.env {
			t.Errorf("row %d: expected %s/%s, got %s/%s", i, e.service, e.env, rows[i].Service, rows[i].Env)
		}
	}
}

func TestGetStatusFilteredByEnv(t *testing.T) {
	cfg := testConfig()
	deploys := map[string]deploy{
		"backend:staging":  {Service: "backend", Env: "staging", Tag: "tag1", Uptime: time.Hour},
		"frontend:staging": {Service: "frontend", Env: "staging", Tag: "tag2", Uptime: time.Hour},
	}
	p, _ := testProviders(nil, deploys)

	rows, err := getStatus(context.Background(), cfg, p, "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	for _, r := range rows {
		if r.Env != "staging" {
			t.Errorf("expected env staging, got %s", r.Env)
		}
	}
}

func TestGetStatusHealthValues(t *testing.T) {
	cfg := testConfig()
	deploys := map[string]deploy{
		"backend:staging":  {Service: "backend", Env: "staging", Tag: "tag1", Uptime: time.Hour},
		"frontend:staging": {Service: "frontend", Env: "staging", Tag: "tag2", Uptime: time.Hour},
	}
	p, _ := testProviders(nil, deploys)

	rows, err := getStatus(context.Background(), cfg, p, "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var backendHealth, frontendHealth string
	for _, r := range rows {
		if r.Service == "backend" {
			backendHealth = r.Health
		}
		if r.Service == "frontend" {
			frontendHealth = r.Health
		}
	}

	if backendHealth != "healthy" {
		t.Errorf("expected backend health 'healthy', got %q", backendHealth)
	}
	if frontendHealth != "-" {
		t.Errorf("expected frontend health '-', got %q", frontendHealth)
	}
}

func TestGetStatusMissingDeploy(t *testing.T) {
	cfg := testConfig()
	// Only backend:staging has a deploy; frontend:staging returns zero deploy (not deployed yet)
	deploys := map[string]deploy{
		"backend:staging": {Service: "backend", Env: "staging", Tag: "tag1", Uptime: time.Hour},
	}
	p, _ := testProviders(nil, deploys)

	rows, err := getStatus(context.Background(), cfg, p, "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	for _, r := range rows {
		if r.Service == "frontend" && r.Tag != "" {
			t.Errorf("expected empty tag for frontend with missing deploy, got %q", r.Tag)
		}
	}
}

func TestGetStatusProviderError(t *testing.T) {
	cfg := testConfig()
	mh := &mockHistoryProvider{
		currentErrors: map[string]error{
			"backend:staging": fmt.Errorf("SSH connection refused"),
		},
	}
	p := providers{
		history: map[string]historyProvider{
			"server": mh,
			"static": mh,
		},
	}

	_, err := getStatus(context.Background(), cfg, p, "staging")
	if err == nil {
		t.Fatal("expected error from history provider")
	}
	if !contains(err.Error(), "SSH connection refused") {
		t.Errorf("expected underlying error, got: %v", err)
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{40 * time.Minute, "40m"},
		{2 * time.Hour, "2h"},
		{24 * time.Hour, "1d"},
		{72 * time.Hour, "3d"},
		{30 * time.Minute, "30m"},
		{0, "0m"},
	}
	for _, tt := range tests {
		got := formatUptime(tt.d)
		if got != tt.want {
			t.Errorf("formatUptime(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFormatStatusTable(t *testing.T) {
	rows := []statusRow{
		{Service: "backend", Env: "staging", Tag: "main-abc1234-20250101000000", Uptime: 3 * time.Hour, Health: "healthy"},
		{Service: "frontend", Env: "staging", Tag: "main-abc1234-20250101000000", Uptime: 1 * time.Hour, Health: "-"},
	}
	output := formatStatusTable(rows)
	if output == "" {
		t.Fatal("expected non-empty output")
	}
	if !contains(output, "SERVICE") || !contains(output, "TAG") || !contains(output, "UPTIME") || !contains(output, "HEALTH") {
		t.Error("expected table headers")
	}
	if !contains(output, "backend-staging") {
		t.Error("expected backend-staging label")
	}
	if !contains(output, "frontend-staging") {
		t.Error("expected frontend-staging label")
	}
}

func TestFormatStatusTableEmpty(t *testing.T) {
	output := formatStatusTable(nil)
	if output != "No services found.\n" {
		t.Errorf("expected 'No services found.' message, got %q", output)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
