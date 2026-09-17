# Java and Python releases

Java and Python are generated client bindings with shared feature constants.
They expose the protocol; they do not implement a higher-level SQL API or imply
that an engine supports every feature. The package version and protocol version
are independent. Both new packages start at `0.1.0`.

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

| Package | Coordinates | Runtime floor |
|---|---|---|
| Java | `com.kubling:kubling-grpc` | Java 21, protobuf-java 4.36.1, gRPC 1.84.0 |
| Python | `kubling-grpc`, imports `kubling.v1` / `kubling.features` | Python 3.10, protobuf 7.36.1, grpcio 1.84.0 |

Java uses matching generator/runtime versions. Python declares compatible
runtime ranges with these lower bounds; its optional `status` extra installs
the rich-status helper. Update pins and bounds together and run the package
checks whenever generators change.

## CI and local validation

Pull requests and pushes to main build the packages without publishing. SDK tags
also validate without publishing. Java compiles and runs tests with Oracle
GraalVM 25 and produces Java 21 bytecode, following Kubling Core. CI checks both
the build JDK and effective Maven compiler release, then inspects the packaged
class versions. The Maven Wrapper also matches Core's Maven 3.9.0 / Wrapper 3.2.0.
Python builds
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

## Publish an approved version

1. Set the intended version in `sdk-java/pom.xml` or `sdk-python/pyproject.toml`.
   Update usage examples, validate, review and merge the change to main.
2. Create the matching immutable tag: `sdk-java/v0.1.0` or `sdk-python/v0.1.0`.
   Each language releases independently; a version mismatch rejects publication.
3. Run **Java client** or **Python client** manually from the main workflow,
   passing the existing tag. Leave `publish` false for a complete dry validation.
4. For publication, run with the same tag and `publish` true. The tag must point
   to a commit on main.
5. Verify the artifact on Maven Central/PyPI and install the published version.
   A green local build does not establish a successful registry publication.

The Java publish job regenerates, tests, packages and signs the tagged sources
before deployment. Python uploads the exact distributions from the tested build
job. A failed or partly completed publication must be checked in the registry
before rerunning; versions must never be replaced or silently skipped.

References: [gRPC Java generation](https://github.com/grpc/grpc-java#generated-code),
[Central requirements](https://central.sonatype.org/publish/requirements/),
[Central publishing plugin](https://central.sonatype.org/publish/publish-portal-maven/),
[PyPA publishing workflow](https://packaging.python.org/en/latest/guides/publishing-package-distribution-releases-using-github-actions-ci-cd-workflows/).
