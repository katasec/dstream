# Performance Baseline: stdin/stdout Pipe Relay

> Benchmark results and architectural rationale for DStream's stdin/stdout provider model.

## Benchmark Results

**Environment**: Apple Silicon (arm64), macOS, Go 1.25.4  
**Date**: 2026-04-05  
**Method**: Real subprocess relay — input provider → CLI scanner → output provider, using OS pipes. Measures the actual `executeFullPipeline` data path.

| Scenario | Per message | Throughput | Notes |
|----------|------------|------------|-------|
| Typical CDC envelope (~500B) | **9.07µs** | 110K msg/s | Representative row-change event |
| Small message (~43B) | **8.25µs** | 121K msg/s | Minimal JSON envelope |
| Large message (~3.5KB) | **11.07µs** | 90K msg/s | Wide table with 50 columns |
| Handshake latency | **6.37ms** | one-time | Process start → ready signal |

### What the numbers mean

The relay overhead — the cost DStream adds by being in the middle — is **~9µs per message**. For context:

- A MSSQL CDC poll: **10–100ms**
- An Azure Service Bus send: **5–50ms**
- The pipe relay: **0.009ms**

The relay adds <0.01% to end-to-end latency. It is effectively invisible in any real pipeline.

## Comparison: stdin/stdout vs gRPC

gRPC (as used by HashiCorp go-plugin, the previous DStream plugin model) typically adds **50–200µs per message** due to protobuf serialization, HTTP/2 framing, header compression, and runtime machinery.

| Dimension | stdin/stdout pipes | gRPC (go-plugin) |
|-----------|-------------------|-----------------|
| Per-message latency | ~9µs | ~50–200µs |
| Serialization | None — raw bytes pass through | Protobuf marshal/unmarshal at every boundary |
| Transport | OS kernel pipe — memory copy, no network stack | TCP + HTTP/2 + TLS (even localhost) |
| Framework overhead | Zero — `scanner.Scan()` → `fmt.Fprintln()` | Connection management, stream mux, flow control, keepalives |
| Language support | Any language, any platform — even bash/PowerShell | Requires gRPC codegen and runtime per language |
| Schema enforcement | Convention-based (JSON lines) | Strong (protobuf IDL) |
| Bidirectional streaming | Not native (would need protocol extension) | Built-in |
| Backpressure | Implicit via pipe buffer pressure | Explicit via flow control frames |

**Where gRPC wins**: bidirectional streaming, schema enforcement, backpressure signaling.  
**Where stdin/stdout wins**: everything else for this use case.

Note: Terraform's go-plugin itself moved toward simpler stdio transport in later versions for similar reasons.

## Why stdin/stdout

The choice of stdin/stdout as the provider communication model was a deliberate architectural decision, not a simplification compromise. Three motivations:

### 1. Universal language and platform support

Any language on any platform can read stdin and write stdout. A provider can be a Go binary, a .NET app, a Python script, or even a bash one-liner:

```bash
#!/bin/bash
echo '{"status":"ready"}'
while IFS= read -r line; do
    echo "$line" >> /tmp/debug.log
done
```

This dramatically reduces the friction of writing a provider. With gRPC, every language needs codegen tooling, a gRPC runtime, and protobuf bindings. With stdin/stdout, you need `readline()` and `println()`.

### 2. The I/O analogy

Data flows from a source to a destination. A DStream pipeline is logically analogous to a single process with I/O — the input provider is a reader, the output provider is a writer, and the CLI is the pipe between them.

This maps directly onto fundamental abstractions that every language already has:

| Language | Reader | Writer |
|----------|--------|--------|
| Go | `io.Reader` | `io.Writer` |
| C# | `TextReader` / `Stream` | `TextWriter` / `Stream` |
| Python | `sys.stdin` | `sys.stdout` |
| Bash | `read` | `echo` |

By leaning on these universal abstractions, the conceptual model stays simple: a provider is just a program that reads input and writes output. There is no SDK to learn, no protocol to implement, no codegen to run. The SDK exists to make it *easier*, not to make it *possible*.

### 3. Reduced cognitive overhead

The integration surface between CLI and provider is one line of JSON on stdin (the command envelope) and one line of JSON on stdout (the handshake). After that, it's just lines of JSON flowing through. A developer debugging a pipeline can `cat` a file into a provider and see what comes out. They can `tee` the pipe to inspect data mid-flow. Standard Unix tooling just works.

This simplicity compounds: fewer concepts to learn, fewer things to misconfigure, fewer failure modes to debug.

## Reproducing the benchmarks

```bash
cd /path/to/dstream
go test ./pkg/executor/ -run "^$" -bench Benchmark -benchmem -timeout 120s
```

The benchmarks use real subprocesses (not mocks) to capture actual OS pipe overhead, process startup cost, and memory allocation patterns.
