package worker

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// SandboxImageURLs returns the fully qualified references (image:tag) of every
// sandbox container image the worker needs in order to run analysis. If tag is
// empty, "latest" is used to match the default applied by the sandbox package.
func SandboxImageURLs(tag string) []string {
	if tag == "" {
		tag = "latest"
	}
	return []string{
		fmt.Sprintf("%s:%s", defaultStaticAnalysisImage, tag),
		fmt.Sprintf("%s:%s", defaultDynamicAnalysisImage, tag),
	}
}

// ImageChecker reports whether a single image reference is accessible. It
// returns a non-nil error if the image cannot be resolved. It exists so that
// CheckSandboxImagesAvailable can be unit tested without making real network
// calls.
type ImageChecker func(ctx context.Context, image string) error

// remoteImageChecker confirms that image is accessible in its registry by
// performing a HEAD/GET against the manifest using go-containerregistry's
// remote.Get. It does not pull the image.
func remoteImageChecker(ctx context.Context, image string) error {
	ref, err := name.ParseReference(image)
	if err != nil {
		return fmt.Errorf("parsing image reference: %w", err)
	}
	if _, err := remote.Get(ref, remote.WithContext(ctx), remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
		return err
	}
	return nil
}

// CheckSandboxImagesAvailable verifies that every image in images is accessible
// using check. It returns an error naming each image that could not be
// accessed, or nil if they are all available.
func CheckSandboxImagesAvailable(ctx context.Context, images []string, check ImageChecker) error {
	var errs []error
	for _, image := range images {
		if err := check(ctx, image); err != nil {
			errs = append(errs, fmt.Errorf("sandbox image %q is not accessible: %w", image, err))
		}
	}
	return errors.Join(errs...)
}

// CheckSandboxImagesAccessible confirms that all sandbox images configured for
// the given tag can be reached in their registry before the worker starts
// consuming messages. It fails loudly so a misconfigured or unavailable image
// is caught at startup rather than silently acking messages later on.
func CheckSandboxImagesAccessible(ctx context.Context, tag string) error {
	return CheckSandboxImagesAvailable(ctx, SandboxImageURLs(tag), remoteImageChecker)
}
