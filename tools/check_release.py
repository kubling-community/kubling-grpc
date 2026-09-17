#!/usr/bin/env python3
"""Validate the canonical release version, manifests and coordinated tags."""

import argparse
from pathlib import Path
import re
import subprocess
import xml.etree.ElementTree as ET

try:
    import tomllib
except ModuleNotFoundError:
    import tomli as tomllib

ROOT = Path(__file__).resolve().parents[1]
STABLE_VERSION = re.compile(r"(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)")
TAG_PREFIXES = {
    "proto": "proto/v",
    "go": "sdk-go/v",
    "java": "sdk-java/v",
    "python": "sdk-python/v",
}


def release_version(root=ROOT):
    version = (root / "VERSION").read_text(encoding="utf-8").strip()
    if not STABLE_VERSION.fullmatch(version):
        raise ValueError("VERSION must contain one stable MAJOR.MINOR.PATCH version")
    return version


def package_versions(root=ROOT):
    java = ET.parse(root / "sdk-java/pom.xml").findtext(
        "{http://maven.apache.org/POM/4.0.0}version"
    )
    with (root / "sdk-python/pyproject.toml").open("rb") as manifest:
        python = tomllib.load(manifest)["project"]["version"]
    return {"java": java, "python": python}


def package_version(language, root=ROOT):
    return package_versions(root)[language]


def validate_manifests(root=ROOT):
    version = release_version(root)
    mismatches = {
        language: value
        for language, value in package_versions(root).items()
        if value != version
    }
    if mismatches:
        details = ", ".join(f"{language}={value}" for language, value in mismatches.items())
        raise ValueError(f"Package versions must equal VERSION {version}: {details}")
    expected_documentation = {
        "sdk-go/README.md": f"sdk-go@v{version}",
        "sdk-java/README.md": f"<version>{version}</version>",
        "sdk-python/README.md": f"kubling-grpc=={version}",
    }
    stale = [
        relative
        for relative, marker in expected_documentation.items()
        if marker not in (root / relative).read_text(encoding="utf-8")
    ]
    if stale:
        raise ValueError("Public installation examples do not match VERSION: " + ", ".join(stale))
    return version


def release_tags(version):
    if not STABLE_VERSION.fullmatch(version):
        raise ValueError(f"Invalid stable release version: {version}")
    return {language: prefix + version for language, prefix in TAG_PREFIXES.items()}


def validate_tag(language, tag, root=ROOT):
    version = validate_manifests(root)
    expected = release_tags(version)[language]
    if tag != expected:
        raise ValueError(f"Expected tag {expected}, got {tag}")
    return expected


def git(*arguments, root=ROOT, check=True):
    result = subprocess.run(
        ["git", *arguments],
        cwd=root,
        check=check,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    return result.stdout.strip()


def tag_commit(tag, root=ROOT):
    result = git("rev-parse", "--verify", f"refs/tags/{tag}^{{commit}}", root=root, check=False)
    return result or None


def validate_tags_available(root=ROOT):
    version = validate_manifests(root)
    existing = [tag for tag in release_tags(version).values() if tag_commit(tag, root)]
    if existing:
        raise ValueError("Release tags already exist: " + ", ".join(existing))
    return version


def validate_release_train(expected_commit="HEAD", root=ROOT):
    version = validate_manifests(root)
    expected = git("rev-parse", f"{expected_commit}^{{commit}}", root=root)
    resolved = {tag: tag_commit(tag, root) for tag in release_tags(version).values()}
    missing = [tag for tag, commit in resolved.items() if commit is None]
    if missing:
        raise ValueError("Missing release tags: " + ", ".join(missing))
    mismatches = {tag: commit for tag, commit in resolved.items() if commit != expected}
    if mismatches:
        details = ", ".join(f"{tag}={commit}" for tag, commit in mismatches.items())
        raise ValueError(f"Release tags must resolve to {expected}: {details}")
    return version, expected


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "target",
        choices=("all", "available", "train", *TAG_PREFIXES),
        help="Validation to run",
    )
    parser.add_argument("value", nargs="?", help="Tag or expected commit")
    args = parser.parse_args()

    try:
        if args.target == "all":
            if args.value:
                parser.error("all does not accept a value")
            version = validate_manifests()
            print(f"Validated coordinated version {version}")
        elif args.target == "available":
            if args.value:
                parser.error("available does not accept a value")
            version = validate_tags_available()
            print(f"Validated release tag availability for {version}")
        elif args.target == "train":
            version, commit = validate_release_train(args.value or "HEAD")
            print(f"Validated release train {version} at {commit}")
        else:
            if not args.value:
                parser.error(f"{args.target} requires a tag")
            tag = validate_tag(args.target, args.value)
            print(f"Validated {tag}")
    except (OSError, ValueError, subprocess.CalledProcessError) as error:
        parser.error(str(error))


if __name__ == "__main__":
    main()
