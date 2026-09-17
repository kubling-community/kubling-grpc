import copy
import json
from pathlib import Path
import sys
import unittest

TOOLS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOLS))
import conformance
import generate_features

REGISTRY = json.loads((TOOLS.parent / "protocol/features.json").read_text())
CASES = json.loads((TOOLS.parent / "protocol/conformance/cases.json").read_text())


class RegistryTests(unittest.TestCase):
    def test_current_registry(self):
        generate_features.validate_registry(REGISTRY)

    def test_malformed_registries(self):
        mutations = {
            "unknown_schema": lambda r: r.update(schema_version=3),
            "non_integer_schema": lambda r: r.update(schema_version=2.0),
            "unknown_root_key": lambda r: r.update(extra=True),
            "missing_activation": lambda r: r["features"][0].pop("activation"),
            "unknown_activation_key": lambda r: r["features"][0]["activation"].update(extra=True),
            "non_boolean_acceptance": lambda r: r["features"][0]["activation"].update(request_requires_acceptance=1),
            "mixed_operators": lambda r: r["features"][0]["activation"]["request"].update(accepted_features=True),
            "unknown_predicate": lambda r: r["features"][1]["activation"]["request"].update(test="truthy"),
            "invalid_field": lambda r: r["features"][1]["activation"]["request"].update(field="local.field"),
            "invalid_rpc": lambda r: r["features"][0]["activation"]["request"].update(rpc="Execute"),
            "advertised_request": lambda r: r["features"][0]["activation"].update(request={"advertised": True}),
            "false_acceptance_operator": lambda r: r["features"][-1]["activation"].update(response={"accepted_features": False}),
            "empty_group": lambda r: r["features"][0]["activation"].update(request={"all_of": []}),
            "duplicate_feature": lambda r: r["features"].append(copy.deepcopy(r["features"][0])),
            "unknown_dependency": lambda r: r["features"][0]["requires"].append("missing_v1"),
            "dependency_cycle": lambda r: r["features"][0]["requires"].append("array_values_v1"),
        }
        for name, mutate in mutations.items():
            with self.subTest(name=name):
                registry = copy.deepcopy(REGISTRY)
                mutate(registry)
                with self.assertRaises(ValueError):
                    generate_features.validate_registry(registry)


class ConformanceTests(unittest.TestCase):
    def test_transactions(self):
        for case in CASES["transactions"]:
            with self.subTest(name=case["name"]):
                self.assertEqual(case["valid"], conformance.transaction_valid(case["status"], case["known_id"]))

    def test_parameters(self):
        for case in CASES["parameters"]:
            with self.subTest(name=case["name"]):
                self.assertEqual(case["valid"], conformance.parameter_valid(case["parameter"]))

    def test_feature_activation(self):
        features = generate_features.validate_registry(REGISTRY)
        for case in CASES["features"]:
            with self.subTest(name=case["name"]):
                context = {"advertised": CASES["default_advertised"], **case}
                self.assertEqual((case["request_allowed"], case["response_allowed"]),
                                 conformance.feature_decisions(features, context, case["response_feature"]))

    def test_lob_leases(self):
        for case in CASES["lob_leases"]:
            with self.subTest(name=case["name"]):
                self.assertEqual(case["lease_valid"], conformance.lease_valid(case))
                self.assertEqual(case["readable"], conformance.reference_readable(case))


if __name__ == "__main__":
    unittest.main()
