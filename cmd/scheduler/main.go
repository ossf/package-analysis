package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/ossf/package-feeds/pkg/feeds"
	"go.uber.org/zap"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"

	"github.com/ossf/package-analysis/cmd/scheduler/proxy"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

type ManagerConfig struct {
	// ExcludeVersions is a list of regexp expressions, where if a version of
	// any package has a version string matching an expression in this list,
	// that package version will be ignored.
	ExcludeVersions []*regexp.Regexp

	// Ecosystem is the internal name of the ecosystem.
	Ecosystem pkgecosystem.Ecosystem
}

func (m *ManagerConfig) SkipVersion(version string) bool {
	if m == nil {
		return true
	}
	if m.ExcludeVersions == nil || len(m.ExcludeVersions) == 0 {
		return false
	}
	for _, f := range m.ExcludeVersions {
		if f.MatchString(version) {
			return true
		}
	}
	return false
}

// supportedPkgManagers lists the package managers Package Analysis can
// analyze. It is a map from ossf/package-feeds package types, to a
// config for the package manager's feed.
var supportedPkgManagers = map[string]*ManagerConfig{
	"npm":      {Ecosystem: pkgecosystem.NPM},
	"pypi":     {Ecosystem: pkgecosystem.PyPI},
	"rubygems": {Ecosystem: pkgecosystem.RubyGems},
	"packagist": {
		Ecosystem:       pkgecosystem.Packagist,
		ExcludeVersions: []*regexp.Regexp{regexp.MustCompile(`^dev-`), regexp.MustCompile(`\.x-dev$`)},
	},
	"crates": {Ecosystem: pkgecosystem.CratesIO},
}

func main() {
	subscriptionURL := os.Getenv("OSSMALWARE_SUBSCRIPTION_URL")
	topicURL := os.Getenv("OSSMALWARE_WORKER_TOPIC")
	logger := log.Initialize(os.Getenv("LOGGER_ENV"))

	err := listenLoop(logger, subscriptionURL, topicURL)
	if err != nil {
		logger.With(zap.Error(err)).Error("Error encountered")
	}
}

func listenLoop(logger *zap.Logger, subURL, topicURL string) error {
	ctx := context.Background()

	sub, err := pubsub.OpenSubscription(ctx, subURL)
	if err != nil {
		return err
	}

	topic, err := pubsub.OpenTopic(ctx, topicURL)
	if err != nil {
		return err
	}

	srv := proxy.New(topic, sub)
	logger.Info("Listening for messages to proxy...")

	err = srv.Listen(ctx, logger, func(m *pubsub.Message) (*pubsub.Message, error) {
		logger.With(
			zap.ByteString("body", m.Body),
		).Info("Handling message")
		pkg := feeds.Package{}
		if err := json.Unmarshal(m.Body, &pkg); err != nil {
			return nil, fmt.Errorf("error unmarshalling json: %w", err)
		}
		config, ok := supportedPkgManagers[pkg.Type]
		if !ok {
			return nil, fmt.Errorf("package type is not supported: %v", pkg.Type)
		}
		if config.SkipVersion(pkg.Version) {
			return nil, fmt.Errorf("package version '%v' is filtered for type: %v", pkg.Version, pkg.Type)
		}
		return &pubsub.Message{
			Body: []byte{},
			Metadata: map[string]string{
				"name":      pkg.Name,
				"ecosystem": config.Ecosystem.String(),
				"version":   pkg.Version,
			},
		}, nil
	})

	return err
}
