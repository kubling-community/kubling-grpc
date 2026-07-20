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

## Official SDKs

| Language | Status                          |
|----------|---------------------------------|
| Go       | ✅ Official                      |
| Java     | Generated from Protocol Buffers |
| Python   | Generated from Protocol Buffers |
| Rust     | Generated from Protocol Buffers |

The Go SDK is the primary maintained SDK and provides a high-level API that abstracts the underlying gRPC protocol.

See:

```
sdk-go/README.md
```

## Code Generation

This repository uses Buf.

Generate all bindings from the repository root:

```bash
./scripts/generate.sh
```

The bootstrap script installs the required tooling automatically.

## License

Apache License 2.0