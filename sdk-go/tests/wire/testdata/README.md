# Legacy wire baseline

`legacy-v0.1.1.binpb` is the unmodified descriptor set from the published
`sdk-go/v0.1.1` tag, without source information. It intentionally does not evolve
with the current schema. Regenerate only when deliberately adding a new baseline:

```sh
tools/bin/buf build '.git#tag=sdk-go/v0.1.1' --exclude-source-info \
  -o sdk-go/tests/wire/testdata/legacy-v0.1.1.binpb
```

The tests use an independent descriptor registry so old and new schemas with the
same fully qualified names can coexist in one process. No running engine needed.
