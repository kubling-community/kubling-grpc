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
    KublingWarning unknownRole = KublingWarning.newBuilder().setRoleValue(999).build();
    assertEquals(WarningRole.UNRECOGNIZED, KublingWarning.parseFrom(unknownRole.toByteArray()).getRole());
  }

  @Test
  void preservesPartialOutcomeAndWarningPresence() throws Exception {
    KublingWarning warning = KublingWarning.newBuilder()
        .setStableCode("KBL_SOURCE_RESULT_OMITTED")
        .setMessage("Some sources were omitted")
        .setRole(WarningRole.WARNING_ROLE_PARTIAL_RESULT_CAUSE)
        .setVendorCode(0)
        .setOccurrenceCount(2)
        .addAffectedResourceIds("bmc-1")
        .setContext(WarningContext.newBuilder().setResultId(1).setSourceId("inventory-source"))
        .build();
    ExecutionEnd end = ExecutionEnd.newBuilder()
        .setResultCount(1)
        .setCompleteness(ExecutionCompleteness.EXECUTION_COMPLETENESS_PARTIAL)
        .addWarnings(warning)
        .setOmittedWarningCount(3)
        .build();
    ExecutionEnd decoded = ExecutionEnd.parseFrom(end.toByteArray());
    assertEquals(ExecutionCompleteness.EXECUTION_COMPLETENESS_PARTIAL, decoded.getCompleteness());
    assertEquals(WarningRole.WARNING_ROLE_PARTIAL_RESULT_CAUSE, decoded.getWarnings(0).getRole());
    assertTrue(decoded.getWarnings(0).hasVendorCode());
    assertEquals(0, decoded.getWarnings(0).getVendorCode());
    assertTrue(decoded.getWarnings(0).getContext().hasResultId());
  }

  @Test
  void includesLegacyAndNewServiceDescriptors() {
    assertEquals(MethodDescriptor.MethodType.SERVER_STREAMING, QueryServiceGrpc.getExecuteMethod().getType());
    assertEquals(MethodDescriptor.MethodType.SERVER_STREAMING, QueryServiceGrpc.getQueryMethod().getType());
    assertEquals(MethodDescriptor.MethodType.UNARY, QueryServiceGrpc.getExecMethod().getType());
    assertEquals(MethodDescriptor.MethodType.CLIENT_STREAMING, LobServiceGrpc.getWriteLobMethod().getType());
    assertEquals("generic_execute_v1", Features.GENERIC_EXECUTE_V1);
    assertEquals("partial_results_v1", Features.PARTIAL_RESULTS_V1);
    assertEquals(2, WarningRole.WARNING_ROLE_PARTIAL_RESULT_CAUSE.getNumber());
    assertNotNull(getClass().getResource("/META-INF/proto/kubling/v1/command.proto"));
    assertNotNull(getClass().getResource("/META-INF/proto/kubling/v1/warning.proto"));
    assertNotNull(getClass().getResource("/META-INF/kubling/features.json"));
  }
}
