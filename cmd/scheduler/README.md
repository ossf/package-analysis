# Scheduler

This directory contains code to schedule analysis jobs based on incoming package update
notifications from [Package Feeds](https://github.com/ossf/package-feeds)

## Overview

The Scheduler is a Golang app that runs on Kubernetes and is deployed with [ko](https://github.com/google/ko).
It is currently deployed in a GKE cluster.

### Local deployment

Install ko

```bash
go install github.com/google/ko@latest
```

Then run

```bash
KO_DOCKER_REPO=gcr.io/ossf-malware-analysis ko resolve -f deployment.yaml | kubectl apply -f -
```

## Design

Package Feeds provides a Pub/Sub feed that provides package update notifications.
Each such notification corresponds to a single package event (update / new package).

The Scheduler handles ACKing the Package Feeds Pub/Sub feed, filtering out package ecosystems that are unsupported by Package Analysis and sending out another Pub/Sub notification to the Worker which triggers the actual analysis. The Worker then downloads, installs and imports (where applicable) the corresponding package, and monitors runtime behaviour.

The following ecosystems are supported
- [`PyPI`](https://pypi.org/)
- [`npmjs`](https://registry.npmjs.org/)
- [`RubyGems`](https://rubygems.org/)
- [`cargo`](https://crates.io/)
- [`Packagist`](https://packagist.org/)
