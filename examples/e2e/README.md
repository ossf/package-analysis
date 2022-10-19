# End to End with Package-Feeds integration

This example provides a simple way to spin up an end to end deployment of package-feeds and the package-analysis sub projects, allowing for easy demonstration of behaviour of the system as a whole.

## Running

```shell
$ cd examples/e2e # must be run from the e2e folder
$ docker-compose up -d
$ curl localhost:8080
```

Requesting `localhost:8080` will trigger package-feeds to poll its feeds and send the packages downstream to package-analysis.

You may need to run `curl` again if package-feeds is not yet running.

## Analysis Output

Output can be found at http://localhost:9000/minio/package-analysis,
using the following credentials for authentication:

- username: `minio`
- password: `minio123`

## Logs Access

`docker-compose logs -f feeds` will provide information on the packages which have been send downstream.

`docker-compose logs -f scheduler` will provide information on the packages which have been received and proxied onto the analysis workers.

`docker-compose logs -f analysis` provides too much stdout to be useful, outputs are sent to minio as mentioned above.

## Additional Notes

### Limitations

- Locally built sandbox images are currently ignored.

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

### Obtaining the necessary images (Optional)

Images are pulled automatically from the relevant repositories.

To build the necessary images yourself for the docker-compose, you can do the following:

```
# In package-analysis
cd build
./build_docker.sh

# In package-feeds
docker build . -t gcr.io/ossf-malware-analysis/scheduled-feeds:latest
```