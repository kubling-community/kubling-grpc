#!/usr/bin/env python3
"""Verify release signatures using a public key retrieved from a keyserver."""

from pathlib import Path
import subprocess
import tempfile

from check_release import package_version

ROOT = Path(__file__).resolve().parents[1]


def main():
    name = f"kubling-grpc-{package_version('java')}"
    target = ROOT / "sdk-java/target"
    artifacts = [target / (name + suffix) for suffix in
                 (".pom", ".jar", "-sources.jar", "-javadoc.jar")]
    fingerprints = set()
    # An empty keyring tests public-key discoverability and cannot use a local
    # private key, cached passphrase or pre-existing trust configuration.
    with tempfile.TemporaryDirectory(prefix="kubling-signature-check-") as keyring:
        for artifact in artifacts:
            signature = Path(str(artifact) + ".asc")
            if not artifact.is_file() or not signature.is_file():
                raise SystemExit(f"Missing artifact or signature: {artifact.name}")
            try:
                result = subprocess.run(
                    ["gpg", "--batch", "--homedir", keyring, "--status-fd", "1",
                     "--keyserver", "hkps://keyserver.ubuntu.com", "--auto-key-retrieve",
                     "--verify", str(signature), str(artifact)],
                    capture_output=True, text=True, timeout=90,
                )
            except subprocess.TimeoutExpired:
                raise SystemExit(f"Public-key lookup timed out for {artifact.name}") from None
            statuses = [line.split() for line in result.stdout.splitlines()
                        if line.startswith("[GNUPG:] ")]
            valid = [fields for fields in statuses if fields[1] == "VALIDSIG"]
            rejected = {"BADSIG", "ERRSIG", "EXPSIG", "EXPKEYSIG", "REVKEYSIG"}
            if result.returncode or len(valid) != 1 or any(s[1] in rejected for s in statuses):
                # Raw GPG output may include key identities; keep CI diagnostics
                # limited to machine-readable status names and artifact names.
                codes = ", ".join(sorted({s[1] for s in statuses})) or "none"
                raise SystemExit(f"Signature verification failed for {artifact.name}; statuses: {codes}")
            fingerprints.add(valid[0][-1])
            print(f"Verified {artifact.name}")
    if len(fingerprints) != 1:
        raise SystemExit("Release artifacts were signed by different primary keys")
    print(f"Verified all four release signatures; primary key: {fingerprints.pop()}")


if __name__ == "__main__":
    main()
