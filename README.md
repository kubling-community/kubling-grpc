# Kubling gRPC

The official language-neutral communication protocol for Kubling.

Kubling is a federated SQL engine that provides a unified view across whatever datasource that can be represented as tuples.

Kubling gRPC provides a modern, language-neutral protocol designed for cloud-native applications, infrastructure automation, edge computing, and AI-driven operational systems.

This repository contains the official Protocol Buffer definitions together with the reference SDKs maintained by the Kubling project.

## Why gRPC?

Kubling gRPC extends the platform to any language supporting Protocol Buffers and gRPC, allowing applications to execute SQL queries, perform transactional operations, and interact with Kubling using a consistent API regardless of the implementation language.

## Features

- SQL query execution
- Streaming result sets
- Transactions
- Session management
- Strongly typed values
- Generated keys
- TLS and plaintext connectivity
- Language-neutral Protocol Buffers contract

## Protocol

The Protocol Buffers definitions are the source of truth.

```
proto/kubling/v1/command.proto
```

Any language supporting Protocol Buffers can generate client stubs directly from this file.

The additive protocol 1.1 is described in [the client contract](docs/client-contract-v1.md),
with [compatibility rules](docs/compatibility.md) and [acceptance cases](docs/acceptance.md).
Generated definitions do not imply server support: clients must verify capabilities.
Released schemas are available from
[`buf.build/kubling/kubling-grpc`](https://buf.build/kubling/kubling-grpc).

## Official SDKs

| Language | Status                          |
|----------|---------------------------------|
| Go       | ✅ Official                      |
| Java     | Generated client; Maven packaging and release workflow |
| Python   | Generated client; wheel/sdist packaging and release workflow |
| Rust     | Binding generation planned      |

The Go SDK is the primary maintained SDK and provides a high-level API that abstracts the underlying gRPC protocol.

Java and Python expose generated messages, stubs and feature constants. See the
[Java client](sdk-java/README.md), [Python client](sdk-python/README.md) and
[build/publication guide](docs/releases.md). Package availability is established
by a completed registry release, not by the presence of a workflow.

See:

```
sdk-go/README.md
```

## Code Generation

This repository uses Buf.

Generate all bindings from the repository root:

```bash
./generate.sh
```

Use `./generate.sh go`, `java` or `python` to generate a single language. Direct
`buf generate` uses the root Go template. Java and Python builds run generation
automatically and package the resulting sources; those outputs are ignored by
Git. Go generated code remains versioned for module consumers.

The bootstrap script installs the required tooling automatically.

Feature names are defined once in `protocol/features.json`. The generation script
also produces shared constants for Go, Rust, Python and Java (requires Python 3).
Check previously generated constants with `python3 tools/generate_features.py --check`;
use `--language go` (or `rust`, `java`, `python`) to check just one language.
The registry also defines request activation and output acceptance. Portable
semantic cases are in [protocol/conformance](protocol/conformance/README.md);
run their reference checks with `python3 -B -m unittest discover -s tools/tests -v`.

Protocol releases use BSR labels such as `v1.1.0` and matching immutable Git
tags such as `proto/v1.1.0`. Package and SDK versions have independent release
cycles.

## License

Apache License 2.0
