# Kubling gRPC repository guidance

## Project overview

Kubling gRPC is the canonical language-neutral client contract for Kubling. The
repository contains Protocol Buffer definitions, capability and conformance
metadata, generated bindings, official client packages and coordinated release
automation.

## Contract and generated sources

- `proto/kubling/v1/` is the canonical Protocol Buffer source.
- `protocol/features.json` is the canonical shared feature registry.
- Preserve published field numbers, enum values, RPC names and wire semantics.
  Contract evolution is additive unless an incompatible release is explicitly
  approved.
- Keep existing RPCs wire-compatible and gate negotiated behavior through
  declared capabilities.
- Do not maintain an authoritative protocol copy in an SDK or consuming repo.
- Never edit generated bindings or feature constants manually.
- Generate coordinated outputs with `./generate.sh all` and inspect the diff.
- Commit Go bindings and Go/Rust feature constants. Java and Python generated
  sources are build outputs and remain untracked.
- A protocol or feature change must regenerate and validate every maintained
  SDK, including package contents and committed generated sources.

## Versioning and releases

Use the version in `VERSION` for the protocol and every official SDK. Keep Java
and Python metadata and public installation examples aligned with it.

A change to `proto/` or `protocol/features.json` requires one coordinated Buf,
Go, Java and Python release train from the same commit. Do not advance an
official SDK independently, reference an unpublished protocol version from a
consumer, move published tags or replace registry artifacts.

## Validation

Inspect `git status` before editing and preserve unrelated local work. Run the
checks relevant to the change. Protocol and feature changes require at least:

```sh
./tools/bin/buf format --diff --exit-code
./tools/bin/buf lint
./tools/bin/buf build
./generate.sh all
python3 tools/generate_features.py --check
python3 -B -m unittest discover -s tools/tests -v
```

Also check compatibility against the latest released `proto/v*` tag, run Go
tests, build and test Java with the required JDK, and build, inspect and
install-test the Python wheel and sdist. Follow `docs/releases.md` for current
release checks and acceptance criteria.

## Repository hygiene

- Treat tracked files, commits, branches, pull requests, workflow logs and
  release artifacts as public.
- Track only durable documentation intended for users and contributors. Keep
  chats, prompts, handoffs, plans, tracking and investigation notes in
  `.local-notes/`, which is ignored by Git, or outside the checkout.
- Do not expose secret values, customer data, private infrastructure,
  non-public endpoints or unsanitized logs. Documented secret names are allowed;
  their values belong only in the configured secret store.
- Use portable placeholders and repository-relative paths instead of personal
  paths, account identifiers or machine-specific configuration.
- Review new documentation, examples, fixtures and copied command output for
  private context before committing them.

## Scope and approval

- Prefer small, reviewable changes with focused validation.
- Do not alter engines, providers or downstream protocol copies unless that
  scope is separately requested.
- Commit and push require explicit user approval.
- Pull requests, merges, tags, releases and artifact publication each require
  explicit approval. Validation or a dry run does not authorize publication.
