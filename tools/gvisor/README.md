# GVisor Scripts

## `runsc_compat.sh`

This script improves the compatibility of `runsc` when it is used by
[Podman](https://podman.io).

This project uses [GVisor](https://github.com/google/gvisor)'s OCI runtime
`runsc` to provide a sandbox for analyzing packages. The `runsc` sandbox is used
by setting it as the runtime for Podman running inside a Docker container.

Unfortunately there are slight differences in the flags passed from Podman
(specifically `conmon`) to the `runsc`.

In particular, when `podman exec` is called on a running container, the `-d`
(detach) flag is passed by `conmon` to the OCI runtime. However this flag is not
supported by `runsc`. Instead `runsc` supports `-detach`.

So, to ensure `runsc` works correctly with Podman this script will turn `-d`
into `-detach` when `exec` is called.