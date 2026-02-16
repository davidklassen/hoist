package main

import (
	"context"
	"time"
)

type serverDeployer struct{}

func (d *serverDeployer) deploy(_ context.Context, _, _, _, _ string) error {
	// TODO: SSH into node, docker pull/stop/rm/run
	time.Sleep(1 * time.Second)
	return nil
}
