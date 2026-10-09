# Portable semantic cases

`cases.json` records independently expected outcomes for selected transaction,
parameter, feature-activation, Execute terminal-outcome and LOB-lease invariants.
Embedded `status`, `parameter`, `execution_end` and `reference` messages use
ProtoJSON field/enum names; int64/uint64 values use strings. Other keys are
observation context, not new wire fields.

`known_id` is the queried/affected ID available to the observer; empty means no
ID was delivered to that observer. A NONE no-op describes session activity and
does not establish a historical transaction outcome. For LOB cases `creation`
records the original lease promise; `current` represents a renewed capability.
Lower limits for new creations must not revoke the original lease.

Feature `fields` are fully qualified input fields observed recursively in the
request, including nested parameter/array messages. `request_allowed` is only
the feature preflight decision, not full SQL/authentication validation.
`response_allowed` is the feature's independent output permission, including
structured errors when request validation fails; it does not instruct the server
to emit execution events after rejecting a request. Unspecified advertisements
use `default_advertised`. Dependencies are support requirements, not acceptance.
Execution-end cases supply negotiated policy and advertised warning limits as
observer context; those keys are not fields of ExecutionEnd. With
`partial_results_v1` accepted, warning fixtures include an explicit `role`.
COMPLETE permits only GENERAL warnings, while PARTIAL requires authorization and
at least one retained PARTIAL_RESULT_CAUSE warning. The oracle never derives a
role from stableCode, message text or list position.

Run the reference checks with:

```sh
python3 -B -m unittest discover -s tools/tests -v
```

`tools/conformance.py` is a test oracle for these selected semantic rules. It is
not a complete value codec, RPC handler, security boundary, or high-level SDK.
Go wire tests independently parse embedded messages and check binary round-trips
and selector validity against generated descriptors. Other languages/engine can
reuse these fixtures. Passing them does not certify real cancellation, durable
outcomes, resource ownership/retention, or absence of SQL side effects; those
require the server acceptance scenarios in `docs/acceptance.md`.
