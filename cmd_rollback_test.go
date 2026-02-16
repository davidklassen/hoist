package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestResolveRollbackTargetsFilteredByService(t *testing.T) {
	cfg := testConfig()
	mh := &mockHistoryProvider{
		previousDeploys: map[string]deploy{
			"backend:staging":  {Service: "backend", Env: "staging", Tag: "prev-backend-tag"},
			"frontend:staging": {Service: "frontend", Env: "staging", Tag: "prev-frontend-tag"},
		},
	}
	p := providers{
		history: map[string]historyProvider{
			"server": mh,
			"static": mh,
		},
	}

	var buf bytes.Buffer
	res, err := resolveRollbackTargets(context.Background(), cfg, p, []string{"backend"}, "staging", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.targets) != 1 {
		t.Fatalf("expected 1 target, got %d: %v", len(res.targets), res.targets)
	}
	if res.targets[0] != "backend" {
		t.Errorf("expected backend, got %s", res.targets[0])
	}
	if res.tags["backend"] != "prev-backend-tag" {
		t.Errorf("expected prev-backend-tag, got %s", res.tags["backend"])
	}
	if _, ok := res.tags["frontend"]; ok {
		t.Error("frontend should not be in rollback tags")
	}
}

func TestResolveRollbackTargetsSkipNoPrev(t *testing.T) {
	cfg := testConfig()
	mh := &mockHistoryProvider{
		previousDeploys: map[string]deploy{
			"backend:staging": {Service: "backend", Env: "staging", Tag: "prev-backend-tag"},
			// frontend:staging not set → returns deploy{}, nil → Tag is empty → skipped
		},
	}
	p := providers{
		history: map[string]historyProvider{
			"server": mh,
			"static": mh,
		},
	}

	var buf bytes.Buffer
	res, err := resolveRollbackTargets(context.Background(), cfg, p, nil, "staging", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.targets) != 1 {
		t.Fatalf("expected 1 target, got %d: %v", len(res.targets), res.targets)
	}
	if res.targets[0] != "backend" {
		t.Errorf("expected backend, got %s", res.targets[0])
	}
	if len(res.skipped) != 1 || res.skipped[0] != "frontend" {
		t.Errorf("expected frontend skipped, got %v", res.skipped)
	}
	if !strings.Contains(buf.String(), "skipping frontend: no previous deploy") {
		t.Errorf("expected skip message, got %q", buf.String())
	}
}

func TestResolveRollbackTargetsAllSkipped(t *testing.T) {
	cfg := testConfig()
	// Neither service has a previous deploy
	mh := &mockHistoryProvider{
		previousDeploys: map[string]deploy{},
	}
	p := providers{
		history: map[string]historyProvider{
			"server": mh,
			"static": mh,
		},
	}

	var buf bytes.Buffer
	res, err := resolveRollbackTargets(context.Background(), cfg, p, nil, "staging", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.targets) != 0 {
		t.Fatalf("expected 0 targets, got %d: %v", len(res.targets), res.targets)
	}
	if len(res.skipped) != 2 {
		t.Fatalf("expected 2 skipped, got %d: %v", len(res.skipped), res.skipped)
	}
}

func TestResolveRollbackTargetsExplicitServiceNoPrev(t *testing.T) {
	cfg := testConfig()
	mh := &mockHistoryProvider{
		previousDeploys: map[string]deploy{},
	}
	p := providers{
		history: map[string]historyProvider{
			"server": mh,
			"static": mh,
		},
	}

	var buf bytes.Buffer
	res, err := resolveRollbackTargets(context.Background(), cfg, p, []string{"frontend"}, "staging", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.targets) != 0 {
		t.Fatalf("expected 0 targets, got %d", len(res.targets))
	}
	if len(res.skipped) != 1 || res.skipped[0] != "frontend" {
		t.Errorf("expected frontend skipped, got %v", res.skipped)
	}
	if !strings.Contains(buf.String(), "skipping frontend") {
		t.Errorf("expected skip message for frontend, got %q", buf.String())
	}
}

func TestResolveRollbackTargetsUnknownService(t *testing.T) {
	cfg := testConfig()
	p := providers{
		history: map[string]historyProvider{},
	}

	var buf bytes.Buffer
	_, err := resolveRollbackTargets(context.Background(), cfg, p, []string{"nonexistent"}, "staging", &buf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown service") {
		t.Errorf("expected 'unknown service' error, got: %v", err)
	}
}

func TestResolveRollbackTargetsServiceEnvNotFound(t *testing.T) {
	cfg := testConfig()
	p := providers{
		history: map[string]historyProvider{},
	}

	var buf bytes.Buffer
	_, err := resolveRollbackTargets(context.Background(), cfg, p, []string{"backend"}, "nonexistent", &buf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "has no environment") {
		t.Errorf("expected 'has no environment' error, got: %v", err)
	}
}

func TestResolveRollbackTargetsProviderError(t *testing.T) {
	cfg := testConfig()
	mh := &mockHistoryProvider{
		previousDeploys: map[string]deploy{},
		previousErrors: map[string]error{
			"backend:staging": errForTest("SSH connection refused"),
		},
	}
	p := providers{
		history: map[string]historyProvider{
			"server": mh,
			"static": mh,
		},
	}

	var buf bytes.Buffer
	_, err := resolveRollbackTargets(context.Background(), cfg, p, []string{"backend"}, "staging", &buf)
	if err == nil {
		t.Fatal("expected error from provider")
	}
	if !strings.Contains(err.Error(), "SSH connection refused") {
		t.Errorf("expected underlying error, got: %v", err)
	}
}

func errForTest(msg string) error {
	return &testError{msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
