package main

import "context"

type staticDeployer struct{}

func (d *staticDeployer) deploy(_ context.Context, _, _, _, _ string) error {
	// TODO: S3 sync + CloudFront invalidation
	return nil
}
