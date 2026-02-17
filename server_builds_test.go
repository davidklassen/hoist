package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

type stubECR struct {
	pages []ecr.DescribeImagesOutput
	err   error
}

func (s *stubECR) DescribeImages(_ context.Context, _ *ecr.DescribeImagesInput, _ ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error) {
	if s.err != nil {
		return nil, s.err
	}
	if len(s.pages) == 0 {
		return &ecr.DescribeImagesOutput{}, nil
	}
	page := s.pages[0]
	s.pages = s.pages[1:]
	return &page, nil
}

func TestParseECRRepo(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  string
	}{
		{"full ECR URL", "123456.dkr.ecr.us-east-1.amazonaws.com/demoagent-backend", "demoagent-backend"},
		{"ECR URL with nested repo", "123456.dkr.ecr.us-east-1.amazonaws.com/org/backend", "org/backend"},
		{"short name", "myapp/backend", "myapp/backend"},
		{"bare name", "backend", "backend"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseECRRepo(tt.image)
			if got != tt.want {
				t.Errorf("parseECRRepo(%q) = %q, want %q", tt.image, got, tt.want)
			}
		})
	}
}

func TestServerBuildsFiltering(t *testing.T) {
	stub := &stubECR{
		pages: []ecr.DescribeImagesOutput{
			{
				ImageDetails: []types.ImageDetail{
					{ImageTags: []string{"main-abc1234-20250101100000"}},
					{ImageTags: []string{"cache"}},
					{ImageTags: []string{"latest"}},
					{ImageTags: []string{"not-a-valid-tag"}},
					{ImageTags: []string{"feat-xyz-def5678-20250101090000"}},
				},
			},
		},
	}

	p := &serverBuildsProvider{ecr: stub, repoName: "test-repo"}
	builds, err := p.listBuilds(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(builds) != 2 {
		t.Fatalf("expected 2 builds, got %d", len(builds))
	}
}

func TestServerBuildsSorting(t *testing.T) {
	stub := &stubECR{
		pages: []ecr.DescribeImagesOutput{
			{
				ImageDetails: []types.ImageDetail{
					{ImageTags: []string{"main-abc1234-20250101080000"}},
					{ImageTags: []string{"main-def5678-20250101100000"}},
					{ImageTags: []string{"main-aae9012-20250101090000"}},
				},
			},
		},
	}

	p := &serverBuildsProvider{ecr: stub, repoName: "test-repo"}
	builds, err := p.listBuilds(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(builds) != 3 {
		t.Fatalf("expected 3 builds, got %d", len(builds))
	}
	if builds[0].SHA != "def5678" {
		t.Errorf("builds[0].SHA = %q, want %q", builds[0].SHA, "def5678")
	}
	if builds[1].SHA != "aae9012" {
		t.Errorf("builds[1].SHA = %q, want %q", builds[1].SHA, "aae9012")
	}
	if builds[2].SHA != "abc1234" {
		t.Errorf("builds[2].SHA = %q, want %q", builds[2].SHA, "abc1234")
	}
}

func TestServerBuildsOffsetLimit(t *testing.T) {
	stub := &stubECR{
		pages: []ecr.DescribeImagesOutput{
			{
				ImageDetails: []types.ImageDetail{
					{ImageTags: []string{"main-aaa1111-20250101050000"}},
					{ImageTags: []string{"main-bbb2222-20250101040000"}},
					{ImageTags: []string{"main-ccc3333-20250101030000"}},
					{ImageTags: []string{"main-ddd4444-20250101020000"}},
					{ImageTags: []string{"main-eee5555-20250101010000"}},
				},
			},
		},
	}

	p := &serverBuildsProvider{ecr: stub, repoName: "test-repo"}
	builds, err := p.listBuilds(context.Background(), 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(builds) != 2 {
		t.Fatalf("expected 2 builds, got %d", len(builds))
	}
	// After sorting descending: aaa(05), bbb(04), ccc(03), ddd(02), eee(01)
	// offset=1 skips aaa, limit=2 gives bbb and ccc
	if builds[0].SHA != "bbb2222" {
		t.Errorf("builds[0].SHA = %q, want %q", builds[0].SHA, "bbb2222")
	}
	if builds[1].SHA != "ccc3333" {
		t.Errorf("builds[1].SHA = %q, want %q", builds[1].SHA, "ccc3333")
	}
}

func TestServerBuildsOffsetPastEnd(t *testing.T) {
	stub := &stubECR{
		pages: []ecr.DescribeImagesOutput{
			{
				ImageDetails: []types.ImageDetail{
					{ImageTags: []string{"main-abc1234-20250101100000"}},
				},
			},
		},
	}

	p := &serverBuildsProvider{ecr: stub, repoName: "test-repo"}
	builds, err := p.listBuilds(context.Background(), 10, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if builds != nil {
		t.Errorf("expected nil builds for offset past end, got %d", len(builds))
	}
}

func TestServerBuildsPagination(t *testing.T) {
	nextToken := "page2"
	stub := &stubECR{
		pages: []ecr.DescribeImagesOutput{
			{
				ImageDetails: []types.ImageDetail{
					{ImageTags: []string{"main-abc1234-20250101100000"}},
				},
				NextToken: &nextToken,
			},
			{
				ImageDetails: []types.ImageDetail{
					{ImageTags: []string{"main-def5678-20250101090000"}},
				},
			},
		},
	}

	p := &serverBuildsProvider{ecr: stub, repoName: "test-repo"}
	builds, err := p.listBuilds(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(builds) != 2 {
		t.Fatalf("expected 2 builds from 2 pages, got %d", len(builds))
	}
}

func TestServerBuildsECRError(t *testing.T) {
	stub := &stubECR{err: fmt.Errorf("access denied")}
	p := &serverBuildsProvider{ecr: stub, repoName: "test-repo"}

	_, err := p.listBuilds(context.Background(), 10, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServerBuildsEmpty(t *testing.T) {
	stub := &stubECR{
		pages: []ecr.DescribeImagesOutput{
			{ImageDetails: []types.ImageDetail{}},
		},
	}

	p := &serverBuildsProvider{ecr: stub, repoName: "test-repo"}
	builds, err := p.listBuilds(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if builds != nil {
		t.Errorf("expected nil builds, got %d", len(builds))
	}
}
