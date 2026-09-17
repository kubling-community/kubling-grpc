# Protocol and client releases

Go, Java and Python are official client artifacts generated from the canonical
protocol. Go also retains its existing higher-level helpers for the legacy RPCs.
Generated bindings expose the protocol; they do not imply that an engine
supports every feature. The protocol and official clients follow the
[unified versioning policy](versioning.md): every complete release train uses
one version and one source commit.

## Generated source policy

| Output | Versioned in Git | Published |
|---|---|---|
| Canonical protos and feature registry | Yes | Included in Java/Python packages |
| Go bindings and feature constants | Yes | Go module consumers build directly from Git |
| Java bindings and feature constants | No | Binary, sources and Javadoc JARs |
| Python bindings and feature constants | No | Wheel and sdist |
| Rust feature constants | Yes | No crate or binding pipeline yet |
| Local audit notes and implementation tracking | No | Never included in packages |

Keep private work notes in `.local-notes/`, which is ignored by Git. Public
contract semantics, compatibility rules, acceptance scenarios and executable
fixtures are project documentation/tests rather than implementation tracking.

`generate.sh` accepts `all`, `go`, `java` or `python`. Each language has pinned
Buf plugins; Java and Python outputs are cleaned before generation. Java's
Maven `generate-sources` phase and Python's PEP 517 backend invoke this pipeline.
Python sdists include generated sources and build without external generators.
Java/Python installations from registries require neither Buf nor protoc.

## Protocol releases

The canonical module is `buf.build/kubling/kubling-grpc`. Pushes to `main` that
change the schema update its `main` label. Versioned releases are run manually
from the main branch and publish both `main` and a versioned BSR label
`vMAJOR.MINOR.PATCH`; the matching immutable Git tag is
`proto/vMAJOR.MINOR.PATCH`. The workflow rejects reuse of that tag. Go, Java and
Python publish the same `MAJOR.MINOR.PATCH` from the same release commit.

Before publishing, the workflow builds and lints the module and checks it for
breaking changes against the currently published `main` label. The Go workflow
regenerates its checked-in bindings and fails if they differ from Git. Java and
Python regenerate during their package builds. A protocol change is releasable
only after all three SDK validations pass for the same commit.

| Package | Coordinates | Runtime floor |
|---|---|---|
| Go | `github.com/kubling-community/kubling-grpc/sdk-go` | Go 1.25 |
| Java | `com.kubling:kubling-grpc` | Java 21, protobuf-java 4.36.1, gRPC 1.84.0 |
| Python | `kubling-grpc`, imports `kubling.v1` / `kubling.features` | Python 3.10, protobuf 7.36.1, grpcio 1.84.0 |

Java uses matching generator/runtime versions. Python declares compatible
runtime ranges with these lower bounds; its optional `status` extra installs
the rich-status helper. Update pins and bounds together and run the package
checks whenever generators change.

## CI and local validation

Pull requests and pushes to main regenerate, build and test the Go SDK and build
the Java/Python packages without publishing. A Go SDK tag validates that it is a
semantic version on main, repeats generation and tests, and creates the GitHub
release. Java compiles and runs tests with Oracle GraalVM 25 and produces Java 21
bytecode, following Kubling Core. CI checks both the build JDK and effective
Maven compiler release, then inspects the packaged class versions. The Maven
Wrapper also matches Core's Maven 3.9.0 / Wrapper 3.2.0. Python builds
an sdist and then a wheel from that sdist, and installs/tests both distributions
outside the source tree on Python 3.10 and 3.14. Neither workflow needs publishing
credentials during validation. Both check the contract against `sdk-go/v0.1.1`.

```sh
bash tools/check_java_build.sh
./mvnw --batch-mode --no-transfer-progress -f sdk-java/pom.xml clean verify
python3 tools/check_packages.py java
python -m pip install -r sdk-python/requirements-build.txt
python -m build sdk-python
python -m twine check --strict sdk-python/dist/*
python tools/check_packages.py python
```

Use a fresh virtual environment to install the wheel and run
`python -m unittest discover -s /absolute/path/to/sdk-python/tests -v` from a
directory outside the checkout. Remove stale `sdk-python/dist` outputs before
building a new version. Checks against real engine versions remain a separate
requirement for claiming runtime feature support.

## Registry configuration

Configure these GitHub Actions secrets before publication. Java uses repository
or organization secrets with the same names as Kubling Core; no additional Java
environment is required. Python retains its `pypi` environment.

| Scope | Secrets |
|---|---|
| Java repository / organization | `MAVEN_CENTRAL_USERNAME`, `MAVEN_CENTRAL_PASSWORD`, `MAVEN_GPG_PRIVATE_KEY`, `MAVEN_GPG_PASSPHRASE` |
| `pypi` | `PYPI_API_TOKEN` |

Central credentials are a **Central Portal user token**, not an account password
or an old OSSRH token. The account must own the `com.kubling` namespace. Use an
ASCII-armored signing key whose public key is discoverable by Central. The
Bouncy Castle signer reads the private key and passphrase directly from job
environment variables. The release profile signs the POM and JARs, supplies sources/Javadoc/checksums, and
waits for Central to report the publication complete.

The PyPI token must permit publishing `kubling-grpc`; confirm the project name
is available or owned by the publisher before its first release. The current
workflow accepts that token. Trusted Publishing is an alternative: register
`kubling-community/kubling-grpc`, workflow `python.yml`, environment `pypi` with
PyPI, replace the password input with OIDC, grant `id-token: write` only to the
publish job, and enable attestations. Never store credentials in this repository.

## Check signing credentials before release

Run **Java signing check** manually on the intended branch after configuring the
two GPG secrets. Once the workflow is on main, it can also be launched with:

```sh
gh workflow run java-signing.yml --ref main
```

The check builds and tests with GraalVM 25, runs the same Maven release profile
through `verify`, and signs the POM, binary, sources and Javadoc JARs. It disables
agent passphrase fallback so the configured secrets must work together. A fresh
GPG keyring retrieves the public key from `keyserver.ubuntu.com` and verifies all
four signatures. No package is uploaded and no release tag is required. This
check validates signing credentials and public-key discoverability; registry
token permissions are exercised during publication.

## Publish an approved version

1. Choose one `X.Y.Z` for the complete release train. Set that version in
   `sdk-java/pom.xml` and `sdk-python/pyproject.toml` and update public examples.
2. Run generation, contract checks and every SDK build/test. Merge the exact
   release commit to `main` only after all checks pass.
3. Publish BSR `vX.Y.Z` and create `proto/vX.Y.Z` from the release commit.
4. Create annotated `sdk-go/vX.Y.Z`, `sdk-java/vX.Y.Z` and
   `sdk-python/vX.Y.Z` tags on that same commit.
5. Let the Go tag workflow validate and create its GitHub Release. Verify the
   module resolves through the public Go proxy, for example:

   ```sh
   GOPROXY=https://proxy.golang.org go list -m \
     github.com/kubling-community/kubling-grpc/sdk-go@v1.1.1
   ```

6. Run **Java client** and **Python client** with their existing matching tags.
   Leave `publish` false first for a complete dry validation, then rerun each
   with `publish` true.
7. Verify BSR, the public Go proxy, Maven Central and PyPI independently. Announce
   the train only after every required artifact resolves at `X.Y.Z`.

The Java publish job regenerates, tests, packages and signs the tagged sources
before deployment. Python uploads the exact distributions from the tested build
job. A failed or partly completed publication must be checked in the registry
before rerunning. The coordinated train remains incomplete until all artifacts
are verified; versions and tags must never be replaced or silently skipped.

References: [gRPC Java generation](https://github.com/grpc/grpc-java#generated-code),
[Central requirements](https://central.sonatype.org/publish/requirements/),
[Central publishing plugin](https://central.sonatype.org/publish/publish-portal-maven/),
[PyPA publishing workflow](https://packaging.python.org/en/latest/guides/publishing-package-distribution-releases-using-github-actions-ci-cd-workflows/).
