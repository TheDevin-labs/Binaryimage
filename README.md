# Binary Image

> *A guild instrument of TheDevinLabs — forged in Go, tempered by bare iron.*

**Binary Image** is a two-part Go toolchain for compiling Go source directly into raw machine instructions and dispatching those instructions to the CPU without the mediation of an operating system, runtime, or format header. It speaks to the processor in its own tongue.

---

## The .sbimg Format

All compiled output is sealed within a `.sbimg` *(Secure Binary Image)* file. The format is purposefully minimal:

```
Offset  Size   Field
──────────────────────────────
0x00    3      Magic bytes  [ 0x53 0x42 0x49 ]  ("SBI")
0x03    1      Version
0x04    1      Arch          0x01 = amd64  |  0x02 = arm64
0x05    3      Reserved
0x08    8      Code size     (little-endian uint64)
0x10    N      Raw machine code
```

No ELF header. No PE header. No OS assumptions. Raw instructions from offset `0x10` onward.

---

## Architecture

```
┌─────────────────────────────────────────┐
│               binary-image              │
├─────────────────┬───────────────────────┤
│  compile        │  run                  │
│                 │                       │
│  Go source      │  .sbimg file          │
│      │          │       │               │
│  go build       │  sbimg.Read()         │
│  (GOOS=linux)   │       │               │
│      │          │  mmap + mprotect      │
│  objcopy        │  (RX memory)          │
│  (.text only)   │       │               │
│      │          │  signal.Arm()         │
│  sbimg.Write()  │       │               │
│      │          │  fn() — CPU executes  │
│  .sbimg         │       │               │
│                 │  fault intercept      │
│                 │  or clean exit        │
└─────────────────┴───────────────────────┘
```

---

## Components

### `pkg/sbimg` — Format Layer

Read and write `.sbimg` files. Validates magic, version, and architecture on read. Writes raw code with a packed binary header.

### `internal/compiler` — Compile Pipeline

Invokes `go build` with `CGO_ENABLED=0` and `GOOS=linux` to produce an ELF binary, then strips it to raw `.text` section bytes via `objcopy`. The result is sealed into a `.sbimg` file.

### `internal/runner` — Bare-Metal Executor *(linux/amd64, linux/arm64)*

- Allocates anonymous memory via `mmap`
- Copies machine code into the allocation
- Marks the region executable via `mprotect`
- Arms hardware fault interception via `signal.Arm`
- Casts the memory address to a Go function pointer and invokes it
- On hardware exception *(SIGSEGV, SIGBUS, SIGILL, SIGFPE, SIGTRAP)* — intercepts the signal, extracts fault metadata, logs it, and terminates safely without harming the host OS

### `internal/signal` — Fault Guard

Registers OS signal handlers for hardware exceptions before execution begins. On fault, captures signal, code, and faulting address, returns control to the runner without process crash.

---

## Requirements

| Dependency | Purpose |
|---|---|
| Go 1.22+ | Build toolchain |
| `objcopy` (binutils) | Strip ELF to raw `.text` |
| Linux kernel | `mmap` / `mprotect` / signal delivery |
| amd64 or arm64 host | Direct execution |

---

## Installation

```sh
git clone https://github.com/TheDevinLabs/binary-image
cd binary-image
make all
```

Binaries land in `release/`.

---

## Usage

### Compile a Go package into a `.sbimg`

```sh
release/binary-image-compile -o payload.sbimg ./path/to/package
```

Target a specific architecture:

```sh
release/binary-image-compile -arch arm64 -o payload.sbimg ./path/to/package
```

### Execute a `.sbimg` directly on the CPU

```sh
release/binary-image-run payload.sbimg
```

Output on clean exit:

```
loading: payload.sbimg  arch=amd64  size=4096 bytes
execution complete
```

Output on hardware fault:

```
runner: hardware exception intercepted
  signal  : segmentation fault
  code    : 11
  address : 0x7f3a00000000
terminated: hardware exception — hardware fault: signal=segmentation fault code=11 addr=0x7f3a00000000
```

---

## Security Posture

The runner makes no attempt to sandbox the executed code at the syscall or memory isolation level. It is a **trusted injector** — the `.sbimg` payload runs with the full privileges of the invoking process. The fault guard exists to prevent a hardware exception from propagating as an unhandled OS signal that would crash the host program; it is not a security boundary.

Run untrusted `.sbimg` files inside a container or sandboxed environment.

---

## Platform Support

| OS | amd64 | arm64 |
|---|---|---|
| Linux | ✓ compile + run | ✓ compile + run |
| macOS | ✓ compile only | ✓ compile only |
| Windows | ✗ | ✗ |

---

## Project Structure

```
binary-image/
├── cmd/
│   ├── compile/        entry point — compile pipeline
│   └── run/            entry point — bare-metal executor
├── internal/
│   ├── compiler/       go build + objcopy pipeline
│   ├── runner/         mmap / mprotect / exec
│   └── signal/         hardware fault interception
├── pkg/
│   └── sbimg/          .sbimg format: read / write
├── Makefile
└── go.mod
```

---

## License

BSD 3 Clause © TheDevinLabs
