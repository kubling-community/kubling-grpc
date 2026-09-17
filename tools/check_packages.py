#!/usr/bin/env python3
"""Inspect the distributable files, not just the compiler output directories."""

import argparse
from pathlib import Path
import tarfile
import zipfile

from check_release import release_version

ROOT = Path(__file__).resolve().parents[1]


def require_members(actual, required):
    missing = set(required) - set(actual)
    if missing:
        raise ValueError(f"Missing package members: {sorted(missing)}")
    if any(".local-notes" in name or "__pycache__" in name for name in actual):
        raise ValueError("Local notes or bytecode caches leaked into a distribution")


def java():
    version = release_version()
    jars = list((ROOT / "sdk-java/target").glob("*.jar"))
    binaries = [p for p in jars if not p.name.endswith(("-sources.jar", "-javadoc.jar"))]
    if len(binaries) != 1:
        raise ValueError("Expected exactly one Java binary JAR; run clean verify")
    binary = binaries[0]
    if binary.name != f"kubling-grpc-{version}.jar":
        raise ValueError(f"Java artifact does not match VERSION {version}: {binary.name}")
    package = "com/kubling/transport/grpc/"
    with zipfile.ZipFile(binary) as archive:
        require_members(archive.namelist(), [package + name + ".class" for name in
                        ("Features", "ExecuteRequest", "KublingError", "QueryServiceGrpc", "LobServiceGrpc")]
                        + ["META-INF/LICENSE", "META-INF/proto/kubling/v1/command.proto", "META-INF/kubling/features.json"])
        for name in archive.namelist():
            if name.endswith(".class") and int.from_bytes(archive.read(name)[6:8], "big") != 65:
                raise ValueError("The Java artifact must contain Java 21 bytecode")
    with zipfile.ZipFile(binary.with_name(binary.stem + "-sources.jar")) as archive:
        require_members(archive.namelist(), [package + "Features.java", package + "ExecuteRequest.java"])
    with zipfile.ZipFile(binary.with_name(binary.stem + "-javadoc.jar")) as archive:
        require_members(archive.namelist(), ["index.html", package + "ExecuteRequest.html"])


def python():
    version = release_version()
    dist = ROOT / "sdk-python/dist"
    wheels, sources = list(dist.glob("*.whl")), list(dist.glob("*.tar.gz"))
    if len(wheels) != 1 or len(sources) != 1:
        raise ValueError("Expected one wheel and one sdist; clear stale dist artifacts")
    prefix = f"kubling_grpc-{version}"
    if not wheels[0].name.startswith(prefix + "-") or sources[0].name != prefix + ".tar.gz":
        raise ValueError(f"Python artifacts do not match VERSION {version}")
    required = ["kubling/features.py", "kubling/features.json", "kubling/proto/kubling/v1/command.proto"]
    required += [f"kubling/v1/{name}_pb2.py" for name in
                 ("command", "value", "capability", "transaction", "error", "lob")]
    required += [f"kubling/v1/{name}_pb2_grpc.py" for name in ("command", "lob")]
    with zipfile.ZipFile(wheels[0]) as archive:
        names = archive.namelist()
        require_members(names, required)
        if not any(name.endswith(".dist-info/licenses/LICENSE") for name in names):
            raise ValueError("Wheel license missing")
        metadata = [name for name in names if name.endswith(".dist-info/METADATA")]
        if len(metadata) != 1 or f"\nVersion: {version}\n" not in archive.read(metadata[0]).decode():
            raise ValueError("Wheel metadata version does not match VERSION")
    with tarfile.open(sources[0]) as archive:
        members = archive.getnames()
        if not members or any(name != prefix and not name.startswith(prefix + "/") for name in members):
            raise ValueError("Source distribution root does not match VERSION")
        names = [name.partition("/")[2] for name in members]
        require_members(names, ["generated/" + name for name in required]
                        + ["LICENSE", "pyproject.toml", "_build_backend.py"])


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("language", choices=("java", "python"))
    args = parser.parse_args()
    {"java": java, "python": python}[args.language]()
    print(f"Validated {args.language} distribution contents")
