# Unified protocol and SDK versioning

Starting with the `1.1.1` release train, Kubling gRPC uses one release version
for the protocol and every official SDK. For a release `X.Y.Z`, users must be
able to select `X.Y.Z` in every supported language and receive bindings
generated from the same canonical contract.

This rule removes the need to infer whether, for example, a Go `0.2.0` client,
a Java `0.1.0` client and a protocol `1.1.0` label describe the same API.

## Release train

The canonical release version is `X.Y.Z`, without a leading `v`. A complete
release train publishes all of the following from one source commit:

| Artifact | Public version or tag |
|---|---|
| Buf Schema Registry module | `buf.build/kubling/kubling-grpc:vX.Y.Z` |
| Protocol Git tag | `proto/vX.Y.Z` |
| Go module and GitHub Release | `sdk-go/vX.Y.Z` |
| Java package and GitHub tag | `com.kubling:kubling-grpc:X.Y.Z`, `sdk-java/vX.Y.Z` |
| Python package and GitHub tag | `kubling-grpc==X.Y.Z`, `sdk-python/vX.Y.Z` |

All Git tags in the train must resolve to the same commit. The BSR source URL,
Java POM, Python project metadata and published package contents must identify
that release. A tag, package or BSR label with a different version does not
belong to the train.

Rust is not currently a published SDK. Its generated shared constants remain
validated in the repository. When an official Rust package is introduced, it
joins the current release train instead of starting an independent `0.x` line.

## What requires a coordinated release

Any change to `proto/` or the normative `protocol/features.json` registry
requires a new release train. The release must regenerate Go, Java, Python and
every other supported SDK, even when a particular generator produces no textual
difference. Every official SDK is then built, tested and published at the same
new version.

A public SDK correction that requires a package release also advances the whole
train. The protocol may be byte-for-byte unchanged, but it receives the same new
BSR label and Git tag so that the equality between public versions is preserved.
Documentation-only and CI-only changes do not require a release when they do not
alter the shipped contract, generated sources or package behavior.

Generated-source storage remains language-specific:

- Go generated sources and feature constants are committed.
- Java and Python sources are generated during their builds and packaged.
- Every release build regenerates from the canonical sources; no SDK consumes a
  copied or independently maintained `.proto` tree.

## Semantic version selection

The release owner chooses the version once for the complete train:

- `MAJOR` changes only for an intentionally incompatible public contract or SDK
  API change.
- `MINOR` changes for additive protocol features or additive public SDK APIs.
- `PATCH` changes for compatible SDK, packaging, generation or release fixes
  that add no protocol feature.

Stable releases use the exact same `X.Y.Z` in every registry. Prereleases remain
disabled until the workflows define and validate one canonical mapping for the
different Maven, Python and Go prerelease syntaxes.

Version alignment identifies the contract carried by an SDK. It does not prove
that a server implements every feature in that contract. Clients still discover
server support through protocol version, feature names, limits, supported types
and affinity requirements. They must continue to follow the mixed-deployment
rules in the client contract.

## Release procedure

1. Choose one `X.Y.Z` for the train. Update the Java and Python package metadata
   and every public installation example to that value.
2. Regenerate all supported SDKs from `proto/` and
   `protocol/features.json`. Commit required generated outputs.
3. Run protocol compatibility checks, feature/conformance tests, every SDK build
   and every SDK test against the exact release commit.
4. Merge the release commit to `main`. Do not create release tags from a branch
   or from different commits.
5. Create the BSR label and the `proto/`, `sdk-go/`, `sdk-java/` and
   `sdk-python/` tags with the same `vX.Y.Z` suffix.
6. Publish the Go release, Maven package and PyPI package from those tags. The
   artifacts must be the outputs validated in step 3.
7. Verify every registry independently: resolve the BSR label, download the Go
   module through the public proxy, consume the Java package from Maven Central
   and install the Python distribution from PyPI.
8. Publish coordinated release notes only after every required artifact and tag
   is available and verified.

Registry publication cannot be atomic. If one publication fails, the train is
incomplete and must not be announced as released. Preserve successful immutable
artifacts and retry only the missing publication after checking that the target
version is still unused. If the same version cannot be completed safely, advance
the entire train to the next patch version; never move a tag, overwrite a package
or silently assign one SDK a different version.

## Release acceptance

A release is complete only when all of these statements are true:

- all release tags exist and resolve to the same commit on `main`;
- BSR exposes `vX.Y.Z` and its source matches the canonical `proto/` directory;
- regenerated SDK outputs have no uncommitted differences;
- Go, Java and Python validation workflows pass for the release commit;
- Go resolves as `github.com/kubling-community/kubling-grpc/sdk-go@vX.Y.Z`;
- Maven Central resolves `com.kubling:kubling-grpc:X.Y.Z`;
- PyPI resolves and installs `kubling-grpc==X.Y.Z`;
- the published SDKs contain the same protocol descriptors and feature registry;
- release notes describe one coordinated `X.Y.Z` train.

## Migration from the independent version lines

The existing `proto/v1.1.0`, `sdk-go/v0.2.0`, `sdk-java/v0.1.0` and
`sdk-python/v0.1.0` tags are historical pre-policy identifiers. They remain
immutable and are not renamed or repointed.

The first planned unified release train is `1.1.1`. It republishes the current
compatible protocol and all official SDKs as `v1.1.1` from one release commit.
After that train, no protocol or official SDK release may advance on its own.
