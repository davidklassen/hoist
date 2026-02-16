package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

type statusRow struct {
	Service string
	Env     string
	Tag     string
	Uptime  time.Duration
	Health  string
}

func getStatus(ctx context.Context, cfg config, p providers, envFilter string) ([]statusRow, error) {
	var rows []statusRow
	for _, name := range sortedServiceNames(cfg) {
		svc := cfg.Services[name]
		envs := make([]string, 0, len(svc.Env))
		for e := range svc.Env {
			envs = append(envs, e)
		}
		sort.Strings(envs)

		for _, env := range envs {
			if envFilter != "" && env != envFilter {
				continue
			}

			hp, ok := p.history[svc.Type]
			if !ok {
				continue
			}

			cur, err := hp.current(ctx, name, env)
			if err != nil {
				return nil, fmt.Errorf("getting status for %s/%s: %w", name, env, err)
			}

			row := statusRow{
				Service: name,
				Env:     env,
				Tag:     cur.Tag,
				Uptime:  cur.Uptime,
			}

			if svc.Type == "server" {
				row.Health = "healthy"
			} else {
				row.Health = "-"
			}

			rows = append(rows, row)
		}
	}
	return rows, nil
}

func formatUptime(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func formatStatusTable(rows []statusRow) string {
	if len(rows) == 0 {
		return "No services found.\n"
	}

	// Calculate column widths
	svcW, envW, tagW, upW, healthW := len("SERVICE"), len("ENV"), len("TAG"), len("UPTIME"), len("HEALTH")
	for _, r := range rows {
		label := r.Service + "-" + r.Env
		if len(label) > svcW {
			svcW = len(label)
		}
		if len(r.Env) > envW {
			envW = len(r.Env)
		}
		if len(r.Tag) > tagW {
			tagW = len(r.Tag)
		}
		u := formatUptime(r.Uptime)
		if len(u) > upW {
			upW = len(u)
		}
		if len(r.Health) > healthW {
			healthW = len(r.Health)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%-*s  %-*s  %-*s  %-*s\n", svcW, "SERVICE", tagW, "TAG", upW, "UPTIME", healthW, "HEALTH")
	for _, r := range rows {
		label := r.Service + "-" + r.Env
		fmt.Fprintf(&b, "%-*s  %-*s  %-*s  %-*s\n", svcW, label, tagW, r.Tag, upW, formatUptime(r.Uptime), healthW, r.Health)
	}
	return b.String()
}
