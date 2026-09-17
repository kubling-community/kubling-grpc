#!/usr/bin/env python3
"""Require a language-specific release tag to match the package manifest."""

import argparse
from pathlib import Path
import re
import xml.etree.ElementTree as ET

try:
    import tomllib
except ModuleNotFoundError:
    import tomli as tomllib

ROOT = Path(__file__).resolve().parents[1]


def package_version(language):
    if language == "java":
        return ET.parse(ROOT / "sdk-java/pom.xml").findtext("{http://maven.apache.org/POM/4.0.0}version")
    with (ROOT / "sdk-python/pyproject.toml").open("rb") as manifest:
        return tomllib.load(manifest)["project"]["version"]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("language", choices=("java", "python"))
    parser.add_argument("tag")
    args = parser.parse_args()
    version = package_version(args.language)
    suffix = r"(?:-RC[0-9]+)?" if args.language == "java" else r"(?:rc[0-9]+)?"
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+" + suffix, version):
        parser.error("Release requires an explicit stable or RC version in the package manifest")
    expected = f"sdk-{args.language}/v{version}"
    if args.tag != expected:
        parser.error(f"Expected tag {expected}")
    print(f"Validated {expected}")


if __name__ == "__main__":
    main()
