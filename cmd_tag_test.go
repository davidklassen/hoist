package main

import "testing"

func TestResolveGitInfoFromEnvVars(t *testing.T) {
	t.Setenv("GITHUB_REF_NAME", "ci-branch")
	t.Setenv("GITHUB_SHA", "abc1234def5678")

	branch, sha, err := resolveGitInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "ci-branch" {
		t.Errorf("branch = %q, want %q", branch, "ci-branch")
	}
	if sha != "abc1234def5678" {
		t.Errorf("sha = %q, want %q", sha, "abc1234def5678")
	}
}

func TestResolveGitInfoEnvVarsPrecedence(t *testing.T) {
	// Even though we're in a git repo, env vars should take precedence
	t.Setenv("GITHUB_REF_NAME", "override-branch")
	t.Setenv("GITHUB_SHA", "override1234567")

	branch, sha, err := resolveGitInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "override-branch" {
		t.Errorf("branch = %q, want env var value", branch)
	}
	if sha != "override1234567" {
		t.Errorf("sha = %q, want env var value", sha)
	}
}

func TestResolveGitInfoFromLocalGit(t *testing.T) {
	// Clear CI env vars to force local git resolution
	t.Setenv("GITHUB_REF_NAME", "")
	t.Setenv("GITHUB_SHA", "")

	branch, sha, err := resolveGitInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch == "" {
		t.Error("expected non-empty branch from local git")
	}
	if sha == "" {
		t.Error("expected non-empty SHA from local git")
	}
	if len(sha) != 40 {
		t.Errorf("expected 40-char SHA from git rev-parse, got %d chars", len(sha))
	}
}
