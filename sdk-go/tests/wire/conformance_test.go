package wire_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestActivationSelectorsMatchCompiledContract(t *testing.T) {
	data, err := os.ReadFile("../../../protocol/features.json")
	if err != nil {
		t.Fatal(err)
	}
	var registry struct {
		Features []struct {
			Name       string
			Activation struct{ Request, Response map[string]json.RawMessage }
		}
	}
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	var visit func(*testing.T, map[string]json.RawMessage)
	visit = func(t *testing.T, expression map[string]json.RawMessage) {
		t.Helper()
		var selector, predicate string
		if raw, ok := expression["rpc"]; ok {
			if err := json.Unmarshal(raw, &selector); err != nil {
				t.Fatal(err)
			}
			parts := strings.Split(selector, "/")
			if len(parts) != 3 {
				t.Fatalf("invalid RPC selector: %q", selector)
			}
			d, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(parts[1]))
			if err != nil {
				t.Fatal(err)
			}
			if d.(protoreflect.ServiceDescriptor).Methods().ByName(protoreflect.Name(parts[2])) == nil {
				t.Fatalf("unknown RPC: %s", selector)
			}
		}
		if raw, ok := expression["field"]; ok {
			if err := json.Unmarshal(raw, &selector); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(expression["test"], &predicate); err != nil {
				t.Fatal(err)
			}
			d, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(selector))
			if err != nil {
				t.Fatal(err)
			}
			field := d.(protoreflect.FieldDescriptor)
			if (predicate == "present" && !field.HasPresence()) ||
				(predicate == "nonempty" && field.Kind() != protoreflect.StringKind) ||
				(predicate == "true" && field.Kind() != protoreflect.BoolKind) {
				t.Fatalf("predicate %s does not match %s", predicate, selector)
			}
		}
		for _, operator := range []string{"any_of", "all_of"} {
			if raw, ok := expression[operator]; ok {
				var children []map[string]json.RawMessage
				if err := json.Unmarshal(raw, &children); err != nil {
					t.Fatal(err)
				}
				for _, child := range children {
					visit(t, child)
				}
			}
		}
	}
	for _, feature := range registry.Features {
		t.Run(feature.Name, func(t *testing.T) {
			visit(t, feature.Activation.Request)
			visit(t, feature.Activation.Response)
		})
	}
}

func TestSemanticFixtureMessagesRoundTrip(t *testing.T) {
	data, err := os.ReadFile("../../../protocol/conformance/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases map[string][]map[string]json.RawMessage
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	cases = make(map[string][]map[string]json.RawMessage)
	for _, group := range []string{"transactions", "parameters", "lob_leases"} {
		var rows []map[string]json.RawMessage
		if err := json.Unmarshal(document[group], &rows); err != nil {
			t.Fatal(err)
		}
		cases[group] = rows
	}
	for group, rows := range cases {
		for _, row := range rows {
			var name string
			if err := json.Unmarshal(row["name"], &name); err != nil {
				t.Fatal(err)
			}
			t.Run(group+"/"+name, func(t *testing.T) {
				var message proto.Message
				var field string
				switch group {
				case "transactions":
					message, field = new(kublingv1.TransactionStatus), "status"
				case "parameters":
					message, field = new(kublingv1.Parameter), "parameter"
				case "lob_leases":
					message, field = new(kublingv1.LobReference), "reference"
				}
				if err := protojson.Unmarshal(row[field], message); err != nil {
					t.Fatal(err)
				}
				restored := message.ProtoReflect().Type().New().Interface()
				transcode(t, message, restored)
				if !proto.Equal(message, restored) {
					t.Fatal("fixture lost field presence, recursive type or value during binary round-trip")
				}
			})
		}
	}
}
