package com.kubling.transport.grpc;

import static org.junit.jupiter.api.Assertions.*;

import com.google.protobuf.Any;
import io.grpc.MethodDescriptor;
import org.junit.jupiter.api.Test;

class ContractTest {
  @Test
  void preservesTypedNullAndOptionalFalse() throws Exception {
    Parameter parameter = Parameter.newBuilder()
        .setValue(Value.newBuilder().setNullValue(NullValue.getDefaultInstance()))
        .setDeclaredType(TypeDescriptor.newBuilder().setType(ValueType.VALUE_TYPE_STRING))
        .build();
    Parameter decoded = Parameter.parseFrom(parameter.toByteArray());
    assertTrue(decoded.hasDeclaredType());
    assertEquals(Value.KindCase.NULL_VALUE, decoded.getValue().getKindCase());
    KublingError detail = KublingError.newBuilder()
        .setStableCode("INVALID_PARAMETER").setSqlExecuted(false).build();
    KublingError unpacked = Any.parseFrom(Any.pack(detail).toByteArray()).unpack(KublingError.class);
    assertTrue(unpacked.hasSqlExecuted());
    assertFalse(unpacked.getSqlExecuted());
    assertFalse(KublingError.getDefaultInstance().hasSqlExecuted());
  }

  @Test
  void preservesRecursiveDescriptorsAndUnknownEnums() throws Exception {
    TypeDescriptor array = TypeDescriptor.newBuilder().setType(ValueType.VALUE_TYPE_ARRAY)
        .setElementType(TypeDescriptor.newBuilder().setType(ValueType.VALUE_TYPE_ARRAY)
            .setElementType(TypeDescriptor.newBuilder().setType(ValueType.VALUE_TYPE_STRING))).build();
    assertEquals(array, TypeDescriptor.parseFrom(array.toByteArray()));
    TransactionStatus unknown = TransactionStatus.newBuilder().setStateValue(999).build();
    assertEquals(TransactionState.UNRECOGNIZED, TransactionStatus.parseFrom(unknown.toByteArray()).getState());
  }

  @Test
  void includesLegacyAndNewServiceDescriptors() {
    assertEquals(MethodDescriptor.MethodType.SERVER_STREAMING, QueryServiceGrpc.getExecuteMethod().getType());
    assertEquals(MethodDescriptor.MethodType.SERVER_STREAMING, QueryServiceGrpc.getQueryMethod().getType());
    assertEquals(MethodDescriptor.MethodType.UNARY, QueryServiceGrpc.getExecMethod().getType());
    assertEquals(MethodDescriptor.MethodType.CLIENT_STREAMING, LobServiceGrpc.getWriteLobMethod().getType());
    assertEquals("generic_execute_v1", Features.GENERIC_EXECUTE_V1);
    assertNotNull(getClass().getResource("/META-INF/proto/kubling/v1/command.proto"));
    assertNotNull(getClass().getResource("/META-INF/kubling/features.json"));
  }
}
