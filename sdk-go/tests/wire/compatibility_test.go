package wire_test

import (
	"os"
	"testing"

	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
)

func descriptorMessage(t *testing.T, baseline string, name protoreflect.FullName) *dynamicpb.Message {
	t.Helper()
	data, err := os.ReadFile(baseline)
	if err != nil {
		t.Fatal(err)
	}
	var set descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(data, &set); err != nil {
		t.Fatal(err)
	}
	files, err := protodesc.NewFiles(&set)
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := files.FindDescriptorByName(name)
	if err != nil {
		t.Fatal(err)
	}
	return dynamicpb.NewMessage(descriptor.(protoreflect.MessageDescriptor))
}

func legacyMessage(t *testing.T, name protoreflect.FullName) *dynamicpb.Message {
	return descriptorMessage(t, "testdata/legacy-v0.1.1.binpb", name)
}

func releasedMessage(t *testing.T, name protoreflect.FullName) *dynamicpb.Message {
	return descriptorMessage(t, "testdata/released-v1.1.1.binpb", name)
}

func transcode(t *testing.T, source, target proto.Message) {
	t.Helper()
	data, err := proto.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := proto.Unmarshal(data, target); err != nil {
		t.Fatal(err)
	}
}

func TestLegacyTypedNullRoundTrip(t *testing.T) {
	newParameter := &kublingv1.Parameter{
		Value: &kublingv1.Value{Kind: &kublingv1.Value_NullValue{NullValue: &kublingv1.NullValue{}}},
		DeclaredType: &kublingv1.TypeDescriptor{
			Type: kublingv1.ValueType_VALUE_TYPE_ARRAY,
			ElementType: &kublingv1.TypeDescriptor{
				Type:      kublingv1.ValueType_VALUE_TYPE_BIGDECIMAL,
				Precision: proto.Int32(38),
				Scale:     proto.Int32(0),
			},
		},
	}
	oldParameter := legacyMessage(t, "kubling.v1.Parameter")
	transcode(t, newParameter, oldParameter)
	valueField := oldParameter.Descriptor().Fields().ByName("value")
	oldValue := oldParameter.Get(valueField).Message()
	if got := oldValue.WhichOneof(oldValue.Descriptor().Oneofs().ByName("kind")); got == nil || got.Name() != "null_value" {
		t.Fatal("legacy reader no longer sees the original explicit null")
	}
	if len(oldParameter.GetUnknown()) == 0 {
		t.Fatal("legacy reader lost the additive declared type")
	}
	var restored kublingv1.Parameter
	transcode(t, oldParameter, &restored)
	if !proto.Equal(newParameter, &restored) {
		t.Fatal("new -> old -> new lost type, nested element type, or explicit zero scale")
	}
}

func TestOldParameterDoesNotAcquireDeclaredType(t *testing.T) {
	oldParameter := legacyMessage(t, "kubling.v1.Parameter")
	valueField := oldParameter.Descriptor().Fields().ByName("value")
	oldValue := oldParameter.Mutable(valueField).Message()
	nullField := oldValue.Descriptor().Fields().ByName("null_value")
	oldValue.Mutable(nullField)
	var current kublingv1.Parameter
	transcode(t, oldParameter, &current)
	if current.DeclaredType != nil {
		t.Fatal("legacy untyped null acquired a declared type")
	}
	if _, ok := current.GetValue().GetKind().(*kublingv1.Value_NullValue); !ok {
		t.Fatal("legacy explicit null was lost")
	}
}

func TestLegacyQueryKeepsOriginalFields(t *testing.T) {
	request := &kublingv1.QueryRequest{
		ExpiringToken: "test-token", SessionId: "test-session",
		Sql: "SELECT ?", BatchSize: 17, TransactionId: "opaque-transaction",
		Params: []*kublingv1.Parameter{{
			Value: &kublingv1.Value{Kind: &kublingv1.Value_BigdecimalValue{
				BigdecimalValue: "123456789012345678901234567890.000000000000000001",
			}},
		}},
	}
	legacy := legacyMessage(t, "kubling.v1.QueryRequest")
	transcode(t, request, legacy)
	for name, want := range map[protoreflect.Name]string{
		"expiring_token": request.ExpiringToken, "session_id": request.SessionId, "sql": request.Sql,
	} {
		if got := legacy.Get(legacy.Descriptor().Fields().ByName(name)).String(); got != want {
			t.Fatalf("legacy %s changed: %q", name, got)
		}
	}
	var restored kublingv1.QueryRequest
	transcode(t, legacy, &restored)
	if !proto.Equal(request, &restored) {
		t.Fatal("legacy fields, exact decimal, batch size or added transaction assertion changed")
	}
}

func TestNewValueIsUnknownToOldReaderNotNull(t *testing.T) {
	value := &kublingv1.Value{Kind: &kublingv1.Value_ArrayValue{
		ArrayValue: &kublingv1.ArrayValue{
			ElementType: &kublingv1.TypeDescriptor{Type: kublingv1.ValueType_VALUE_TYPE_LONG},
		},
	}}
	legacy := legacyMessage(t, "kubling.v1.Value")
	transcode(t, value, legacy)
	if legacy.WhichOneof(legacy.Descriptor().Oneofs().ByName("kind")) != nil {
		t.Fatal("new alternative reused a known legacy tag")
	}
	if len(legacy.GetUnknown()) == 0 {
		t.Fatal("new alternative disappeared instead of remaining unknown")
	}
	var restored kublingv1.Value
	transcode(t, legacy, &restored)
	if !proto.Equal(value, &restored) {
		t.Fatal("unknown alternative did not survive Go's binary forwarding")
	}
}

func TestLegacyServerInfoDoesNotInventCapabilities(t *testing.T) {
	legacy := legacyMessage(t, "kubling.v1.GetServerInfoResponse")
	legacy.Set(legacy.Descriptor().Fields().ByName("server_version"), protoreflect.ValueOfString("26.1"))
	var current kublingv1.GetServerInfoResponse
	transcode(t, legacy, &current)
	if current.Capabilities != nil || len(current.Features) != 0 || current.ServerVersion != "26.1" {
		t.Fatal("legacy product version must not imply new capability support")
	}
}

func TestV111ReaderForwardsNegotiatedExecutionOutcome(t *testing.T) {
	resultID := uint64(1)
	current := &kublingv1.ExecutionEnd{
		ResultCount:  1,
		Completeness: kublingv1.ExecutionCompleteness_EXECUTION_COMPLETENESS_PARTIAL,
		Warnings: []*kublingv1.KublingWarning{{
			StableCode: "KBL_SOURCE_RESULT_OMITTED",
			Message:    "Some sources were omitted",
			Role:       kublingv1.WarningRole_WARNING_ROLE_PARTIAL_RESULT_CAUSE,
			VendorCode: proto.Int32(0),
			Context: &kublingv1.WarningContext{
				ResultId: proto.Uint64(resultID),
				SourceId: proto.String("inventory-source"),
			},
			OccurrenceCount:                100,
			AffectedResourceIds:            []string{"bmc-1", "bmc-2"},
			OmittedAffectedResourceIdCount: 98,
		}},
		OmittedWarningCount: 2,
	}
	released := releasedMessage(t, "kubling.v1.ExecutionEnd")
	transcode(t, current, released)
	if got := released.Get(released.Descriptor().Fields().ByName("result_count")).Uint(); got != 1 {
		t.Fatalf("v1.1.1 reader lost result_count: %d", got)
	}
	if len(released.GetUnknown()) == 0 {
		t.Fatal("v1.1.1 reader lost additive completeness and warnings")
	}
	var restored kublingv1.ExecutionEnd
	transcode(t, released, &restored)
	if !proto.Equal(current, &restored) || restored.Warnings[0].VendorCode == nil {
		t.Fatal("v1.1.1 forwarding lost negotiated outcome fields or explicit vendor zero")
	}

	request := &kublingv1.ExecuteRequest{
		AcceptedFeatures:    []string{"partial_results_v1"},
		AllowPartialResults: true,
	}
	oldRequest := releasedMessage(t, "kubling.v1.ExecuteRequest")
	transcode(t, request, oldRequest)
	if len(oldRequest.GetUnknown()) == 0 {
		t.Fatal("v1.1.1 reader lost additive partial-result authorization")
	}
	var restoredRequest kublingv1.ExecuteRequest
	transcode(t, oldRequest, &restoredRequest)
	if !proto.Equal(request, &restoredRequest) {
		t.Fatal("v1.1.1 forwarding lost partial-result authorization")
	}
}

func TestRichErrorPreservesMachineFieldsAndPresence(t *testing.T) {
	detail := &kublingv1.KublingError{
		StableCode: "KBL_PARAMETER_TYPE_MISMATCH",
		SqlState:   proto.String("22000"), VendorCode: proto.Int32(0),
		Category:     kublingv1.ErrorCategory_ERROR_CATEGORY_VALIDATION,
		Retryability: kublingv1.Retryability_RETRYABILITY_NEVER,
		SqlExecuted:  proto.Bool(false),
		Transaction: &kublingv1.TransactionStatus{
			TransactionId: "opaque-id", State: kublingv1.TransactionState_TRANSACTION_STATE_UNKNOWN,
		},
	}
	for _, text := range []string{"invalid parameter", "parámetro inválido"} {
		rich, err := status.New(codes.InvalidArgument, text).WithDetails(detail)
		if err != nil {
			t.Fatal(err)
		}
		decoded := status.Convert(rich.Err())
		if decoded.Code() != codes.InvalidArgument || len(decoded.Details()) != 1 {
			t.Fatal("gRPC status or detail count changed")
		}
		got, ok := decoded.Details()[0].(*kublingv1.KublingError)
		if !ok || !proto.Equal(detail, got) || got.VendorCode == nil || got.SqlExecuted == nil || *got.SqlExecuted {
			t.Fatal("structured fields or explicit vendor zero lost")
		}
	}
	var absent kublingv1.KublingError
	transcode(t, &kublingv1.KublingError{StableCode: "KBL_VALUE_UNSUPPORTED"}, &absent)
	if absent.VendorCode != nil || absent.SqlState != nil || absent.Transaction != nil || absent.SqlExecuted != nil {
		t.Fatal("missing error fields acquired fabricated observations")
	}
}

func TestOldClientCanIgnoreUnknownRichError(t *testing.T) {
	rich, err := status.New(codes.InvalidArgument, "diagnostic").WithDetails(&kublingv1.KublingError{
		StableCode: "KBL_PARAMETER_TYPE_MISMATCH", SqlExecuted: proto.Bool(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	// Standard gRPC status still decodes without registering the Kubling detail.
	oldStatus := dynamicpb.NewMessage(rich.Proto().ProtoReflect().Descriptor())
	transcode(t, rich.Proto(), oldStatus)
	if got := oldStatus.Get(oldStatus.Descriptor().Fields().ByName("code")).Int(); got != int64(codes.InvalidArgument) {
		t.Fatalf("legacy status code changed: %d", got)
	}
	if got := oldStatus.Get(oldStatus.Descriptor().Fields().ByName("details")).List().Len(); got != 1 {
		t.Fatal("legacy reader lost the opaque Any detail")
	}
	_, err = anypb.UnmarshalNew(rich.Proto().Details[0], proto.UnmarshalOptions{Resolver: new(protoregistry.Types)})
	if err == nil {
		t.Fatal("empty legacy type registry unexpectedly knows KublingError")
	}
	if status.Convert(rich.Err()).Code() != codes.InvalidArgument {
		t.Fatal("unknown detail must not prevent using standard gRPC status")
	}
}
