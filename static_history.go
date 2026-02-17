package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type s3GetObjectAPI interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type staticHistoryProvider struct {
	cfg config
	s3  s3GetObjectAPI
	now func() time.Time
}

func (p *staticHistoryProvider) current(ctx context.Context, service, env string) (deploy, error) {
	return p.readMarker(ctx, service, env, "current-tag")
}

func (p *staticHistoryProvider) previous(ctx context.Context, service, env string) (deploy, error) {
	return p.readMarker(ctx, service, env, "previous-tag")
}

func (p *staticHistoryProvider) readMarker(ctx context.Context, service, env, key string) (deploy, error) {
	bucket := p.cfg.Services[service].Env[env].Bucket

	out, err := p.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return deploy{}, nil
		}
		return deploy{}, fmt.Errorf("reading %s from s3://%s: %w", key, bucket, err)
	}
	defer out.Body.Close()

	body, err := io.ReadAll(out.Body)
	if err != nil {
		return deploy{}, fmt.Errorf("reading %s body: %w", key, err)
	}

	tag := strings.TrimSpace(string(body))
	if tag == "" {
		return deploy{}, nil
	}

	var uptime time.Duration
	if out.LastModified != nil {
		now := p.now
		if now == nil {
			now = time.Now
		}
		uptime = now().Sub(*out.LastModified)
	}

	return deploy{
		Service: service,
		Env:     env,
		Tag:     tag,
		Uptime:  uptime,
	}, nil
}
