# Package Analysis

This repo contains a few components to aid in the analysis of open source
packages, in particular to look for malicious software. This code is designed
to work with the [Package Feeds](https://github.com/ossf/package-feeds) project,
and originally started there.

The components are:

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
[e2e example docker-compose](examples/e2e/docker-compose.yml).

### Analysis

`OSSMALWARE_WORKER_SUBSCRIPTION` - Can be used to set the subscription URL for
the data coming out of scheduler. Values should follow
[goclouddev subscriptions](https://gocloud.dev/howto/pubsub/subscribe/).

`OSSF_MALWARE_ANALYSIS_RESULTS` - **OPTIONAL**: Can be used to set the bucket
URL to publish results to. Values should follow
[goclouddev buckets](https://gocloud.dev/howto/blob/).

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
`gcr.io/ossf-malware-analysis/analysis`. This can be built from
`./build/build_docker.sh`.

This container uses `podman` to run a nested, sandboxed ([gVisor]) container for
analysis.

The following commands will dump the JSON result to `/tmp/results`.

[gVisor]: https://gvisor.dev/

### Live package

To run this on a live package (e.g. the "Django" package on
[pypi.org](https://pypi.org))

```bash
$ mkdir /tmp/results
$ docker run --privileged -ti \
    -v /tmp/results:/results \
    gcr.io/ossf-malware-analysis/analysis analyze \
    -package Django -ecosystem pypi \
    -upload file:///results/
```

### Local package

To run this on a local package archive (e.g. `/path/to/test.whl` for a package
named `test`), it needs to be mounted into the the container.

```bash
$ mkdir /tmp/results
$ docker run --privileged -ti \
    -v /tmp/results:/results \
    -v /path/to/test.whl:/test.whl \
    gcr.io/ossf-malware-analysis/analysis analyze \
    -local /test.whl -package test -ecosystem pypi \
    -upload file:///results/
```

# Contributing

If you want to get involved or have ideas you'd like to chat about, we discuss this project in the [OSSF Securing Critical Projects Working Group](https://github.com/ossf/wg-securing-critical-projects) meetings.

See the [Community Calendar](https://calendar.google.com/calendar?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ) for the schedule and meeting invitations.
