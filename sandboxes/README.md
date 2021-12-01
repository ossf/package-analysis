# Sandboxes for Analysis

## Sandbox Image Testing

By default the analysis command will update the sandbox images from the grc.io
repository.

To use local images for analysis:

1. The images need to be manually pulled into `/var/lib/containers` using the
   `buildah` tool included with `podman`.
2. `/var/lib/containers` then needs to be mounted inside the container so it can
   be used by `podman`.
3. Finally, the `analysis` command should be run using the `-nopull` option to
   use the local image only.

For example, assuming there is a locally built npm image:

```bash
# Prepare the local machine.
$ sudo apt install podman
$ mkdir /tmp/results

# Populate the local podman images with the image we want to use from docker.
$ sudo buildah pull docker-daemon:gcr.io/ossf-malware-analysis/node:latest

# Run the analysis locally, without pulling the image again.
$ docker run --privileged -ti \
    -v /tmp/results:/results \
    -v /var/lib/containers:/var/lib/containers \
    gcr.io/ossf-malware-analysis/analysis analyze \
    -package test -ecosystem npm \
    -upload file:///results/ -nopull
```

## Adding a new Runtime Analysis Sandbox

Each runtime analysis sandbox requires:

- a command to trigger the runtime analysis phases, usually named
  `analyze.[ext]`.
- a `Dockerfile` for constructing the sandbox.

### Analysis Command

Each command must conform to the following API specification.

#### Arguments

```
Usage:
  analyze.[ext] [--local FILE | --version VERSION] PHASE PACKAGE
```

- `PHASE` - the phase name:
  - must support at least `all` and `install`.
  - may also include `import` and any other phases useful to the
    ecosystem.
- `PACKAGE` - package:
  - refers to a local file contain a package to install if `--local` is set;
    otherwise
  - the name of the package in the package repository to install.
- `--version VERSION` - a version of the package (optional):
  - if no version is not set when passing the name of the package, the version
    used will be chosen by the package manager.
  - only used if `PHASE` is `all` or `install`.
  - cannot be used with `--local`
- `--local FILE` - a local file to install instead of using the registry
  (optional):
  - if set, `FILE` will be installed. `PACKAGE` must match the name of the
    package being installed.
  - only used if `PHASE` is `all` or `install`.
  - canot be used with `--version`

#### Phases

##### all

- should run all the phases in the following order:
  - `install`
  - `import`
  - any other phases specified

**NOTE:** this is a transitional step while support is added to the library

##### install

Install the package specified using the standard mechanism for the given
package ecosystem.

For example, it may execute a shell command:

```shell
$ pip install django==9.3.4
```

- This phase will always be called first.
- This phase is required for all other phases.
- Successful install must return an exit status code of 0.
- An unsuccessful install must return an exit status code of any number *other
  than* 0.

##### import

Iterates through the installed package's modules, etc and attempts to import
them.

This is relevant to languages that execute code at import time.

- Errors that occur while importing a specific module should be logged, but
  execution should continue.
- If all imports were successfully attempted (regardless of whether the import
  worked or not) an exit status code of 0 must be returned.
- If an issue prevented all imports from being attempted an exit status code
  *other than* 0 must be returned.

##### Any other phases

- If the phase completed successfully an exit status code of 0 must be returned.
- It the phase failed to complete an exit status code *other than* 0 must be
  returned.

*Note*: Failures resulting in a non-0 exit code should exclude any failures of
the package being analyzed to function. For example, an import failing because
a dependency is missing is not treated as a failure.

#### Stdio

##### stdin

All phases run without receiving any input on `stdin`.

It is possible that a package being analyzed may request user input. It may be
preferable to send an `EOF` or close the `stdin` file handle.

##### stdout

Output from `stdout` will be considered debug level output.

##### stderr

Output from `stderr` will be considered info level output.

#### Additional Notes

The command should be as minimalistic as possible and not contribute too much to
the output of the runtime analysis.

### Dockerfile

The `Dockerfile` for the sandbox is responsible for creating a container where
the packages can be installed and analyzed.

The container must include:

- the analysis command
- the package manager used by the analysis command
- system libraries used by the majority of packages

The `Dockerfile` must also set the analysis command as the `ENTRYPOINT`.

### Wiring It Up

1. Update [build_docker.sh](../build/build_docker.sh) to reference the image.
2. Extend [internal/pkgecosystem](../internal/pkgecosystem) to add support for
   the new sandbox.
3. Ensure your new ecosystem is supported by
   [package-feeds](https://github.com/ossf/package-feeds).
4. Make sure [cmd/scheduler](../cmd/scheduler) marks the new ecosystem as
   supported.
