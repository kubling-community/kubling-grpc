# Kubling client contract: protocol 1.1

These definitions specify the protocol. Clients must discover server capabilities
before using extensions; generated bindings do not establish server support.

The source definitions are under `proto/kubling/v1`. Existing file names, package
names, services, RPC signatures, field types, field numbers and enum numbers
remain intact. `command.proto` retains its public value import. Legacy requests
remain valid; a successful rollback with no active transaction is explicitly
defined as an idempotent no-op, including legacy requests.

## Discovery and mixed deployments

1. Login normally. Preserve returned session identity and `Affinity`, if present.
2. Call `GetServerInfo` with that session's token, using its affinity. An old
   server can ignore the added token field. `UNIMPLEMENTED`, absent capabilities,
   or missing `generic_execute_v1` mean generic execution is unavailable.
   Authentication/transport failures are not successful capability discovery.
3. For Execute require protocol major 1, minor >= 1, `generic_execute_v1`, a
   nonempty `capability_id`, positive request/response/batch limits, supported
   types, and a known affinity requirement. Feature names, not minor versions,
   determine optional support. Unknown feature strings in discovery are ignored.
4. Input fields/variants and RPC selection activate features according to the
   canonical registry. `accepted_features` authorizes output representations
   and multiple results; it is not required to send a typed parameter or array.
   Dependencies must be advertised, not recursively added to accepted_features.
5. Before SQL, the server validates identity, session/VDB/node scope, affinity,
   accepted features, parameters and transaction assertions. Unknown or
   unavailable requested features fail; they must not be silently ignored.

`capability_id` is opaque, not authentication. It binds the effective features,
limits, type support, session, and node. A server must reject an identity it
cannot validate before SQL execution. On a configuration change it either honors
the snapshot for the complete call or rejects it before starting. A snapshot
from node A is not proof that node B can execute the same request.

`routing_token` is an ASCII opaque routing hint echoed in request metadata named
`kubling-affinity`. Node and session IDs are not addresses or credentials. Routing
infrastructure must preserve affinity, or the receiving node must reject before
execution. `UNSPECIFIED` is not equivalent to no affinity. Reconnecting a channel
does not re-create the session or transaction on another node.

Without generic execution, callers may explicitly select an existing RPC only
when they already know the expected statement result kind. Unknown kind fails
locally without sending SQL. A SQL RPC failure, including `UNIMPLEMENTED`, never
triggers automatic replay through a different SQL RPC. This applies even when
no response event was received. This proposal provides no automatic SQL replay.

Legacy Query/Exec never return Value tags 23-26. They retain existing encodings.
New representation input is restricted to Execute and activates by presence;
new representation output requires explicit acceptance. Transaction IDs and declared types may also be used with
legacy RPCs when the corresponding feature is verified on the owning session;
old servers ignoring unknown fields is not sufficient semantic support.

## Canonical feature activation

`protocol/features.json` schema version 2 is normative. `activation.request`
describes an input trigger, `request_requires_acceptance` adds an acceptance
requirement to that trigger, and `activation.response` grants output permission.
Every triggered, explicitly accepted, or emitted feature must be advertised with
all its `requires` dependencies. Dependencies describe server support, not client
output acceptance. All known request incompatibilities are rejected before SQL.

Expressions contain one operator: `rpc` selects a full gRPC method; `field` plus
`test` selects a fully qualified field in any recursively visited input message;
`present` tests protobuf presence, `nonempty` tests a nonempty string and `true`
tests a true bool. `accepted_features: true` tests this feature's name in the
Execute request; `advertised: true` permits additive output independent of that
list. `all_of`/`any_of` combine expressions. A null expression has no activation.
Fields inside nested/ragged arrays must also be visited. `accepted_features` is
empty on RPCs that do not define it. Duplicate/unknown requested names fail.

| Feature | Request activation | Response permission |
|---|---|---|
| generic_execute_v1 | Execute RPC | Execute response stream |
| typed_parameters_v1 | Parameter.declared_type presence | No output acceptance needed; Execute column descriptors are part of generic execution |
| structured_errors_v1 | None | Advertised details may always accompany errors, including old RPCs |
| transaction_ids_v1 | Nonempty transaction_id | Advertised additive response IDs |
| transaction_status_v1 | GetTransactionStatus RPC | Advertised additive status fields |
| array_values_v1 | array_value input variant | Execute and accepted feature |
| spatial_values_v1 | geometry_with_crs/geography_with_crs input variant | Execute and accepted feature |
| lob_read_v1 | ReadLob or ReleaseLob RPC | Execute LobReference only with acceptance |
| lob_write_v1 | WriteLob RPC or lob_reference input variant | WriteLob response reference, implicit in selecting the upload RPC |
| multiple_results_v1 | accepted_features | Execute and accepted feature |
| generated_keys_v1 | return_generated_keys=true AND explicit acceptance | Same flag and acceptance on Execute |

Input arrays/spatial values/LOB references do not authorize the same output
encoding. Conversely, output acceptance alone does not fabricate input data.
An advertised structured error or additive transaction observation never depends
on accepted_features. Generated-key acceptance without the request flag produces
no keys; a true flag without acceptance fails before SQL. The legacy
ExecRequest.returnGeneratedKeys field keeps its old behavior.

TransactionStatus is a shared representation: the required ExecutionEnd status
belongs to generic_execute_v1, and an available transaction observation can be
part of structured_errors_v1. Neither implies support for GetTransactionStatus
or its retention guarantees; those specifically require transaction_status_v1.

## Execute sequence

Every RPC executes one submitted SQL statement once. No statement-kind probing
or Execute-to-Query/Exec fallback occurs. This says nothing about exactly-once
effects across transport retries, crashes, source failures, or new RPC calls.

The successful event grammar is:

```text
execution   := primary_result+ ExecutionEnd grpc-OK
primary_result := query_result | update_result
query_result := ResultSetStart(QUERY) ResultRows* ResultSetEnd
update_result := UpdateResult generated_keys?
generated_keys := ResultSetStart(GENERATED_KEYS, parent_result_id)
                  ResultRows* ResultSetEnd
```

- Without accepted `multiple_results_v1`, exactly one primary result is allowed.
  Generated keys do not count as another primary result.
  If SQL unexpectedly produces more, report an unsupported non-OK outcome
  instead of silently discarding results. Reject before SQL when this is known
  in advance; otherwise report the failure after the original execution, without
  implying that its effects were rolled back.
- Every result, including generated keys, gets a consecutive ID starting at 1.
  Results never interleave. All rows and the end refer to the currently open ID.
- Start supplies the entire schema before any row. Each column has a concrete
  `declared_type`; existing `data_type`, precision and scale remain populated
  consistently. No schema changes are allowed within a result. Duplicate column
  names are legal: values are positional.
- A zero-row result is Start(schema), End(row_count=0), then the execution end.
  `ResultRows` contains at least one row. Every row has exactly the schema width.
- `ResultSetEnd.row_count` equals the sum of rows across its batches.
- `UpdateResult` is complete by itself, contains a nonempty ordered list of
  counts, and needs no ResultSetEnd. Each count has exactly one alternative:
  a nonnegative affected-row count (including zero), or an explicit unknown.
  SQL errors use final non-OK status, never negative sentinel counts. No batch
  SQL input API or atomicity promise is introduced by the list of counts.
  DDL, SET and equivalent commands without a natural row set use this result
  too: preserve a real reported count or emit UnknownUpdateCount when no reliable
  count exists. Do not invent zero or classify SQL by its text prefix.
- Generated keys require `generated_keys_v1` in the accepted set and
  `return_generated_keys=true`. They immediately follow their update and refer
  to its ID. If there are no key columns, omit the key result; if key columns
  exist but no keys were produced, send its schema and zero-row end.
- Unsupported key requests fail before executing SQL. Never run the statement
  a second time to obtain keys.
- `ExecutionEnd` occurs exactly once, last; `result_count` counts all result IDs,
  including generated keys. Its transaction observation is required, with
  UNKNOWN when no authoritative observation is available.
- A successful transport close without ExecutionEnd is an incomplete protocol
  result. ExecutionEnd without final gRPC OK is not successful completion.
- Unknown event alternatives, missing event alternatives, unexpected ordering,
  wrong IDs, or inconsistent counts fail the client stream validation. They must
  not be interpreted as null, ignored updates, or successful completion.

Any non-OK status terminates the sequence, possibly after partial rows/results.
Already delivered data does not imply successful overall execution or rollback
of earlier effects. Cancellation/deadline expiry has the same uncertainty.
Future output parameters or event variants need their own negotiated feature.
The grammar prepares multiple results without advertising engine support today.

## Cancellation, deadlines and backpressure

The server must propagate RPC cancellation and deadline expiry to the running
engine work, cancel the statement/request, and close ResultSet, Statement and
execution resources on every exit path. Merely stopping event emission/reads is
insufficient. Cleanup must also release resources allocated before the first row.
Cancelling a statement does not automatically close its session unless the
session is unusable. An affected transaction outcome remains UNKNOWN unless
independent authoritative evidence establishes it; preserve any issued ID.

A cancelled/failed execution emits no ExecutionEnd and terminates non-OK when
the transport permits. If completion was already emitted before cancellation
became observable, it cannot be retracted; the client still requires final gRPC
OK. This does not convert partial results into a successful complete result.

Server buffering must have simultaneous finite row and byte bounds, including
queued batches/chunks. A slow consumer must apply backpressure to engine fetch
and LOB reads, not cause unlimited accumulation. Cancellation must work while
waiting for downstream demand. The bounds apply to streaming implementation,
not only to the size of an individual serialized message. The existing buffered
Go Query helper is unchanged and does not demonstrate these server guarantees.

## Types, parameters, and message limits

`Parameter.value` keeps tag 1. An absent `declared_type` preserves legacy inference
and untyped null. With `typed_parameters_v1`, an explicit descriptor must be
valid and non-UNKNOWN. For typed null, send `null_value` plus that descriptor.
An omitted Value or unset kind is not a typed null. Explicit types and non-null
values must agree; no silent type coercion is introduced here. New fields may
not make requests that were valid in legacy mode invalid merely by being absent.

`TypeDescriptor.element_type` is present exactly for ARRAY and recursively
describes nested arrays. `precision` and `scale` are optional: absence is not
zero. The complete qualifier rules for this revision are:

| Type | precision | scale |
|---|---|---|
| BIGINTEGER | Optional, positive count of decimal digits | Forbidden |
| BIGDECIMAL | Optional, positive count of decimal digits | Optional signed decimal scale; requires precision and must be <= precision |
| ARRAY and all other types | Forbidden | Forbidden |

Negative decimal scales are permitted (e.g. scale -2 denotes multiples of 100).
Numeric values must fit explicit qualifiers exactly, without rounding or silent
coercion; decimal padding may preserve the same exact value. An unknown enum
number or VALUE_TYPE_UNKNOWN in an explicit descriptor is invalid. The old
Column.precision/scale fields retain their metadata meaning (including lengths
for nonnumeric types); do not blindly copy them into TypeDescriptor qualifiers.
Servers reject invalid descriptors or incompatible values before executing SQL.
The descriptor is mandatory in every Execute column, including empty results.

ArrayValue contains its element descriptor even when empty. Null array, empty
array, and array containing null are different values. All non-null elements
match the descriptor; nested/ragged arrays are permitted only within the
advertised `max_array_dimensions`. Parameter, schema and value descriptors must
agree. Servers advertising `array_values_v1` publish a positive depth limit.

SpatialValue carries plain WKB and optional CRS information. Geometry and
geography remain distinct oneof alternatives. Absent SRID/CRS means unknown;
SRID zero does not encode absence. SRID-only values identify an EPSG code. A
CRS is authority-qualified or an absolute URI. Both identifiers must agree when
present together; unresolved/contradictory identifiers fail explicitly. No
implicit reprojection, coordinate reordering or EPSG:4326 assumption is made.

Object has no wire representation and is explicitly unsupported. Do not encode
arbitrary objects using Java serialization, pickle, undocumented bytes, JSON
fallbacks, or google.protobuf.Any. Known JSON values retain their own type.

Legacy numeric, decimal, temporal and binary encodings are unchanged. Decimal
strings retain exact decimal values on the wire; converting them to a binary
float is an SDK choice, not a protocol rule. Timestamps remain local timestamp
strings, not implicit UTC instants. CHAR remains a single non-surrogate BMP
Unicode scalar.

Request/response limits apply to each complete uncompressed protobuf message;
HTTP/2 framing is excluded. `max_batch_bytes` includes the complete ResultRows
ExecuteResponse envelope, not just payload cells. A ResultRows message must meet
both batch and response limits. Positive request batch_size is an upper bound,
also bounded by server limits. Other events, including a large schema, must fit
the response limit. Unknown limits are zero, never unlimited. An oversized row
cannot be silently split or truncated: use negotiated LOB references for its
LOB cells, or fail explicitly. Parameters that exceed the request limit require
a different supported representation before SQL is sent, not a retry after SQL.

## LOB lifecycle

`lob_read_v1` enables output LobReference, ReadLob and ReleaseLob.
`lob_write_v1` additionally enables WriteLob and input references; it depends on
lob_read_v1. Both the appropriate type direction and feature must be supported.

- References identify immutable BLOB bytes or UTF-8 CLOB bytes in one authorized
  session. They survive statement and transaction completion, but not session
  closure, explicit release, declared expiry or node loss. All access honors
  affinity; no distributed reference recovery is promised.
- ReadLob, ReleaseLob and LobWriteStart validate a current compatible capability
  identity for the owning session/node. A reference created with snapshot A can
  be read/released with snapshot B. The reference is not bound to A, and replacing
  or invalidating A does not invalidate the reference. A is not required later.
- `size_bytes`, if present, is exact; absence is unknown, not an empty LOB.
  Every reference issued by Execute or WriteLob must have expires_at_unix_ms;
  wire optionality permits presence detection, not an unbounded/missing lease.
  At issue time, expiry must be at least the creation time plus the advertised
  positive lob_reference_retention_seconds. Only expiry, release, session closure
  or node loss can invalidate an issued reference. Resource pressure cannot evict
  it early; reject new creations rather than breaking existing leases.
- max_lob_bytes bounds each newly created LOB's payload (UTF-8 bytes for CLOB).
  Both it and lob_reference_retention_seconds must be positive with LOB support.
  Unknown-size uploads are checked incrementally before publication. A lower
  limit/retention in snapshot B applies to new references, never to existing
  leases. New chunk/message bounds can be honored by changing chunk sizes.
  Global/per-session quotas may return RESOURCE_EXHAUSTED; advertised limits
  do not reserve capacity for future creations. Read/release support must remain
  available for existing leases even if creation is disabled by configuration.
- Offsets, sizes and chunk limits are byte counts, including for CLOB. Chunk
  boundaries may split a UTF-8 sequence; consumers decode incrementally. The
  complete CLOB must be valid UTF-8. Arbitrary byte-range reads need not be valid
  standalone text.
- Read offsets must be <= size. An omitted length reads to the end; an explicit
  zero returns an empty completed range. A requested range past the end is
  shortened to the remaining bytes without arithmetic overflow. Responses are
  contiguous from the requested offset, with no gaps/overlap. Only the last
  response marks end_of_read; empty reads still emit that final response.
- Every chunk respects both max_lob_chunk_bytes and the message limit including
  its envelope. Positive client max_chunk_bytes can reduce the payload bound.
- Upload starts with exactly one LobWriteStart, followed by contiguous nonempty
  chunks from offset zero; empty LOBs need only Start. The client half-closes to
  finish. Declared size, if present, must match; only then may a reference be
  returned. No resume/replay guarantee is offered. Partial uploads are discarded
  or reclaimed by the server. The server must bound upload resources and may
  reject resource exhaustion without publishing a reference.
- Release invalidates the reference; repeating release by its authorized owner
  is a successful no-op, including an already expired reference. Ownership and
  session authorization are still checked; no other session's reference can be
  invalidated. IDs must not become credentials or expose another user's data.
  Uploads whose reply is lost are reclaimed no later than session closure.
- Reading a missing, released or expired reference fails explicitly. A broken
  read is incomplete; never substitute empty bytes or null. Large inline legacy
  LOBs do not silently become references in legacy responses.

## Transactions and uncertainty

Legacy transaction semantics are session-based, including clients sharing a
token. This is preserved. Nonempty transaction_id adds an assertion, not a new
authentication method: the ID must be active, belong to this authorized session,
and be usable on this node. IDs are opaque and never reused. Empty IDs preserve
legacy behavior. A stale explicit ID cannot operate on a newer transaction.

With `transaction_ids_v1`, Begin returns an ID even if the underlying database
defers actual transaction start. A repeated Begin on a session already in a
transaction returns the same active ID, preserving the shared-session model.
Concurrent begin/commit/rollback/SQL on a session must be serialized or rejected
before effects; this contract does not promise concurrent statements on one
session. Terminal responses describe the identified transaction. A successful
new Commit response with an explicit ID sets success=true and COMMITTED;
Rollback of an active transaction reports ROLLED_BACK with its issued ID. When
no transaction is active, Rollback is a successful idempotent no-op: success=true,
NONE and an empty response ID. This also applies to legacy requests and repeated
rollback. An explicit historical request ID does not turn this no-op into proof
that the historical transaction rolled back; use GetTransactionStatus for that
ID. If another transaction is active, a mismatched explicit ID must fail without
affecting it. Checking and acting must be atomic with respect to session work.
Reading GetServerInfo does not switch legacy request behavior.

TransactionState meanings:

| State | Meaning |
|---|---|
| UNKNOWN | Preserve the queried/affected ID when available; empty only if no ID was ever issued |
| NONE | No active transaction; transaction_id must be empty |
| ACTIVE | Active observation; transaction_id must be nonempty |
| ROLLBACK_ONLY | Cannot commit successfully; nonempty transaction_id |
| COMMITTED | Authoritative terminal commit; nonempty transaction_id |
| ROLLED_BACK | Authoritative terminal rollback; nonempty transaction_id |

These invariants apply wherever TransactionStatus occurs, including error
details. An optional absent status remains unavailable, not an implicit NONE.
If an ID was issued, the server cannot discard it when emitting UNKNOWN. An
interrupted autocommit for which no ID was issued may have UNKNOWN with empty ID.
A client that never received an ID (for example because the Begin reply was
lost) can only keep a local UNKNOWN observation without an ID; it must not invent
one or assume no transaction exists. Once an ID is known it must be preserved.
Old servers without ID support
cannot establish new ID-bearing observations through an old active/success bool.

Successful autocommit requires ExecutionEnd, final gRPC OK, and a transaction
observation NONE with empty ID. It must not fabricate a COMMITTED transaction
ID for an implicit autocommit unit. This does not imply autocommit work is
recoverable through GetTransactionStatus.

On a cut/deadline during Begin, Execute, Commit or Rollback, the client marks the affected
outcome UNKNOWN unless it already has independent authoritative evidence. A
previous ACTIVE observation cannot establish the current outcome. gRPC status
UNKNOWN and TransactionState.UNKNOWN are different concepts. No automatic replay
is authorized. The detail message may be absent precisely when it is most needed.

`transaction_status_v1` requires GetTransactionStatus and a positive minimum
terminal-outcome retention in capabilities. The lookup is authenticated and
uses the same affinity and requires a nonempty input ID. Return the exact requested
ID with a known state or UNKNOWN; an unrecognized/expired record must not return
NONE/ROLLED_BACK or discard that ID. Wrong-owner IDs
must not disclose another session's state. Non-OK transport/authorization errors
leave the client's outcome unknown. Retention is not a promise of recovery after
node loss. IsInTransaction=false only describes session activity and never
substitutes for an outcome lookup. A known transaction outcome also does not
reconstruct missing statement rows or prove a particular statement ran.

### operation_id evaluation (deferred)

The protocol does not provide an operation idempotency guarantee. A future operation_id requires:

1. A defined authenticated session/principal and method namespace, not merely a
   caller-generated string.
2. Atomic registration with execution, payload identity rules, concurrent
   duplicate handling, and rejection of the same ID with different content.
3. A specified durability/retention window and scope across node failure.
4. Handling of the gap between remote source effects and the local outcome log.
5. Explicit retention/replay rules for rows, multiple results and generated keys,
   or a status-only outcome lookup that cannot replay those data.
6. Reconciliation semantics for a Begin reply lost before its ID is received.

The specified no-active-transaction rollback no-op does not imply idempotent
execution or commit replay. GetTransactionStatus improves observation; it does not itself provide any
of these execution guarantees. No operation_id field or feature is advertised
until these obligations can be implemented and tested.

## Structured errors

When `structured_errors_v1` is advertised, application errors attach exactly one KublingError
inside google.rpc.Status.details, encoded in grpc-status-details-bin. Its type
URL is `type.googleapis.com/kubling.v1.KublingError`. Standard gRPC status remains
authoritative for transport handling. Infrastructure-generated errors may lack
details, and unknown Any types must be tolerated by old and new clients. This
feature never requires accepted_features, including on Query/Exec and errors
raised before an Execute request's negotiation succeeds.

- stable_code is mandatory and nonempty for errors originated/classified by
  Kubling; diagnostic text can change without changing the stable identifier.
- sql_state, if present, is a valid, stable SQLSTATE matching [0-9A-Z]{5} and
  identifying the same SQL condition across messages/locales. Omit an unknown
  SQLSTATE rather than invent one. vendor_code may be absent
  or explicitly zero; these cases are distinct.
- Missing/unknown category or retryability implies no automatic retry. SAFE is
  only allowed when the server can certify replay cannot duplicate effects.
  RESTART_TRANSACTION applies only with an established safe transaction outcome;
  CHECK_TRANSACTION_STATUS is advice to observe, not to repeat SQL.
- Transaction observations must be authoritative. Missing details/status or
  UNKNOWN cannot be inferred from exception text, bool success, or inactivity.
- sql_executed=false is required on pre-SQL validation errors, including feature,
  capability, affinity and parameter failures. It certifies SQL never started.
  True means SQL started, not that effects committed or execution completed.
  Absent means unknown/not applicable, never false. This observation alone does
  not authorize an automatic replay or a different SQL RPC.
- Partial results are followed by non-OK status and no ExecutionEnd, not an error event plus OK.
  Details should remain small and must not contain rows, parameters, credentials
  or raw source exception payloads.

Initial protocol stable codes (engine-specific SQL failures may have other
documented stable codes):

| stable_code | gRPC status | Use |
|---|---|---|
| KBL_CAPABILITY_MISMATCH | FAILED_PRECONDITION | Missing/stale/wrong-scope capability identity, before SQL |
| KBL_FEATURE_UNSUPPORTED | UNIMPLEMENTED | Requested extension is unavailable, before SQL |
| KBL_AFFINITY_MISMATCH | FAILED_PRECONDITION | Wrong routing scope, before SQL |
| KBL_PARAMETER_TYPE_MISMATCH | INVALID_ARGUMENT | Invalid descriptor/value agreement, before SQL |
| KBL_TRANSACTION_MISMATCH | FAILED_PRECONDITION | Explicit ID does not identify the active owned transaction |
| KBL_VALUE_UNSUPPORTED | UNIMPLEMENTED | Object or an unrepresentable/unaccepted value |
| KBL_LIMIT_EXCEEDED | RESOURCE_EXHAUSTED | Message, batch, depth or resource bound exceeded |
| KBL_LOB_NOT_FOUND | NOT_FOUND | Owned reference is unavailable or expired |
| KBL_LOB_INVALID | INVALID_ARGUMENT | Invalid range, offset sequence, type, size or UTF-8 |

The status/code alone never certifies replay safety. Some failures, such as an
unrepresentable result or exceeded result size, can occur after execution.

## Features and acceptance

`protocol/features.json` is the canonical feature-name/dependency/activation registry.
`tools/generate_features.py` produces constants for Go, Rust, Python and Java;
`--check` detects stale generated files. Features are strings on the wire, never
language-specific enum ordinals. Registry membership does not advertise support.
A server advertises a feature only after its applicable acceptance cases pass.

See [language compatibility](compatibility.md) and [acceptance cases](acceptance.md).
