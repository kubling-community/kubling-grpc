# Frozen wire baselines

`legacy-v0.1.1.binpb` is the unmodified descriptor set from the historical
`sdk-go/v0.1.1` tag. It covers compatibility with the original public RPCs and
values.

`released-v1.1.1.binpb` is the unmodified descriptor set from the coordinated
`proto/v1.1.1` release. It contains `Execute` and lets the tests verify that a
v1.1.1 reader retains new terminal outcome fields as unknown wire data. Feature
negotiation remains necessary because that reader cannot interpret those fields.

Both files exclude source information and must remain frozen. Regenerate only
when deliberately adding a release baseline:

```sh
tools/bin/buf build '.git#tag=sdk-go/v0.1.1' --exclude-source-info \
  -o sdk-go/tests/wire/testdata/legacy-v0.1.1.binpb
tools/bin/buf build '.git#tag=proto/v1.1.1' --exclude-source-info \
  -o sdk-go/tests/wire/testdata/released-v1.1.1.binpb
```

The tests use independent descriptor registries so old and new schemas with the
same fully qualified names can coexist in one process. No running engine is
needed.
