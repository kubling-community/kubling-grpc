"""Reference checks for portable contract fixtures, not SDK or engine handlers.

These checks make selected semantic invariants executable. They do not implement
SQL execution, authentication, resource ownership, cancellation or transactions.
"""

from decimal import Decimal, InvalidOperation

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
