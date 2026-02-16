package main

import (
	"context"
	"time"
)

type staticHistoryProvider struct{}

func (p *staticHistoryProvider) current(_ context.Context, service, env string) (deploy, error) {
	// TODO: read current-tag marker from S3 bucket
	return deploy{
		Service: service,
		Env:     env,
		Tag:     "stub-static-tag",
		Uptime:  1 * time.Hour,
	}, nil
}

func (p *staticHistoryProvider) previous(_ context.Context, service, env string) (deploy, error) {
	// TODO: read previous-tag marker from S3 bucket
	return deploy{Service: service, Env: env, Tag: "stub-static-prev-tag", Uptime: 48 * time.Hour}, nil
}
