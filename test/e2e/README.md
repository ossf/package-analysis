# End to End Testing with Package-Feeds integration

This directory helps run end-to-end tests of the package analysis system
to ensure everything is working properly.
In particular, local changes to both the worker/analysis and sandbox images can be tested
before they are pushed to the docker registry.

The test is orchestrated using docker-compose, using an adapted setup based on the one in
`configs/e2e`. All the necessary commands can be run via the project Makefile.

## Running

### Starting the test

In the top-level project directory, run

```shell
$ make RELEASE_TAG=test build_prod_images sync_prod_sandboxes # rebuild images with 'test' tag
$ make e2e_test_start

```

### Stopping the test

In the top-level project directory, run

```shell
$ make e2e_test_stop
```

## Analysis Output

Output can be found at http://localhost:9000/minio/package-analysis,
using the following credentials for authentication:

- username: `minio`
- password: `minio123`

## Logs Access

In the top-level project directory, run

`make e2e_test_logs_feeds` to see information on the packages which have been send downstream.

`make e2e_test_logs_scheduler` to see information on the packages which have been received and proxied onto the analysis workers.

`make e2e_test_logs_analysis` to see analysis stdout (too much to be useful); better to check minio output as described above.

## PubSub (Kafka) Inspection

Output from the Kafka PubSub topics can be inspected using
[KafkaCat](https://github.com/edenhill/kcat).

1. Install `kafkacat` or `kcat` (e.g. `sudo apt install kafkacat`)
2. Run `kafkacat` to observe the topics:
    - package-feeds: `kafkacat -C -J -b localhost:9094 -t package-feeds`
    - workers: `kafkacat -C -J -b localhost:9094 -t workers`
    - notifications: `kafkacat -C -J -b localhost:9094 -t notifications`

## Troubleshooting

### Feeds does not start (missing config)

This can happen if `./config` is not world-readable. You will see the error message `open /config/feeds.yml: permission denied` in the feeds logs.

To fix simply run:

```shell
$ chmod ugo+rx ./config
$ chmod ugo+r ./config/feeds.yml
```

### Sandbox container is not starting (cgroups v2)

If the `analysis` logs show failures when trying to start the sandbox container, your machine may need to be configured to use cgroups v2.

To work with cgroups v2 you will need to:

1. add/edit `/etc/docker/daemon.json` and the following:

```json
{
    "default-cgroupns-mode": "host"
}
```

2. restart dockerd (if it is running). e.g.:

```shell
$ systemctl restart docker.service
```
