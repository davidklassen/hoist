package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

type ecrDescribeImagesAPI interface {
	DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error)
}

type serverBuildsProvider struct {
	ecr      ecrDescribeImagesAPI
	repoName string
}

// parseECRRepo extracts the repository name from a full ECR image URL.
// "123456.dkr.ecr.us-east-1.amazonaws.com/myapp" returns "myapp".
// Non-ECR image strings are returned as-is.
func parseECRRepo(image string) string {
	if idx := strings.Index(image, ".amazonaws.com/"); idx != -1 {
		return image[idx+len(".amazonaws.com/"):]
	}
	return image
}

func (p *serverBuildsProvider) listBuilds(ctx context.Context, limit, offset int) ([]build, error) {
	var all []build

	input := &ecr.DescribeImagesInput{
		RepositoryName: &p.repoName,
		Filter: &types.DescribeImagesFilter{
			TagStatus: types.TagStatusTagged,
		},
	}

	for {
		out, err := p.ecr.DescribeImages(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("describing ECR images: %w", err)
		}

		for _, img := range out.ImageDetails {
			for _, tagStr := range img.ImageTags {
				t, err := parseTag(tagStr)
				if err != nil {
					continue // skip non-hoist tags (cache, latest, etc.)
				}
				all = append(all, buildFromTag(t))
			}
		}

		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Time.After(all[j].Time)
	})

	if offset >= len(all) {
		return nil, nil
	}
	all = all[offset:]

	if limit < len(all) {
		all = all[:limit]
	}

	return all, nil
}
