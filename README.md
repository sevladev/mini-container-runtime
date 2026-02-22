# minic — Mini Container Runtime

A minimal container runtime built from scratch in Go using Linux primitives. No Docker libraries, no containerd — just raw namespaces, cgroups v2, pivot_root, OverlayFS, and veth networking.

## Why

If you use Docker and Kubernetes daily, building a container runtime from scratch proves you understand what happens under the hood. This project implements the core isolation mechanisms that production runtimes like runc use.

## Features

- **Process isolation** — PID, MNT, UTS, IPC, NET namespaces via `clone()`
- **Filesystem isolation** — `pivot_root` (not chroot) with OverlayFS copy-on-write layers
- **Resource limits** — cgroups v2 for memory, CPU, and PID limits
- **Networking** — bridge (`minic0`), veth pairs, IPAM, NAT with iptables
- **Image management** — pull Alpine Linux minirootfs, list local images
- **Container lifecycle** — run, exec, stop, rm, ps, logs
- **Background containers** — detach mode with log capture

## Architecture

```
+------------------------------------------------------------------+
|                         minic CLI (cobra)                         |
|  run | exec | ps | stop | rm | images | pull | logs              |
+------------------------------------------------------------------+
                              |
                              v
+------------------------------------------------------------------+
|                    container/lifecycle                            |
|  Run() -> Stop() -> Remove()                                     |
+------------------------------------------------------------------+
        |              |              |              |
        v              v              v              v
+-------------+ +-------------+ +-------------+ +-------------+
|  namespace  | |   cgroup    | | filesystem  | |   network   |
|  PID, MNT,  | | memory.max  | | overlayfs   | | bridge      |
|  UTS, NET,  | | cpu.max     | | pivot_root  | | veth pairs  |
|  IPC        | | pids.max    | | /proc mount | | IPAM, NAT   |
+-------------+ +-------------+ +-------------+ +-------------+
        |              |              |              |
        v              v              v              v
+------------------------------------------------------------------+
|                     Linux Kernel Syscalls                         |
|  clone | mount | pivot_root | write(cgroupfs) | netlink          |
+------------------------------------------------------------------+
```

## Prerequisites

- [Docker](https://www.docker.com/) (development runs inside a privileged container)
- [Make](https://www.gnu.org/software/make/)

## Quick Start

### 1. Start the development environment

```bash
make shell
```

This launches a privileged Docker container with Go, iptables, iproute2, and all required tools.

### 2. Build and install

```bash
go build -o /usr/local/bin/minic ./cmd/minic
```

### 3. Pull an image

```bash
minic pull alpine
```

### 4. Run a container

```bash
# Interactive shell
minic run alpine /bin/sh

# With resource limits
minic run --memory 50m --cpus 0.5 --pids 20 alpine /bin/sh

# Background
minic run -d alpine sleep 300
```

### 5. Manage containers

```bash
minic ps                        # List running containers
minic ps -a                     # List all containers
minic exec <id> /bin/ls         # Execute in running container
minic logs <id>                 # Show container logs
minic stop <id>                 # Stop a container
minic rm <id>                   # Remove a stopped container
```

Container IDs support prefix matching — `minic stop a3f` works if unambiguous.

## Commands

| Command | Description |
|---------|-------------|
| `minic run [flags] IMAGE CMD` | Create and run a container |
| `minic exec CONTAINER CMD` | Execute command in running container |
| `minic ps [-a]` | List containers |
| `minic stop CONTAINER` | Stop a running container |
| `minic rm CONTAINER` | Remove a stopped container |
| `minic logs CONTAINER` | Show container stdout/stderr |
| `minic pull IMAGE` | Download a rootfs image |
| `minic images` | List local images |

### Run flags

| Flag | Description |
|------|-------------|
| `--memory, -m` | Memory limit (e.g. `100m`, `1g`) |
| `--cpus` | CPU limit (e.g. `0.5`, `2.0`) |
| `--pids` | Max number of PIDs |
| `--hostname` | Container hostname |
| `--net` | Network mode: `bridge` (default) or `none` |
| `--detach, -d` | Run in background |
| `--volume, -v` | Bind mount (`host:container`) |

## How It Works

### Namespaces

Each container runs in its own set of Linux namespaces:

| Namespace | Isolation |
|-----------|-----------|
| PID | Process sees only its own PIDs (PID 1 inside) |
| MNT | Independent filesystem mounts |
| UTS | Separate hostname |
| NET | Own network stack, interfaces, routing |
| IPC | Isolated message queues and semaphores |

### pivot_root (not chroot)

The runtime uses `pivot_root` instead of `chroot` for security. After pivoting, the old root is unmounted and removed — there's no path back to the host filesystem. This is the same approach used by runc in production.

### Cgroups v2

Resource limits are applied via the cgroups v2 unified hierarchy by writing directly to `/sys/fs/cgroup/minic/container-<id>/`:

- `memory.max` — memory limit in bytes
- `cpu.max` — CPU quota (e.g. `50000 100000` = 50%)
- `pids.max` — process count limit

### OverlayFS

Each container gets a copy-on-write filesystem:

```
image rootfs (read-only lower layer)
        + container upper layer (writable)
        = merged view (what the container sees)
```

Multiple containers share the same base image without copying it.

### Networking

```
  Host
 +------------------------------------------+
 |  minic0 (bridge: 10.0.42.1/24)           |
 |     |           |                         |
 |   veth-a1     veth-b1        iptables NAT |
 +-----|-----------|-------------------------+
       |           |
 +-----|--+  +-----|--+
 | eth0   |  | eth0   |
 | .2     |  | .3     |
 | CNT-A  |  | CNT-B  |
 +---------+ +---------+
```

- Bridge `minic0` connects all containers
- Each container gets a veth pair and unique IP (10.0.42.x)
- iptables MASQUERADE provides internet access

### The /proc/self/exe Technique

The binary re-executes itself with an internal `init` command to run setup code inside the new namespaces before executing the user's command. This is the same technique used by runc.

## Project Structure

```
├── cmd/minic/main.go              # Entry point, intercepts "init"
├── internal/
│   ├── cli/                       # Cobra commands (run, exec, ps, stop, rm, ...)
│   ├── container/
│   │   ├── container.go           # Types: Config, State, ResourceLimits
│   │   ├── lifecycle.go           # Run, Stop, Remove (Linux)
│   │   ├── init.go                # Child process: pivot_root, mounts, exec
│   │   └── state.go               # JSON persistence, prefix lookup
│   ├── namespace/                 # Clone flags, nsenter exec
│   ├── cgroup/                    # Cgroups v2 via filesystem writes
│   ├── filesystem/
│   │   ├── pivot.go               # pivot_root implementation
│   │   ├── mounts.go              # /proc, /sys, /dev, device nodes
│   │   └── overlay.go             # OverlayFS mount/unmount
│   ├── network/
│   │   ├── bridge.go              # Bridge creation (netlink)
│   │   ├── veth.go                # Veth pairs + container config
│   │   ├── ipam.go                # Sequential IP allocation
│   │   ├── nat.go                 # iptables MASQUERADE
│   │   └── setup.go               # Orchestration
│   └── image/                     # Pull, list, manifest storage
├── pkg/
│   ├── idgen/                     # Container ID generation
│   └── units/                     # Parse "100m" → bytes
├── Dockerfile.dev                 # Go + Linux tools
├── docker-compose.yml             # Privileged dev environment
├── Makefile                       # build, shell, test, lint
└── go.mod
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go |
| CLI | Cobra |
| Namespaces | clone(), pivot_root, mount |
| Cgroups | v2 unified hierarchy (direct fs writes) |
| Filesystem | OverlayFS |
| Networking | netlink, iptables |
| Image | Alpine Linux minirootfs |

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| pivot_root vs chroot | pivot_root | More secure, used by runc in production |
| Cgroups library vs manual | Manual filesystem writes | Demonstrates deep understanding |
| netlink vs shell commands | netlink library | Type-safe, no subprocess overhead |
| OCI registry vs direct download | Direct Alpine download | Focus on runtime, not image distribution |

## License

MIT
