package main

import (
	"strings"
	"testing"
	"time"
)

func TestFormatBuildsTable(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		builds  []build
		limit   int
		hasMore bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "fewer builds than limit",
			builds:  []build{{Tag: "main-abc1234-20250615103000", Message: "fix bug", Author: "alice", Time: now}},
			limit:   10,
			hasMore: false,
			check: func(t *testing.T, output string) {
				if strings.Contains(output, "load more") {
					t.Error("should not show pagination hint")
				}
				if !strings.Contains(output, "main-abc1234-20250615103000") {
					t.Error("expected build tag in output")
				}
			},
		},
		{
			name: "more builds than limit",
			builds: []build{
				{Tag: "main-abc1234-20250615103000", Message: "fix bug", Author: "alice", Time: now},
				{Tag: "main-def5678-20250614103000", Message: "add feature", Author: "bob", Time: now.Add(-24 * time.Hour)},
			},
			limit:   2,
			hasMore: true,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "load more") {
					t.Error("expected pagination hint")
				}
			},
		},
		{
			name:    "no builds",
			builds:  nil,
			limit:   10,
			hasMore: false,
			check: func(t *testing.T, output string) {
				if output != "No builds found.\n" {
					t.Errorf("expected 'No builds found.' message, got %q", output)
				}
			},
		},
		{
			name: "with metadata",
			builds: []build{
				{Tag: "main-abc1234-20250615103000", Message: "fix: resolve timeout", Author: "alice", Time: now},
			},
			limit:   10,
			hasMore: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "fix: resolve timeout") {
					t.Error("expected commit message in output")
				}
				if !strings.Contains(output, "alice") {
					t.Error("expected author in output")
				}
			},
		},
		{
			name: "without metadata",
			builds: []build{
				{Tag: "main-abc1234-20250615103000", Time: now},
			},
			limit:   10,
			hasMore: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "main-abc1234-20250615103000") {
					t.Error("expected build tag in output")
				}
				if !strings.Contains(output, "BUILD") {
					t.Error("expected table headers")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatBuildsTable(tt.builds, tt.limit, tt.hasMore)
			tt.check(t, output)
		})
	}
}

func TestFormatBuildsTableHeaders(t *testing.T) {
	builds := []build{
		{Tag: "main-abc1234-20250615103000", Message: "msg", Author: "who", Time: time.Now()},
	}
	output := formatBuildsTable(builds, 10, false)
	for _, header := range []string{"BUILD", "COMMIT", "AUTHOR", "TIME"} {
		if !strings.Contains(output, header) {
			t.Errorf("expected header %q in output", header)
		}
	}
}

func TestFormatBuildTime(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{"zero time", time.Time{}, ""},
		{"valid time", time.Date(2025, 6, 15, 10, 30, 0, 0, time.Local), "Jun 15 10:30"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBuildTime(tt.time)
			if got != tt.want {
				t.Errorf("formatBuildTime() = %q, want %q", got, tt.want)
			}
		})
	}
}
