[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/ossf/package-analysis/badge)](https://api.securityscorecards.dev/projects/github.com/ossf/package-analysis)

# Package Analysis

The Package Analysis project analyses the capabilities of packages available on open source repositories. The project looks for behaviors that indicate malicious software: 

- What files do they access? 
- What addresses do they connect to? 
- What commands do they run? 

The project also tracks changes in how packages behave over time, to identify when previously safe software begins acting suspiciously. 

This effort is meant to improve the security of open source software by detecting malicious behavior, informing consumers selecting packages, and providing researchers with data about the ecosystem. 

This code is designed to work with the
[Package Feeds](https://github.com/ossf/package-feeds) project,
and originally started there.

For examples of what this project has detected, check out the
[case studies](docs/case_studies.md).

## How it works

The project's components are:

- A [scheduler](./cmd/scheduler/) - creates jobs for the analysis worker from
  Package Feeds.
- Analysis (one-shot [analyze](./cmd/analyze/) and [worker](./cmd/worker/)) -
  collects package behavior data through static and dynamic analysis of each
  package.
- A [loader](./function/loader/) - pushes the analysis results into BigQuery.

The goal is for all of these components to work together and provide extensible,
community-run infrastructure to study behavior of open source packages and to
look for malicious software. We also hope that the components can be used
independently, to provide package feeds or runtime behavior data for anyone
interested.

The Package Analysis project currently consists of the following pipeline:

![image](docs/images/Pipeline%20diagram.png)

1. Package repositories are monitored for new packages.
1. Each new package is scheduled to be analyzed by a pool of workers.
1. A worker performs dynamic analysis of the package inside a sandbox.
1. Results are stored and imported into BigQuery for inspection.

Sandboxing via [gVisor](https://gvisor.dev/) containers ensures the packages are
isolated. Detonating a package inside the sandbox allows us to capture strace
and packet data that can indicate malicious interactions with the system as well
as network connections that can be used to leak sensitive data or allow remote
access.

## Public Data
This data is available in the public [BigQuery dataset](https://console.cloud.google.com/bigquery?d=packages&p=ossf-malware-analysis&t=analysis&page=table).

## Configuration

Configuration for these subprojects consist of a collection of environment
variables for the various endpoints. These endpoints are configured using
goclouddev compatible URL strings. In these cases, documentation will be linked
to and `DRIVER-Constructor` sections should be ignored in favour of `DRIVER`
sections as these are appropriate to the configurations in place throughout
these subprojects. Note that not all drivers will be supported but they can be
added quite simply with a minor patch to the repository. See the addition of
kafka for scheduler in
[one line](https://github.com/ossf/package-analysis/commit/985ab76a67d29d2fc8582b3920643e7eb963da8a#diff-8565ef29cfb886db7902792675eddce1e7a0ccfe33428a59e7f2e365b354af88R12).

An example of these variables can be found in the
[e2e example docker-compose](configs/e2e/docker-compose.yml).

### Analysis

`OSSMALWARE_WORKER_SUBSCRIPTION` - Can be used to set the subscription URL for
the data coming out of scheduler. Values should follow
[goclouddev subscriptions](https://gocloud.dev/howto/pubsub/subscribe/).

`OSSF_MALWARE_ANALYSIS_RESULTS` - **OPTIONAL**: Can be used to set the bucket
URL to publish results to. Values should follow
[goclouddev buckets](https://gocloud.dev/howto/blob/).

`OSSF_MALWARE_ANALYSIS_PACKAGES` - **OPTIONAL**: Can be used to set the bucket
URL to get custom uploaded packages from. Values should follow
[goclouddev buckets](https://gocloud.dev/howto/blob/).

`OSSF_MALWARE_NOTIFICATION_TOPIC` - **OPTIONAL**: Can be used to set the topic URL to
publish messages for consumption after a new package analysis is complete. Values should follow
[goclouddev publishing](https://gocloud.dev/howto/pubsub/publish/).

### Scheduler

`OSSMALWARE_WORKER_TOPIC` - Can be used to set the topic URL to publish data for
consumption by Analysis workers. Values should follow
[goclouddev publishing](https://gocloud.dev/howto/pubsub/publish/).

`OSSMALWARE_SUBSCRIPTION_URL` - Can be used to set the subscription URL for the
data coming out of [package-feeds](https://github.com/ossf/package-feeds).
Values should follow
[goclouddev subscriptions](https://gocloud.dev/howto/pubsub/subscribe/).

## Local Analysis

To run the analysis code locally, the easiest way is to use the Docker image
`gcr.io/ossf-malware-analysis/analysis`. This can be built with
`make build_analysis_image`, or the public images can be used instead.

This container uses `podman` to run a nested, sandboxed ([gVisor]) container for
analysis.

The commands below will dump the JSON results to `/tmp/results`
and full logs to `/tmp/dockertmp`.

[gVisor]: https://gvisor.dev/

### Live package

To run this on a live package (e.g. the latest version of the "Django" package on
[pypi.org](https://pypi.org))

```bash
$ scripts/run_analysis.sh -ecosystem pypi -package Django
```

Or with a specific version

```bash
$ scripts/run_analysis.sh -ecosystem pypi -package Django -version 4.1.3
```

### Local package

To run analysis on a local PyPi package named 'test',
located in local archive `/path/to/test.whl`


```bash
$ scripts/run_analysis.sh -ecosystem pypi -package test -local /path/to/test.whl
```

### Docker notes

(Note: these options are handled by the `scripts/run_analysis.sh` script).

`--privileged` and a compatible filesystem are required to properly run nested
containers. `-v /var/lib/containers:/var/lib/containers` is also used as it
allows caching the sandbox images and supports local developement.

## Development

### Testing
See `sample_packages/README.md` for how to use a sample package that simulates malicious activity for testing purposes.

### Required Dependencies

- Go v1.21
- Docker

# Contributing

If you want to get involved or have ideas you'd like to chat about, we discuss this project in the [OSSF Securing Critical Projects Working Group](https://github.com/ossf/wg-securing-critical-projects) meetings.

See the [Community Calendar](https://calendar.google.com/calendar?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ) for the schedule and meeting invitations.
