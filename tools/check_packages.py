#!/usr/bin/env python3
"""Inspect the distributable files, not just the compiler output directories."""

import argparse
from pathlib import Path
import tarfile
import zipfile

ROOT = Path(__file__).resolve().parents[1]


def require_members(actual, required):
    missing = set(required) - set(actual)
    if missing:
        raise ValueError(f"Missing package members: {sorted(missing)}")
    if any(".local-notes" in name or "__pycache__" in name for name in actual):
        raise ValueError("Local notes or bytecode caches leaked into a distribution")


def java():
    jars = list((ROOT / "sdk-java/target").glob("*.jar"))
    binaries = [p for p in jars if not p.name.endswith(("-sources.jar", "-javadoc.jar"))]
    if len(binaries) != 1:
        raise ValueError("Expected exactly one Java binary JAR; run clean verify")
    binary = binaries[0]
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
    dist = ROOT / "sdk-python/dist"
    wheels, sources = list(dist.glob("*.whl")), list(dist.glob("*.tar.gz"))
    if len(wheels) != 1 or len(sources) != 1:
        raise ValueError("Expected one wheel and one sdist; clear stale dist artifacts")
    required = ["kubling/features.py", "kubling/features.json", "kubling/proto/kubling/v1/command.proto"]
    required += [f"kubling/v1/{name}_pb2.py" for name in
                 ("command", "value", "capability", "transaction", "error", "lob")]
    required += [f"kubling/v1/{name}_pb2_grpc.py" for name in ("command", "lob")]
    with zipfile.ZipFile(wheels[0]) as archive:
        names = archive.namelist()
        require_members(names, required)
        if not any(name.endswith(".dist-info/licenses/LICENSE") for name in names):
            raise ValueError("Wheel license missing")
    with tarfile.open(sources[0]) as archive:
        names = [name.partition("/")[2] for name in archive.getnames()]
        require_members(names, ["generated/" + name for name in required]
                        + ["LICENSE", "pyproject.toml", "_build_backend.py"])


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("language", choices=("java", "python"))
    args = parser.parse_args()
    {"java": java, "python": python}[args.language]()
    print(f"Validated {args.language} distribution contents")
