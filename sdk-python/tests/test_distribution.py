"""Run against an installed wheel, outside the generated source directory."""

import importlib
from importlib import metadata, resources
import unittest

import grpc
from google.protobuf.any_pb2 import Any
from kubling import features
from kubling.v1 import command_pb2, command_pb2_grpc, error_pb2, lob_pb2_grpc, value_pb2


class DistributionTest(unittest.TestCase):
    def test_distribution_contains_all_messages_and_contract(self):
        self.assertTrue(metadata.version("kubling-grpc"))
        for name in ("command", "value", "transaction", "capability", "error", "lob"):
            importlib.import_module(f"kubling.v1.{name}_pb2")
        package = resources.files("kubling")
        self.assertTrue(package.joinpath("features.json").is_file())
        self.assertTrue(package.joinpath("proto/kubling/v1/command.proto").is_file())
        self.assertEqual(features.GENERIC_EXECUTE_V1, "generic_execute_v1")

    def test_typed_null_and_recursive_array(self):
        parameter = command_pb2.Parameter(
            value=value_pb2.Value(null_value=value_pb2.NullValue()),
            declared_type=value_pb2.TypeDescriptor(type=value_pb2.VALUE_TYPE_STRING),
        )
        decoded = command_pb2.Parameter.FromString(parameter.SerializeToString())
        self.assertTrue(decoded.HasField("declared_type"))
        self.assertEqual(decoded.value.WhichOneof("kind"), "null_value")
        array = value_pb2.TypeDescriptor(type=value_pb2.VALUE_TYPE_ARRAY)
        array.element_type.type = value_pb2.VALUE_TYPE_ARRAY
        array.element_type.element_type.type = value_pb2.VALUE_TYPE_STRING
        self.assertEqual(array, value_pb2.TypeDescriptor.FromString(array.SerializeToString()))

    def test_error_presence_and_any(self):
        detail = error_pb2.KublingError(stable_code="INVALID_PARAMETER", sql_executed=False)
        packed = Any()
        packed.Pack(detail)
        decoded = error_pb2.KublingError()
        self.assertTrue(packed.Unpack(decoded))
        self.assertTrue(decoded.HasField("sql_executed"))
        self.assertFalse(decoded.sql_executed)
        self.assertFalse(error_pb2.KublingError().HasField("sql_executed"))

    def test_constructs_all_stubs_without_connecting(self):
        with grpc.insecure_channel("localhost:1") as channel:
            query = command_pb2_grpc.QueryServiceStub(channel)
            self.assertIsInstance(query.Execute, grpc.UnaryStreamMultiCallable)
            self.assertIsInstance(query.Query, grpc.UnaryStreamMultiCallable)
            self.assertIsInstance(query.Exec, grpc.UnaryUnaryMultiCallable)
            self.assertIsInstance(lob_pb2_grpc.LobServiceStub(channel).WriteLob, grpc.StreamUnaryMultiCallable)
            self.assertIsNotNone(command_pb2_grpc.SessionServiceStub(channel).Login)


if __name__ == "__main__":
    unittest.main()
