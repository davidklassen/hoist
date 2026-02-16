package main

import (
	"context"
	"fmt"
)

type staticLogsProvider struct {
	cfg config
}

func (p *staticLogsProvider) tail(_ context.Context, service, env string, _ int, _ string) error {
	svc := p.cfg.Services[service]
	ec := svc.Env[env]

	// TODO: fetch CloudFront/S3 access logs
	fmt.Printf("[%s] Would fetch access logs from CloudFront %s (bucket: %s)\n", service, ec.CloudFront, ec.Bucket)
	return nil
}
