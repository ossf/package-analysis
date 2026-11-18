# Analyzing a local package on macOS (including Apple Silicon)

This is a step-by-step tutorial for scanning a **local-only** package — a package
archive sitting on your disk that has *not* been published to a registry — using
the Package Analysis tooling on macOS, **built and run natively** (no x86
emulation) on Apple Silicon (M1/M2/M3/M4) as well as Intel Macs.

By the end you will have built the analysis images for your Mac's architecture,
built a sample (deliberately malicious) Python package, run static and dynamic
analysis on it inside a gVisor sandbox, and read the JSON results describing what
the package did.

> New to the project? Read the top-level [README](../README.md) first for the
> big picture, then come back here.

## How local analysis works

The supported entry point is the [`scripts/run_analysis.sh`](../scripts/run_analysis.sh)
helper. It runs the `gcr.io/ossf-malware-analysis/analysis` Docker image, which
in turn uses `podman` to launch a nested, [gVisor](https://gvisor.dev/)-sandboxed
container that *detonates* the package. While the package installs and imports,
the sandbox captures system calls (file/command activity), network connections,
and DNS queries.

```
run_analysis.sh  ->  docker run … analysis  ->  podman + gVisor sandbox  ->  JSON results on your Mac
```

When you pass `-local <path>`, the script mounts your archive into the container
and analyzes it directly instead of downloading anything from a registry.

## Why you must build the images on Apple Silicon

The **public** images on `gcr.io/ossf-malware-analysis` are published for
`linux/amd64` only. On an Apple Silicon Mac, Docker Desktop runs an **arm64**
Linux VM, so those amd64 images run under qemu emulation. gVisor's `runsc` boots
its own kernel and intercepts the package's system calls, and syscall
interception *inside* an already-emulated container does not work — so `runsc`
fails to boot and you get:

```
ERROR  Dynamic analysis aborted  {"error": "sandbox failed (error starting container: exit status 125)"}
```

The fix is to **build the images natively for `linux/arm64`**. Docker Desktop's
arm64 Linux VM then runs them without emulation, and gVisor — which fully
supports arm64 — boots normally. This branch contains the small changes needed
for the images to build on arm64 (an architecture-aware AWS CLI download, an
arm64 PowerShell install, and a fix so `-offline` dynamic analysis works); see
[What changed for arm64](#what-changed-for-arm64) below.

On an **Intel Mac** the public amd64 images run natively, so you can either use
them as-is or follow the same build steps below (everything works either way).

## Prerequisites (macOS)

### 1. Docker Desktop

- Install [Docker Desktop](https://www.docker.com/products/docker-desktop/) and make sure it is **running** before you start.
- Give it enough resources (Docker Desktop → Settings → Resources): **≥ 6 GB memory** and **plenty of disk** are recommended — the dynamic sandbox image alone is ~3.3 GB, and you are running a container-inside-a-container.
- File sharing must include the directory holding your package and `/tmp`. On macOS `/tmp` is a symlink to `/private/tmp`, which Docker Desktop shares by default, so the default output locations work out of the box.

### 2. Go is NOT required on the host

The Go binaries are compiled **inside** the Docker image build, so you do not
need Go installed on your Mac. You only need Docker and the command-line tools
below.

### 3. GNU command-line tools (required)

`run_analysis.sh` is written for GNU/Linux userland. In particular the `-local`
code path calls `realpath -m`, and the BSD `realpath` that ships with macOS does
**not** understand `-m` — so on a stock Mac the script fails with
`realpath: illegal option -- m`. Install the GNU versions with
[Homebrew](https://brew.sh/):

```bash
brew install coreutils findutils util-linux
```

Then put the GNU tools ahead of the BSD ones on your `PATH` (add to `~/.zshrc` to
make it permanent):

```bash
export PATH="/opt/homebrew/opt/coreutils/libexec/gnubin:$PATH"
export PATH="/opt/homebrew/opt/findutils/libexec/gnubin:$PATH"
export PATH="/opt/homebrew/opt/util-linux/bin:$PATH"
```

> On Intel Macs, Homebrew lives under `/usr/local` instead of `/opt/homebrew` —
> adjust the paths accordingly.

Verify GNU `realpath` is now in front:

```bash
realpath --version   # should print "realpath (GNU coreutils) …"
```

`findmnt` (from `util-linux`) is optional — the script falls back gracefully if
it is missing.

### 4. Clone the repo

```bash
git clone https://github.com/ossf/package-analysis.git
cd package-analysis
```

## Step 1 — Build the images for your Mac's architecture

From the project root:

```bash
make build/sandbox/dynamic_analysis
make build/sandbox/static_analysis
make build/image/analysis
```

(Or just `make build`, which additionally builds the scheduler image you don't
need for local analysis.)

`docker build` on your Mac targets your native architecture by default, so on
Apple Silicon this produces `linux/arm64` images. The dynamic sandbox build is
the long pole — it installs Node, Python, Ruby, Rust, PHP, PowerShell, etc., and
takes several minutes the first time.

Confirm the architecture:

```bash
for img in analysis dynamic-analysis static-analysis; do
  printf "%-18s " "$img"
  docker image inspect "gcr.io/ossf-malware-analysis/${img}:latest" --format '{{.Os}}/{{.Architecture}}'
done
# analysis           linux/arm64
# dynamic-analysis   linux/arm64
# static-analysis    linux/arm64
```

## Step 2 — Make the local sandbox images available to the nested podman (macOS)

This is the one macOS-specific wrinkle. The `analysis` container runs `podman`
*inside* itself to launch the sandbox containers, and that nested podman reads
its image store from `/var/lib/containers`, which `run_analysis.sh` bind-mounts
from the host. Because your sandbox images are built locally (not pushed to a
registry), you must run with `-nopull` and pre-load them into that store.

On Linux this is done with `make sync` (`sudo podman load` on the host), but
macOS has no host podman. Two extra problems appear if you try a plain host
directory: it is shared into the VM over **virtiofs**, on which podman's overlay
storage driver cannot create whiteouts (`kernel does not support overlay fs`).

The solution is to back the podman store with a **Docker named volume** (which
lives on the Docker Desktop VM's own filesystem, where overlay works) and load
the images into it once. A helper script does exactly that:

```bash
scripts/load_sandbox_images_macos.sh
```

This saves the locally-built sandbox images, creates a named volume
`pa-containers`, and loads + retags them into the nested podman store. It only
needs to be re-run when you rebuild a sandbox image.

`run_analysis.sh` treats a `CONTAINER_DIR_OVERRIDE` value with **no leading
slash** as a Docker named volume, so the runs below point at `pa-containers`.

## Step 3 — Get a package to scan

The repo ships a sample Python package that simulates malicious behavior (it
tries to exfiltrate data over HTTPS and read credential files), which makes it a
great subject because the analysis will actually find something.

Build it:

```bash
cd sample_packages
make build_sample_python_package
cd ..
```

This produces a `.tar.gz` in `sample_packages/sample_python_package/output/`:

```bash
ls sample_packages/sample_python_package/output/*.tar.gz
# sample_packages/sample_python_package/output/sample_python_package-0.0.1.tar.gz
```

The package name is `sample_python_package` (from its `pyproject.toml`) — you'll
pass that as `-package` below.

> **Using your own package?** Point `-local` at any archive and set `-ecosystem`
> to match: `.whl`/`.tar.gz` for `pypi`, `.tgz` for `npm`, `.gem` for `rubygems`,
> `.zip` for `packagist`, `.crate`/`.tar.gz` for `crates.io`. The `-package`
> value should be the package's real name (used during the import phase).

## Step 4 — Run the analysis

Set up the environment once per shell (GNU tools on `PATH`, named-volume store):

```bash
export PATH="/opt/homebrew/opt/coreutils/libexec/gnubin:$PATH"
export CONTAINER_DIR_OVERRIDE=pa-containers
```

Then run. `-nopull` is **required** so the script uses your locally-built images
instead of fetching the public amd64 ones. `-nointeractive` avoids needing a TTY.

```bash
scripts/run_analysis.sh -nointeractive -nopull \
  -ecosystem pypi \
  -package sample_python_package \
  -local sample_packages/sample_python_package/output/sample_python_package-0.0.1.tar.gz
```

This runs **both** static and dynamic analysis. When it finishes you'll see a
"Finished analysis" summary listing where each result was written.

### Online vs. offline (a real trade-off)

- **Default (online):** the sandbox has network access, isolated by gVisor plus
  iptables rules that block private IP ranges. This is the path the project's own
  tests use. The package runs to completion and its real DNS/connection behavior
  is captured — for the sample you'll see it reach out to `www.httpbin.org`.
  Untrusted code does run with (contained) network here.

- **`-offline`:** disables the sandbox's network entirely (`--network=none`), so
  nothing can leave your machine — the safest way to detonate untrusted code.
  System calls, file activity, and executed commands are still captured. The
  catch: package managers can no longer download dependencies, so an install that
  needs the registry will fail at the install phase (e.g. pip: `Temporary failure
  in name resolution` → `No matching distribution found for setuptools`), and
  later phases won't run. Use offline for packages that vendor their dependencies
  or when install-time behavior is all you need.

  ```bash
  scripts/run_analysis.sh -nointeractive -nopull -offline \
    -ecosystem pypi -package sample_python_package \
    -local sample_packages/sample_python_package/output/sample_python_package-0.0.1.tar.gz
  ```

- **`-fully-offline`:** additionally disables the analysis **container's** network.
  Requires `-local` **and** `-nopull` (which you're already using).

You can also restrict to a single mode with `-mode static` or `-mode dynamic`.

## Step 5 — Read the results

Results are written to these directories on your Mac (created automatically):

| Directory | Contents |
|-----------|----------|
| `/tmp/results` | Dynamic analysis JSON (syscalls, network, commands) + execution log |
| `/tmp/staticResults` | Static analysis JSON (code signals) |
| `/tmp/writeResults` | Captured file-write contents |
| `/tmp/analyzedPackages` | A copy of the analyzed package |
| `/tmp/straceLogs` | Raw strace logs |
| `/tmp/dockertmp` | Debug logs (mounted as the container's `/tmp`) |

Inspect the dynamic results (Python's `json.tool` is built in on macOS):

```bash
cat /tmp/results/results.json | python3 -m json.tool | less
```

For the sample, an online dynamic run captures its attempted exfiltration — note
the socket and DNS query to `www.httpbin.org`:

```bash
python3 - <<'PY'
import json
d = json.load(open('/tmp/results/results.json'))
for phase, v in d['Analysis'].items():
    print(phase,
          '| sockets:', len(v.get('Sockets') or []),
          '| dns:', len(v.get('DNS') or []),
          '| files:', len(v.get('Files') or []),
          '| commands:', len(v.get('Commands') or []))
PY
# execute | sockets: 9 | dns: 1 | files: 238 | commands: 2   (e.g. 35.x.x.x:443 www.httpbin.org)
```

For an explanation of every field, see [docs/data_schema.md](data_schema.md).

## What changed for arm64

This branch includes three small changes so the toolchain builds and runs
natively on arm64. They are no-ops on amd64:

- **`sandboxes/dynamicanalysis/Dockerfile` — AWS CLI download.** The hard-coded
  `awscli-exe-linux-x86_64.zip` URL is now `…-$(uname -m).zip`, so the correct
  installer is fetched on both `x86_64` and `aarch64`.
- **`sandboxes/dynamicanalysis/Dockerfile` — PowerShell.** Microsoft's apt repo
  only publishes an amd64 PowerShell `.deb`, which breaks the dependency solver
  on arm64. PowerShell is now installed via apt on amd64 (unchanged) and from the
  official `linux-arm64` tarball on arm64.
- **`cmd/analyze/main.go` — offline dynamic analysis.** Dynamic analysis always
  starts a packet capture on the `cni-analysis` bridge, but `InitNetwork` (which
  creates that bridge) was skipped when `-offline` was set, so offline dynamic
  analysis failed before any phase ran (`failed to start packet capture …
  no such network interface`). The bridge is now created for dynamic analysis
  even when offline; the sandbox itself still runs with `--network=none`.

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `realpath: illegal option -- m`, or `-local` seems ignored | macOS BSD `realpath` lacks `-m` | Install GNU coreutils and fix `PATH` (see [Prerequisites](#3-gnu-command-line-tools-required)) |
| `sandbox failed (error starting container: exit status 125)` | gVisor `runsc` can't boot under amd64 emulation | Build the images natively (`make build`) and run with `-nopull`; don't use the public amd64 images on Apple Silicon |
| `kernel does not support overlay fs: unable to create kernel-style whiteout` | Nested podman store on a virtiofs host-dir bind mount | Use the named-volume approach (`scripts/load_sandbox_images_macos.sh` + `CONTAINER_DIR_OVERRIDE=pa-containers`) |
| `failed to start packet capture (… cni-analysis: no such network interface)` | Old code skipped bridge creation when `-offline` | Rebuild the `analysis` image from this branch (`make build/image/analysis`) — the fix is in `cmd/analyze/main.go` |
| Dynamic offline run ends with `No matching distribution found for …` | Offline install can't reach the registry for dependencies | Expected for packages that need network at install time — run online, or analyze a package with vendored deps |
| `Error: image '…' not found locally. Run 'make build' first.` (from the helper) | Images not built yet | Run the Step 1 build commands |
| `Cannot connect to the Docker daemon` | Docker Desktop isn't running | Start Docker Desktop and wait until it reports "running" |
| Mount / permission / file-sharing errors | Package path or `/tmp` not shared with Docker | Add the directory under Docker Desktop → Settings → Resources → File sharing |

## See also

- [README → Local Analysis](../README.md#local-analysis) — the condensed version
- [sample_packages/README.md](../sample_packages/README.md) — about the test packages
- [docs/data_schema.md](data_schema.md) — meaning of every result field
- [docs/case_studies.md](case_studies.md) — real malware this project has caught
