# Kubling gRPC Python client

Generated messages, gRPC stubs and shared feature constants for the official
Kubling protocol. Requires Python 3.10 or later. This package defines no DB-API,
ORM or framework adapter behavior. Server support must be discovered through
capabilities before using optional protocol features.

Distribution name: `kubling-grpc`. The initial package version is `0.1.0`;
availability depends on a completed release to PyPI.

```sh
pip install kubling-grpc
```

```python
from kubling import features
from kubling.v1 import command_pb2, command_pb2_grpc

query = command_pb2_grpc.QueryServiceStub(channel)
```

Install `kubling-grpc[status]` for `grpcio-status` helpers to unpack rich gRPC
status details. `KublingError` itself is always included.

Build from the repository root using Bash, Python 3.10+ and network access:

```sh
python -m pip install -r sdk-python/requirements-build.txt
python -m build sdk-python
python -m twine check sdk-python/dist/*
```

Builds from a checkout regenerate from the canonical protos using the pinned
Buf plugins. Generated code stays out of Git but is included in both the wheel
and sdist. The sdist can produce a wheel without Buf, protoc or a repository
checkout. Installation of either distribution needs no code generation.

The package uses the `kubling` namespace and imports `kubling.v1.*`; applications
must not add `sdk-python/generated` to their path when testing an installed
distribution. See the repository's `docs/releases.md` for publishing.
