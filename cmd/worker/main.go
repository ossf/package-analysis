package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"

	"github.com/ossf/package-analysis/cmd/worker/pubsubextender"
	"github.com/ossf/package-analysis/internal/featureflags"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/notification"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/resultstore"
	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/staticanalysis"
	"github.com/ossf/package-analysis/internal/worker"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

const (
	localPkgPathFmt = "/local/%s"
)

// resultBucketPaths holds bucket paths for the different types of results.
type resultBucketPaths struct {
	dynamicAnalysis string
	staticAnalysis  string
	fileWrites      string
	analyzedPkg     string
}

type sandboxImageSpec struct {
	tag    string
	noPull bool
}

func copyPackageToLocalFile(ctx context.Context, packagesBucket *blob.Bucket, bucketPath string) (string, *os.File, error) {
	if packagesBucket == nil {
		return "", nil, errors.New("packages bucket not set")
	}

	// Copy remote package path to temporary file.
	r, err := packagesBucket.NewReader(ctx, bucketPath, nil)
	if err != nil {
		return "", nil, err
	}
	defer r.Close()

	f, err := os.CreateTemp("", "")
	if err != nil {
		return "", nil, err
	}

	if _, err := io.Copy(f, r); err != nil {
		return "", nil, err
	}

	if err := f.Close(); err != nil {
		return "", nil, err
	}

	return fmt.Sprintf(localPkgPathFmt, path.Base(bucketPath)), f, nil
}

func makeResultStores(dest resultBucketPaths) worker.ResultStores {
	resultStores := worker.ResultStores{}

	if dest.dynamicAnalysis != "" {
		resultStores.DynamicAnalysis = resultstore.New(dest.dynamicAnalysis, resultstore.ConstructPath())
	}
	if dest.staticAnalysis != "" {
		resultStores.StaticAnalysis = resultstore.New(dest.staticAnalysis, resultstore.ConstructPath())
	}
	if dest.fileWrites != "" {
		resultStores.FileWrites = resultstore.New(dest.fileWrites, resultstore.ConstructPath())
	}

	return resultStores
}

func handleMessage(ctx context.Context, msg *pubsub.Message, packagesBucket *blob.Bucket, resultStores *worker.ResultStores, imageSpec sandboxImageSpec, notificationTopic *pubsub.Topic) error {
	name := msg.Metadata["name"]
	if name == "" {
		slog.WarnContext(ctx, "name is empty")
		return nil
	}

	ecosystem := pkgecosystem.Ecosystem(msg.Metadata["ecosystem"])
	if ecosystem == "" {
		slog.WarnContext(ctx, "ecosystem is empty", "name", name)
		return nil
	}

	ctx = log.ContextWithAttrs(ctx,
		log.LabelAttr("ecosystem", ecosystem.String()),
		slog.String("name", name))

	manager := pkgmanager.Manager(ecosystem)
	if manager == nil {
		slog.WarnContext(ctx, "Unsupported pkg manager")
		return nil
	}

	version := msg.Metadata["version"]
	remotePkgPath := msg.Metadata["package_path"]

	resultsBucketOverride := msg.Metadata["results_bucket_override"]
	if resultsBucketOverride != "" {
		resultStores.DynamicAnalysis = resultstore.New(resultsBucketOverride, resultstore.ConstructPath())
	}

	ctx = log.ContextWithAttrs(ctx, slog.String("version", version))

	slog.InfoContext(ctx, "Got request",
		"package_path", remotePkgPath,
		"results_bucket_override", resultsBucketOverride,
	)

	localPkgPath := ""
	sandboxOpts := []sandbox.Option{sandbox.Tag(imageSpec.tag)}

	if remotePkgPath != "" {
		tmpPkgPath, pkgFile, err := copyPackageToLocalFile(ctx, packagesBucket, remotePkgPath)
		if err != nil {
			return err
		}

		defer os.Remove(pkgFile.Name())

		localPkgPath = tmpPkgPath
		sandboxOpts = append(sandboxOpts, sandbox.Volume(pkgFile.Name(), localPkgPath))
	}

	if imageSpec.noPull {
		sandboxOpts = append(sandboxOpts, sandbox.NoPull())
	}

	pkg, err := worker.ResolvePkg(manager, name, version, localPkgPath)
	if err != nil {
		slog.ErrorContext(ctx, "Error resolving package", "error", err)
		return err
	}

	dynamicSandboxOpts := append(worker.DynamicSandboxOptions(), sandboxOpts...)
	result, err := worker.RunDynamicAnalysis(ctx, pkg, dynamicSandboxOpts, "")
	if err != nil {
		return err
	}

	staticSandboxOpts := append(worker.StaticSandboxOptions(), sandboxOpts...)
	var staticResults analysisrun.StaticAnalysisResults
	// TODO run static analysis first and remove the if statement below
	if resultStores.StaticAnalysis != nil {
		staticResults, _, err = worker.RunStaticAnalysis(ctx, pkg, staticSandboxOpts, staticanalysis.All)
		if err != nil {
			return err
		}
	}

	if err := worker.SaveStaticAnalysisData(ctx, pkg, resultStores, staticResults); err != nil {
		return err
	}
	if err := worker.SaveDynamicAnalysisData(ctx, pkg, resultStores, result.AnalysisData); err != nil {
		return err
	}

	resultStores.AnalyzedPackageSaved = false

	if notificationTopic != nil {
		err := notification.PublishAnalysisCompletion(ctx, notificationTopic, name, version, ecosystem)
		if err != nil {
			return err
		}
	}

	return nil
}

func messageLoop(ctx context.Context, subURL, packagesBucket, notificationTopicURL string, imageSpec sandboxImageSpec, resultsBuckets *worker.ResultStores) error {
	sub, err := pubsub.OpenSubscription(ctx, subURL)
	if err != nil {
		return err
	}
	extender, err := pubsubextender.New(ctx, subURL, sub)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Subscription deadline extender",
		"deadline", extender.Deadline,
		"grace_period", extender.GracePeriod)

	// the default value of the notificationTopic object is nil
	// if no environment variable for a notification topic is set,
	// we pass in a nil notificationTopic object to handleMessage
	// and continue with the analysis with no notifications published
	var notificationTopic *pubsub.Topic
	if notificationTopicURL != "" {
		notificationTopic, err = pubsub.OpenTopic(ctx, notificationTopicURL)
		if err != nil {
			return err
		}
		defer notificationTopic.Shutdown(ctx)
	}

	var pkgsBkt *blob.Bucket
	if packagesBucket != "" {
		var err error
		pkgsBkt, err = blob.OpenBucket(ctx, packagesBucket)
		if err != nil {
			return err
		}
		defer pkgsBkt.Close()
	}

	slog.InfoContext(ctx, "Listening for messages to process...")
	for {
		msg, err := sub.Receive(ctx)
		if err != nil {
			// All subsequent receive calls will return the same error, so we bail out.
			return fmt.Errorf("error receiving message: %w", err)
		}

		// Create a message context to wrap the message_id that we are currently processing.
		msgCtx := log.ContextWithAttrs(ctx, slog.String("message_id", msg.LoggableID))

		me, err := extender.Start(msgCtx, msg, func() {
			slog.InfoContext(msgCtx, "Message Ack deadline extended", "message_meta", msg.Metadata)
		})
		if err != nil {
			// If Start fails it will always fail, so we bail out.
			// Nack the message if we can to indicate the failure.
			if msg.Nackable() {
				msg.Nack()
			}
			return fmt.Errorf("error starting message ack deadline extender: %w", err)
		}

		if err := handleMessage(msgCtx, msg, pkgsBkt, resultsBuckets, imageSpec, notificationTopic); err != nil {
			slog.ErrorContext(msgCtx, "Failed to process message", "error", err)
			if err := me.Stop(); err != nil {
				slog.ErrorContext(msgCtx, "Extender failed", "error", err)
			}
			if msg.Nackable() {
				msg.Nack()
			}
		} else {
			if err := me.Stop(); err != nil {
				slog.ErrorContext(msgCtx, "Extender failed", "error", err)
			}
			msg.Ack()
		}
	}
}

func main() {
	log.Initialize(os.Getenv("LOGGER_ENV"))

	ctx := context.Background()
	subURL := os.Getenv("OSSMALWARE_WORKER_SUBSCRIPTION")
	packagesBucket := os.Getenv("OSSF_MALWARE_ANALYSIS_PACKAGES")
	notificationTopicURL := os.Getenv("OSSF_MALWARE_NOTIFICATION_TOPIC")
	enableProfiler := os.Getenv("OSSF_MALWARE_ANALYSIS_ENABLE_PROFILER")

	if err := featureflags.Update(os.Getenv("OSSF_MALWARE_FEATURE_FLAGS")); err != nil {
		slog.Error("Failed to parse feature flags", "error", err)
		os.Exit(1)
	}

	resultsBuckets := resultBucketPaths{
		dynamicAnalysis: os.Getenv("OSSF_MALWARE_ANALYSIS_RESULTS"),
		staticAnalysis:  os.Getenv("OSSF_MALWARE_STATIC_ANALYSIS_RESULTS"),
		fileWrites:      os.Getenv("OSSF_MALWARE_ANALYSIS_FILE_WRITE_RESULTS"),
		analyzedPkg:     os.Getenv("OSSF_MALWARE_ANALYZED_PACKAGES"),
	}
	resultStores := makeResultStores(resultsBuckets)

	imageSpec := sandboxImageSpec{
		tag:    os.Getenv("OSSF_SANDBOX_IMAGE_TAG"),
		noPull: os.Getenv("OSSF_SANDBOX_NOPULL") != "",
	}

	sandbox.InitNetwork(ctx)

	// If configured, start a webserver so that Go's pprof can be accessed for
	// debugging and profiling.
	if enableProfiler != "" {
		go func() {
			slog.Info("Starting profiler")
			http.ListenAndServe(":6060", nil)
		}()
	}

	// Log the configuration of the worker at startup so we can observe it.
	slog.InfoContext(ctx, "Starting worker",
		"subscription", subURL,
		"package_bucket", packagesBucket,
		"results_bucket", resultsBuckets.dynamicAnalysis,
		"static_results_bucket", resultsBuckets.staticAnalysis,
		"file_write_results_bucket", resultsBuckets.fileWrites,
		"analyzed_packages_bucket", resultsBuckets.analyzedPkg,
		"image_tag", imageSpec.tag,
		"image_nopull", imageSpec.noPull,
		"topic_notification", notificationTopicURL,
		"feature_flags", featureflags.State(),
	)

	err := messageLoop(ctx, subURL, packagesBucket, notificationTopicURL, imageSpec, &resultStores)
	if err != nil {
		slog.ErrorContext(ctx, "Error encountered", "error", err)
	}
}
