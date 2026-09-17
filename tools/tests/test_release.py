from pathlib import Path
import sys
import tempfile
import unittest
from unittest import mock

TOOLS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOLS))
import check_release


class ReleaseVersionTests(unittest.TestCase):
    def test_release_tags_share_one_version(self):
        self.assertEqual(
            check_release.release_tags("1.1.1"),
            {
                "proto": "proto/v1.1.1",
                "go": "sdk-go/v1.1.1",
                "java": "sdk-java/v1.1.1",
                "python": "sdk-python/v1.1.1",
            },
        )

    def test_version_requires_stable_semver(self):
        for invalid in ("v1.1.1", "1.1", "1.1.1-rc1", "01.1.1", ""):
            with self.subTest(version=invalid), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                (root / "VERSION").write_text(invalid, encoding="utf-8")
                with self.assertRaises(ValueError):
                    check_release.release_version(root)

    def test_current_manifests_match_version(self):
        self.assertEqual(check_release.validate_manifests(), "1.1.1")

    def test_each_tag_must_match_canonical_version(self):
        self.assertEqual(
            check_release.validate_tag("go", "sdk-go/v1.1.1"),
            "sdk-go/v1.1.1",
        )
        with self.assertRaises(ValueError):
            check_release.validate_tag("go", "sdk-go/v1.1.2")

    @mock.patch("check_release.validate_manifests", return_value="1.1.1")
    @mock.patch("check_release.git", return_value="a" * 40)
    @mock.patch("check_release.tag_commit", return_value="a" * 40)
    def test_release_train_requires_all_tags_on_one_commit(self, *_):
        self.assertEqual(
            check_release.validate_release_train(),
            ("1.1.1", "a" * 40),
        )

    @mock.patch("check_release.validate_manifests", return_value="1.1.1")
    @mock.patch("check_release.git", return_value="a" * 40)
    @mock.patch("check_release.tag_commit")
    def test_release_train_rejects_missing_or_divergent_tags(self, tag_commit, *_):
        tag_commit.side_effect = ["a" * 40, "a" * 40, None, "a" * 40]
        with self.assertRaisesRegex(ValueError, "Missing release tags"):
            check_release.validate_release_train()

        tag_commit.side_effect = ["a" * 40, "b" * 40, "a" * 40, "a" * 40]
        with self.assertRaisesRegex(ValueError, "must resolve"):
            check_release.validate_release_train()


if __name__ == "__main__":
    unittest.main()
