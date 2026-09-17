# Contract acceptance cases

These scenarios define contract conformance. Run them against every advertised
feature and supported SDK/runtime combination.
Verify mutation counts using an independent observer; do not infer absence of
effects from an RPC status. Use deterministic server versions, not `latest`.

| ID | Scenario | Required observation |
|---|---|---|
| C01 | v0.1.1 clients call every existing RPC on new server | Existing behavior preserved, including shared-session transactions |
| C02 | Old Parameter contains only value, including untyped null | Same accepted request/behavior as baseline |
| C03 | New fields cross an old schema reader | Old fields intact; unknown fields never become a typed null |
| C04 | Generated Go/Rust/Python/Java process shared fixtures | Same values, optional presence, event order and structured error fields |
| C05 | New client discovers old GetServerInfo / UNIMPLEMENTED | Generic Execute unavailable; no SQL sent while discovering |
| C06 | Missing generic_execute_v1 and statement kind unknown | Local unsupported error; zero SQL executions |
| C07 | Missing generic_execute_v1 and caller supplies known kind | Exactly one appropriate legacy RPC; no alternate RPC after failure |
| C08 | Execute fails before or after first response | No automatic Query/Exec replay; at most the original submitted execution |
| E01 | SELECT with zero rows | Schema, End(0), ExecutionEnd, gRPC OK |
| E02 | Rows span several batches, exact boundary, final short batch | All rows exactly once, stable schema, matching end count |
| E03 | UPDATE affects zero rows | Present affected_rows=0, distinct from unknown count |
| E04 | Engine cannot determine update count | Explicit unknown alternative, not zero or negative sentinel |
| E05 | Ordered multiple update counts | Order preserved, no invented sum or atomicity |
| E06 | Generated keys: populated, empty with schema, no key columns | Parent association and separate batches; no second execution |
| E07 | Keys requested but unsupported | Rejection before SQL and zero side effects |
| E08 | Procedure with multiple primary results | Monotonic IDs and correct order only with accepted multiple_results_v1 |
| E08b | Additional primary result without acceptance | Explicit non-OK outcome, no silent discard/replay or assumed rollback |
| E09 | Future/unknown event, wrong ID, missing schema/end, wrong row width | Client rejects incomplete/invalid stream; never marks success |
| E10 | Stream fails after some rows or after an update | Partial results remain partial; final error visible, no assumed rollback |
| E11 | ExecutionEnd then transport failure; OK without ExecutionEnd | Client does not report successful complete execution |
| P01 | Typed null for each supported scalar and array type | Declared type retained despite null value |
| P02 | Explicit UNKNOWN type, inconsistent value/type, missing value with type | Structured validation error before SQL |
| P03 | Old server silently ignores declared_type | Client does not rely on typed-null semantics without verified feature |
| V01 | Array null, empty, null element, nested/ragged array | All shapes distinct, element types retained |
| V02 | Mixed non-null element types or excessive array depth | Explicit rejection; no truncation or coercion |
| V03 | Geometry/geography with missing CRS, EPSG SRID, CRS URI | Type distinction and metadata preserved; no implicit reprojection |
| V04 | Contradictory SRID/CRS or invalid WKB | Explicit error; no silent identifier removal |
| V05 | Object or unaccepted new Value on legacy/new paths | Unsupported error, never runtime serialization/string fallback/null |
| V06 | Decimal precision, large integers, local nanosecond timestamp, BMP char | Wire meaning survives every supported language; no numeric/timezone loss |
| L01 | BLOB/CLOB exceeds message limit | Negotiated references/chunks, bounded messages, byte-exact content |
| L02 | Empty LOB, absent size vs size=0, length absent vs length=0 | Presence and completion semantics preserved |
| L03 | CLOB multibyte UTF-8 crosses chunk boundaries | Full text preserved with incremental decoding |
| L04 | Upload gap/overlap/wrong declared size/invalid complete UTF-8 | No published reference; partial upload reclaimed |
| L05 | Read subset, offset at end, offset beyond end, overflow-sized length | Correct contiguous range or explicit invalid-range error |
| L06 | Broken read/upload or lost upload response | No fabricated empty value or successful completion; cleanup bounded by session lifetime |
| L07 | Reference after result/transaction close, expiry, release, session close | Valid for promised lifetime, then explicit unavailable error |
| L08 | Reference from a different session/node | Authorization/affinity enforced; no cross-session data disclosure |
| T01 | Shared session Begin repeated/concurrent | Same active ID or documented pre-effect concurrency rejection, never two hidden transactions |
| T02 | Stale ID after a newer transaction begins | No execution/commit/rollback against the newer transaction |
| T03 | Disconnect before/after commit effects and before response | Client outcome UNKNOWN until authoritative observation; no blind replay |
| T04 | IsInTransaction=false after commit and after rollback | Client does not infer either terminal outcome from this bool |
| T05 | GetTransactionStatus: active, terminal, expired, unrecognized | Known authoritative state or UNKNOWN; never false rollback |
| T06 | Wrong-owner status lookup | No other session's transaction state disclosed |
| T07 | Owning node disappears or session cannot be authenticated | Unknown outcome retained; no fabricated recovery guarantee |
| T08 | Commit/rollback of an active transaction under transaction_ids_v1 succeeds | Nonempty ID and terminal state agree; success=true; mismatched explicit IDs cannot affect a different active transaction |
| A01 | Limits differ across nodes/VDBs or change after discovery | Session/node-bound snapshot honored or rejected before SQL |
| A02 | Batch row/byte/message limits at boundary and boundary+1 | No oversized successful message, no silent row truncation |
| A03 | Client accepts fewer representations than server supports | Only accepted encodings/events emitted |
| A04 | Unknown discovered feature vs unknown requested feature | Ignore discovery extension; reject request extension before SQL |
| A05 | Positive batch/chunk request smaller than server maximum | Client upper bound honored |
| R01 | Error text changes/localizes | stable_code/SQLSTATE/vendor/category drive decisions unchanged |
| R02 | vendor_code absent vs explicitly zero; SQLSTATE absent | Presence preserved in all clients |
| R03 | Application error after partial Execute output | Non-OK status with KublingError; no error event followed by OK |
| R04 | Transport error without details or future unknown detail/enum | No automatic retry or inferred transaction outcome |
| F01 | Constants regenerated for four languages | Names/dependencies match protocol/features.json; drift check passes |
| T09 | Autocommit succeeds | ExecutionEnd + final OK + NONE and empty ID |
| T10 | Autocommit interrupted before outcome is observed | UNKNOWN without invented ID; no assumed rollback/commit |
| T11 | Every TransactionStatus state and ID combination | NONE has empty ID; active/rollback-only/terminal states have nonempty ID; UNKNOWN preserves known ID |
| T12 | Lookup after retention or for an unknown ID | UNKNOWN with exactly the requested ID |
| T13 | Rollback when no transaction is active, repeated or with historical ID | success=true, NONE and empty response ID; no historical outcome inferred |
| T14 | Cuts during Begin, Execute, Commit and Rollback | UNKNOWN unless independent authoritative evidence; preserve every known ID |
| A06 | Input array/spatial/LOB variant or declared_type without accepted_features | Allowed when advertised and valid; does not authorize corresponding output representation |
| A07 | Generated-key flag without acceptance; accepted name without support | Rejected before SQL, sql_executed=false, independently verified zero effects |
| A08 | Advertised structured errors and additive transaction observations without acceptance | Valid on old/new RPCs; never require accepted_features |
| A09 | Capability changes after execution has started | Entire call honors its snapshot; no mid-call rejection solely because snapshot was replaced |
| A10 | Feature registry contains unknown operators, missing activation, ambiguous predicates or dependency cycle | Generator/check fails; selector targets match compiled descriptors |
| L09 | LOB created with capability A, read/released using current B | Same session/node and live lease succeed, even if A expired or new-creation limits decreased |
| L10 | Cross-session or cross-node use of the same LOB | Rejected with no unauthorized data access or release |
| L11 | Reference issued by Execute or WriteLob | Explicit expiry meets guaranteed minimum; absence/short expiry is nonconformant |
| L12 | Resource pressure with outstanding references | Existing leases remain usable; reject new creations instead of early eviction |
| L13 | LOB expiry, release repeated/by wrong owner, session close and node loss | Only permitted invalidations; authorized repeated release succeeds; no distributed recovery promised |
| L14 | Creation at max_lob_bytes and one byte above, including unknown-size upload | Bound enforced before publishing reference; global/session quota exhaustion may reject new creation |
| E12 | DDL/SET without a reliable natural count | UpdateResult with UnknownUpdateCount; no fake zero, prefix classifier or new event |
| P04 | Nested/ragged arrays with recursive descriptors and null elements | Every level described and preserved; incompatible leaf rejected before SQL |
| P05 | Precision on nonnumeric/fixed-width types, zero precision, scale without precision or scale > precision | Pre-SQL validation error and no effects |
| P06 | Exact decimal qualifiers, negative scale and incompatible/rounding-required value | Exact supported values accepted; incompatible value rejected without SQL effects |
| X01 | Cancellation after partial rows, before first row or during fetch | Propagates to engine; statement/result resources closed; no ExecutionEnd for cancelled execution |
| X02 | Deadline while waiting for a slow consumer | Cancellation reaches engine; bounded rows and bytes; no unlimited batch/chunk queue |
| X03 | Statement cancellation with otherwise usable session | Session remains usable; affected transaction outcome UNKNOWN unless independently observed |
| R05 | Pre-SQL capability/affinity/feature/parameter validation failure | Nonempty stable_code and present sql_executed=false; status/error message is not used as proof |
| R06 | Error after partial results | Non-OK, no ExecutionEnd; sql_executed=true if execution is known to have started, otherwise absent |
| R07 | Old status reader receives unknown KublingError Any | Standard gRPC status remains usable without understanding the detail |
| R08 | Infrastructure error without KublingError; absent sql_executed vs false | No inferred retryability, pre-execution guarantee or transaction outcome |

Executable Go wire tests cover descriptor-level backwards/forwards decoding,
typed-field presence, selector targets and rich error round-trips. Portable
semantic fixtures and reference checks are in protocol/conformance and tools;
they cover selected validity and activation decisions, not server enforcement.
They do not certify execution,
failure recovery, flow control, routing or the server conformance cases above.
Buf breaking validates schema evolution, not these operational guarantees.
