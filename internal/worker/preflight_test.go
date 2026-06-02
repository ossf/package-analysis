package worker

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSandboxImageURLs(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want []string
	}{
		{
			name: "explicit tag",
			tag:  "v1.2.3",
			want: []string{
				"gcr.io/ossf-malware-analysis/static-analysis:v1.2.3",
				"gcr.io/ossf-malware-analysis/dynamic-analysis:v1.2.3",
			},
		},
		{
			name: "empty tag defaults to latest",
			tag:  "",
			want: []string{
				"gcr.io/ossf-malware-analysis/static-analysis:latest",
				"gcr.io/ossf-malware-analysis/dynamic-analysis:latest",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SandboxImageURLs(tt.tag)
			if len(got) != len(tt.want) {
				t.Fatalf("SandboxImageURLs(%q) = %v; want %v", tt.tag, got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("SandboxImageURLs(%q)[%d] = %q; want %q", tt.tag, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCheckSandboxImagesAvailable_AllPresent(t *testing.T) {
	images := []string{"registry/a:latest", "registry/b:latest"}
	check := func(ctx context.Context, image string) error { return nil }

	if err := CheckSandboxImagesAvailable(context.Background(), images, check); err != nil {
		t.Errorf("CheckSandboxImagesAvailable() = %v; want nil", err)
	}
}

func TestCheckSandboxImagesAvailable_OneMissing(t *testing.T) {
	const missing = "registry/missing:latest"
	images := []string{"registry/present:latest", missing}

	check := func(ctx context.Context, image string) error {
		if image == missing {
			return errors.New("manifest unknown")
		}
		return nil
	}

	err := CheckSandboxImagesAvailable(context.Background(), images, check)
	if err == nil {
		t.Fatal("CheckSandboxImagesAvailable() = nil; want error naming the missing image")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("error %q does not name the missing image %q", err.Error(), missing)
	}
	if strings.Contains(err.Error(), "registry/present:latest") {
		t.Errorf("error %q should not name the available image", err.Error())
	}
}

func TestCheckSandboxImagesAvailable_MultipleMissing(t *testing.T) {
	images := []string{"registry/a:latest", "registry/b:latest"}
	check := func(ctx context.Context, image string) error {
		return errors.New("manifest unknown")
	}

	err := CheckSandboxImagesAvailable(context.Background(), images, check)
	if err == nil {
		t.Fatal("CheckSandboxImagesAvailable() = nil; want error")
	}
	for _, image := range images {
		if !strings.Contains(err.Error(), image) {
			t.Errorf("error %q does not name image %q", err.Error(), image)
		}
	}
}
