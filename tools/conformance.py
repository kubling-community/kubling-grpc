"""Reference checks for portable contract fixtures, not SDK or engine handlers.

These checks make selected semantic invariants executable. They do not implement
SQL execution, authentication, resource ownership, cancellation or transactions.
"""

from decimal import Decimal, InvalidOperation
import re

KNOWN_TYPES = set("STRING VARBINARY CHAR BOOLEAN BYTE SHORT INTEGER LONG BIGINTEGER FLOAT DOUBLE BIGDECIMAL DATE TIME TIMESTAMP BLOB CLOB GEOMETRY GEOGRAPHY JSON XML ARRAY".split())


def transaction_valid(status, known_id=""):
    state = status.get("state", "TRANSACTION_STATE_UNKNOWN").removeprefix("TRANSACTION_STATE_")
    identifier = status.get("transactionId", "")
    if state == "NONE":
        return identifier == ""
    if state in {"ACTIVE", "ROLLBACK_ONLY", "COMMITTED", "ROLLED_BACK"}:
        return bool(identifier)
    if state == "UNKNOWN":
        return identifier == known_id
    return False


def descriptor_valid(descriptor, max_dimensions=8):
    if not isinstance(descriptor, dict):
        return False
    kind = descriptor.get("type", "").removeprefix("VALUE_TYPE_")
    if kind not in KNOWN_TYPES:
        return False
    if kind == "ARRAY":
        if max_dimensions <= 0 or not descriptor_valid(descriptor.get("elementType"), max_dimensions - 1):
            return False
    elif "elementType" in descriptor:
        return False
    if "precision" in descriptor:
        precision = descriptor["precision"]
        if kind not in {"BIGINTEGER", "BIGDECIMAL"} or type(precision) is not int or not 1 <= precision <= 2147483647:
            return False
    if "scale" in descriptor:
        scale = descriptor["scale"]
        if kind != "BIGDECIMAL" or "precision" not in descriptor or type(scale) is not int or not -2147483648 <= scale <= descriptor["precision"]:
            return False
    return True


def numeric_fits(value, descriptor):
    try:
        number = Decimal(value)
    except (InvalidOperation, ValueError, TypeError):
        return False
    if not number.is_finite():
        return False
    digits, exponent = number.as_tuple().digits, number.as_tuple().exponent
    if not any(digits):
        return True
    # Remove insignificant trailing zeros without Decimal context rounding.
    while len(digits) > 1 and digits[-1] == 0:
        digits = digits[:-1]
        exponent += 1
    kind = descriptor["type"].removeprefix("VALUE_TYPE_")
    target_scale = 0 if kind == "BIGINTEGER" else descriptor.get("scale", -exponent)
    shift = exponent + target_scale
    if shift < 0:
        return False  # Would require rounding.
    return "precision" not in descriptor or len(digits) + shift <= descriptor["precision"]


def value_matches(value, descriptor):
    if not isinstance(value, dict) or len(value) != 1:
        return False
    variant, payload = next(iter(value.items()))
    if variant == "nullValue":
        return payload == {}
    kind = descriptor["type"].removeprefix("VALUE_TYPE_")
    if variant == "arrayValue":
        return (kind == "ARRAY" and payload.get("elementType") == descriptor["elementType"]
                and all(value_matches(v, descriptor["elementType"]) for v in payload.get("elements", [])))
    if variant == "lobReference":
        return kind in {"BLOB", "CLOB"} and payload.get("type") == descriptor["type"]
    actual = {"geometryWithCrs": "GEOMETRY", "geographyWithCrs": "GEOGRAPHY"}.get(variant)
    if actual is None and variant.endswith("Value"):
        actual = variant[:-5].upper()
    if actual != kind:
        return False
    if kind in {"BIGINTEGER", "BIGDECIMAL"}:
        return numeric_fits(payload, descriptor)
    return True


def parameter_valid(parameter):
    if "declaredType" not in parameter:
        return True  # Legacy inference remains the engine's existing behavior.
    descriptor = parameter["declaredType"]
    return descriptor_valid(descriptor) and value_matches(parameter.get("value"), descriptor)


def _unsigned(value, positive=False):
    if type(value) is int:
        parsed = value
    elif isinstance(value, str) and re.fullmatch(r"0|[1-9][0-9]*", value):
        parsed = int(value)
    else:
        return False
    if parsed > 2**64 - 1:
        return False
    return parsed > 0 if positive else parsed >= 0


def warning_valid(warning, max_resource_ids):
    if not isinstance(warning, dict):
        return False
    if not isinstance(warning.get("stableCode"), str) or not warning["stableCode"].strip():
        return False
    if not isinstance(warning.get("message"), str) or not warning["message"].strip():
        return False
    role = warning.get("role", "WARNING_ROLE_UNSPECIFIED")
    if not isinstance(role, str):
        return False
    role = role.removeprefix("WARNING_ROLE_")
    if role not in {"GENERAL", "PARTIAL_RESULT_CAUSE"}:
        return False
    if not _unsigned(warning.get("occurrenceCount"), positive=True):
        return False
    sql_state = warning.get("sqlState")
    if sql_state is not None and (not isinstance(sql_state, str) or re.fullmatch(r"[0-9A-Z]{5}", sql_state) is None):
        return False
    vendor_code = warning.get("vendorCode")
    if vendor_code is not None and (type(vendor_code) is not int or not -(2**31) <= vendor_code < 2**31):
        return False
    identifiers = warning.get("affectedResourceIds", [])
    if (not isinstance(identifiers, list) or len(identifiers) > max_resource_ids
            or any(not isinstance(identifier, str) or not identifier for identifier in identifiers)):
        return False
    return _unsigned(warning.get("omittedAffectedResourceIdCount", 0))


def execution_end_valid(case):
    end = case["execution_end"]
    warnings = end.get("warnings", [])
    omitted = end.get("omittedWarningCount", 0)
    completeness = end.get("completeness", "EXECUTION_COMPLETENESS_UNSPECIFIED")
    if not isinstance(completeness, str):
        return False
    completeness = completeness.removeprefix("EXECUTION_COMPLETENESS_")
    if not case["feature_accepted"]:
        return completeness == "UNSPECIFIED" and warnings == [] and _unsigned(omitted) and int(omitted) == 0
    max_warnings = case["max_warnings_per_execution"]
    max_resource_ids = case["max_affected_resource_ids_per_warning"]
    if (completeness not in {"COMPLETE", "PARTIAL"} or type(max_warnings) is not int
            or type(max_resource_ids) is not int or max_warnings <= 0 or max_resource_ids <= 0
            or not isinstance(warnings, list) or len(warnings) > max_warnings or not _unsigned(omitted)):
        return False
    if not all(warning_valid(warning, max_resource_ids) for warning in warnings):
        return False
    roles = {warning["role"].removeprefix("WARNING_ROLE_") for warning in warnings}
    if completeness == "PARTIAL":
        return case["allow_partial_results"] and "PARTIAL_RESULT_CAUSE" in roles
    return "PARTIAL_RESULT_CAUSE" not in roles


def expression_matches(expression, name, context):
    if expression is None:
        return False
    if "any_of" in expression:
        return any(expression_matches(e, name, context) for e in expression["any_of"])
    if "all_of" in expression:
        return all(expression_matches(e, name, context) for e in expression["all_of"])
    if "rpc" in expression:
        return context["rpc"] == expression["rpc"]
    if "accepted_features" in expression:
        return name in context.get("accepted", [])
    if "advertised" in expression:
        return name in context["advertised"]
    fields, selector = context.get("fields", {}), expression["field"]
    value = fields.get(selector)
    return {"present": selector in fields and value is not None,
            "nonempty": isinstance(value, str) and bool(value),
            "true": value is True}[expression["test"]]


def feature_decisions(features, context, response_feature):
    registry = {f["name"]: f for f in features}
    advertised = set(context["advertised"])
    accepted = context.get("accepted", [])

    def supported(name):
        return name in advertised and all(supported(dep) for dep in registry[name]["requires"])

    allowed = len(set(accepted)) == len(accepted) and all(n in registry and supported(n) for n in accepted)
    for feature in features:
        name, activation = feature["name"], feature["activation"]
        if expression_matches(activation["request"], name, context):
            allowed = allowed and supported(name) and (not activation["request_requires_acceptance"] or name in accepted)
    feature = registry[response_feature]
    response_allowed = supported(response_feature) and expression_matches(feature["activation"]["response"], response_feature, context)
    return allowed, response_allowed


def lease_valid(case):
    reference, creation = case["reference"], case["creation"]
    if not reference.get("lobId") or not reference.get("sessionId") or reference.get("type") not in {"VALUE_TYPE_BLOB", "VALUE_TYPE_CLOB"}:
        return False
    if "expiresAtUnixMs" not in reference or int(creation["retention_seconds"]) <= 0 or int(creation["max_lob_bytes"]) <= 0:
        return False
    minimum = int(creation["issued_at_unix_ms"]) + 1000 * int(creation["retention_seconds"])
    return (int(reference["expiresAtUnixMs"]) >= minimum
            and ("sizeBytes" not in reference or int(reference["sizeBytes"]) <= int(creation["max_lob_bytes"])))


def reference_readable(case):
    reference, current = case["reference"], case["current"]
    return (lease_valid(case) and not case.get("released", False)
            and not case.get("session_closed", False) and not case.get("node_lost", False)
            and int(case["now_unix_ms"]) < int(reference["expiresAtUnixMs"])
            and current["session_id"] == reference["sessionId"]
            and current["node_id"] == case["creation"]["node_id"]
            and current["capability_id"] != "" and current["lob_read_supported"])
