package main

import (
	"context"
	"fmt"
	"strings"
)

type serverLogsProvider struct {
	cfg config
}

func (p *serverLogsProvider) tail(_ context.Context, service, env string, n int, since string) error {
	svc := p.cfg.Services[service]
	ec := svc.Env[env]

	container := service
	follow := n == 0 && since == ""
	args := dockerLogsArgs(container, since, n, follow)

	// TODO: SSH into node and run docker logs
	fmt.Printf("[%s] Would run on node %q: docker %s\n", service, ec.Node, strings.Join(args, " "))
	return nil
}
