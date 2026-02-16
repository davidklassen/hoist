package main

import (
	"strings"
	"testing"
)

func TestDockerLogsArgs(t *testing.T) {
	tests := []struct {
		name      string
		container string
		since     string
		n         int
		follow    bool
		want      string
	}{
		{
			name:      "follow mode",
			container: "backend",
			follow:    true,
			want:      "logs -f backend",
		},
		{
			name:      "tail N lines",
			container: "backend",
			n:         100,
			want:      "logs --tail 100 backend",
		},
		{
			name:      "since duration",
			container: "backend",
			since:     "1h",
			want:      "logs --since 1h backend",
		},
		{
			name:      "tail and since",
			container: "backend",
			n:         50,
			since:     "2h",
			want:      "logs --tail 50 --since 2h backend",
		},
		{
			name:      "follow with since",
			container: "myapp",
			since:     "30m",
			follow:    true,
			want:      "logs --since 30m -f myapp",
		},
		{
			name:      "no flags",
			container: "backend",
			want:      "logs backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dockerLogsArgs(tt.container, tt.since, tt.n, tt.follow)
			result := strings.Join(got, " ")
			if result != tt.want {
				t.Errorf("dockerLogsArgs() = %q, want %q", result, tt.want)
			}
		})
	}
}
