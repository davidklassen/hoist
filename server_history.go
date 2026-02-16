package main

import (
	"context"
	"time"
)

type serverHistoryProvider struct{}

func (p *serverHistoryProvider) current(_ context.Context, service, env string) (deploy, error) {
	// TODO: SSH into node, parse docker ps output for running container name and uptime
	return deploy{
		Service: service,
		Env:     env,
		Tag:     "stub-server-tag",
		Uptime:  3 * time.Hour,
	}, nil
}

func (p *serverHistoryProvider) previous(_ context.Context, service, env string) (deploy, error) {
	// TODO: SSH into node, docker inspect current container, read hoist.previous label
	return deploy{Service: service, Env: env, Tag: "stub-server-prev-tag", Uptime: 24 * time.Hour}, nil
}
