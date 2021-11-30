# End to End with Package-Feeds integration

This example provides a simple way to spin up an end to end deployment of package-feeds and the package-analysis sub projects, allowing for easy demonstration of behaviour of the system as a whole.

## Obtaining the necessary images

Images can be pulled from the relevant repositories however some require credentials (github for package-feeds).
To build the necessary images yourself for the docker-compose, you can do the following:

```
# In package-analysis
cd build
NOPUSH=true ./build_docker.sh

# In package-feeds
docker build . -t docker.pkg.github.com/ossf/package-feeds/packagefeeds:latest
```

## Running the example

```
docker-compose up -d
curl localhost:8080
```

Curling `localhost:8080` will trigger package-feeds to poll it's feeds and send the packages downstream to package-analysis. Outputs can be found at http://localhost:9000/minio/package-analysis,
using the credentials minio:minio123 for authentication.

`docker-compose logs -f feeds` will provide information on the packages which have been send downstream.

`docker-compose logs -f scheduler` will provide information on the packages which have been received and proxied onto the analysis workers.

`docker-compose logs -f analysis` provides too much stdout to be useful, outputs are sent to minio as mentioned above.
