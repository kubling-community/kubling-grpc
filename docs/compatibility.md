# Client compatibility rules

The contract remains proto3/package `kubling.v1`. Binary wire compatibility,
generated-source compatibility, runtime compatibility and behavioral compatibility
are separate checks. An additive schema alone proves only part of the contract.

## Rules shared by all clients

- Do not rename/move old messages, packages, files, RPCs or Java outer classes;
  preserve the public import of value.proto and the current go_package path.
- Do not change or reuse field/enum numbers, move old fields into a oneof, remove
  deprecated xml_value, change RPC streaming cardinality, or change legacy defaults.
- Use field presence for declared types, optional SQLSTATE/vendor codes,
  sql_executed, precision/scale, LOB length/size/expiry and parent result IDs.
  Zero/false is not absence; absent sql_executed is not a pre-execution guarantee.
  LOB expiry is semantically required even though it has wire presence.
- An unrecognized oneof is not null. Fail unsupported required events/values;
  do not continue with incomplete data. Do not emit new values through old RPCs.
- Tolerate unknown discovered feature names and error detail types. Do not accept
  unknown execution extensions, transaction outcomes or retry advice as success.
  Advertised structured errors/additive transaction observations do not require
  accepted_features; input variants do not imply output acceptance. Use the
  activation rules in protocol/features.json rather than language-specific rules.
- Keep exact big integer/decimal strings and int64/uint64 values; do not route
  them through JSON numbers or binary floats. Preserve local timestamp semantics.
- GetServerInfo failure/missing data never proves feature support. No SQL RPC
  replay is allowed to classify a statement. gRPC transparent transport retry is
  not a public statement idempotency guarantee; do not configure application
  retries for Execute/Exec/transaction finalization based only on status codes.
- Keep old high-level APIs available. New streaming consumption is a separate
  API; changing existing buffered Go helpers into lazy readers is out of scope.
- Pin/test code generators and supported runtimes before publishing each SDK.
  Generators are pinned in the root and language-specific buf.gen.yaml files.
  Java/Python manifests align runtime dependencies with generation. Rust binding
  generation remains a prerequisite for publishing a Rust SDK.
- Feature constant files are generated from the shared registry. Including them
  in a language distribution must not imply that all handlers exist on a server.

## Language-specific rules

| Language | Rules |
|---|---|
| Go | Check descriptor pointer presence; inspect oneof concrete alternatives and include an unsupported default. Enum casts may contain unknown integers. Adding RPCs grows generated interfaces: mocks/adapters implementing them must be rebuilt. Server implementations should embed UnimplementedQueryServiceServer. Preserve existing imports and helpers. Generated feature constants are in sdk-go/features. |
| Rust | Establish/pin a binding pipeline before claiming an official SDK build. If using prost/tonic, presence uses Option; enum fields may contain unknown i32 values and oneof matches need an unknown/unsupported path. Adding variants can break exhaustive matches; adding fields can break struct literals, so use Default-based construction or stable wrapper builders. Do not rely on unknown-field forwarding across decode/reencode without a tested runtime guarantee. features.rs is standalone generated source, not a published crate. |
| Python | Use HasField for explicit presence and WhichOneof for variant selection, not truthiness (zero/empty is valid). Keep generated module import paths and package layout consistent. Align protobuf/grpcio with generated code; decode KublingError from rich gRPC status details. Constants are supplied in kubling.features; no DB-API behavior is defined. |
| Java | Keep com.kubling.transport.grpc and existing outer class names. Use hasX and case discriminators, with unknown/default handling including UNRECOGNIZED; regenerating exhaustive switch expressions may require source changes. Align protoc-generated code and runtime dependencies before consumption; never combine newer gencode with an older runtime. Java long carries protobuf uint64 bits, so handle those fields as unsigned where needed. Features.java must be included in the eventual artifact. |

Java/Python generated outputs are ignored by Git and included in their packages.
CI builds Java JARs and Python wheel/sdist distributions and checks their contents
and generated APIs. Rust has no binding pipeline. See [release requirements](releases.md)
for the supported runtime matrix and registry configuration. Package checks do
not establish real-engine support for protocol features.

## Compatibility matrix to execute before release

| Client | Server | Expected |
|---|---|---|
| v0.1.1 generated client | New engine | Existing RPCs, field semantics, shared sessions and old value encodings work |
| New client | Old engine | No generic execution without discovery; known-kind calls can explicitly use legacy RPCs |
| New client | New engine with extension disabled | Fail unsupported requests before SQL where inputs reveal the requirement; no silent fallback |
| New client | Mixed nodes | Capability/affinity mismatch rejected before effects; never replay SQL on another RPC |
| Old schema reader | New binary message | Added fields do not corrupt old fields; new value alternatives remain unsupported |
| New schema reader | Old binary message | Absent fields remain absent/default; no invented type, outcome, size or capabilities |

Repository checks: buf lint/build; buf breaking against sdk-go/v0.1.1; generated
feature drift/schema checks; portable semantic cases; Go build/unit tests and
legacy descriptor wire tests. Run semantic checks with
`python3 -B -m unittest discover -s tools/tests -v`.
Cross-language execution and real-engine acceptance remain release gates, not
results implied by these checks.

References: [Protobuf schema evolution](https://protobuf.dev/programming-guides/proto3/#updating),
[runtime compatibility](https://protobuf.dev/support/cross-version-runtime-guarantee/),
[gRPC errors](https://grpc.io/docs/guides/error/#richer-error-model).
