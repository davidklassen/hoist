package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
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
	type query struct {
		name string
		env  string
		svc  serviceConfig
	}

	var queries []query
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
			if _, ok := p.history[svc.Type]; !ok {
				continue
			}
			queries = append(queries, query{name: name, env: env, svc: svc})
		}
	}

	type result struct {
		index int
		row   statusRow
		err   error
	}

	results := make([]result, len(queries))
	var wg sync.WaitGroup
	for i, q := range queries {
		wg.Add(1)
		go func(i int, q query) {
			defer wg.Done()
			hp := p.history[q.svc.Type]
			cur, err := hp.current(ctx, q.name, q.env)
			if err != nil {
				results[i] = result{err: fmt.Errorf("getting status for %s/%s: %w", q.name, q.env, err)}
				return
			}

			row := statusRow{
				Service: q.name,
				Env:     q.env,
				Tag:     cur.Tag,
				Uptime:  cur.Uptime,
			}
			if q.svc.Type == "server" {
				row.Health = "healthy"
			} else {
				row.Health = "-"
			}
			results[i] = result{row: row}
		}(i, q)
	}
	wg.Wait()

	rows := make([]statusRow, 0, len(queries))
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		rows = append(rows, r.row)
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
	svcW, tagW, upW, healthW := len("SERVICE"), len("TAG"), len("UPTIME"), len("HEALTH")
	for _, r := range rows {
		label := r.Service + "-" + r.Env
		if len(label) > svcW {
			svcW = len(label)
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
