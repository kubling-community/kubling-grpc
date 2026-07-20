# Kubling Go SDK

Official Go SDK for Kubling gRPC.

The SDK provides an idiomatic Go API that abstracts the underlying gRPC protocol.

## Installation

```bash copy
go get github.com/kubling-community/kubling-grpc/sdk-go
```

# Connecting

```go
cli, err := client.NewClient(
    client.Options{
        Address:  "localhost:50051",
        Username: "sa",
        Password: "sa",
        VDBName:  "ExampleVDB",
    },
)
if err != nil {
    log.Fatal(err)
}
defer cli.Close()
```

# Querying

```go
result, err := result.Query(
    cli,
    `
    SELECT
        metadata__namespace,
        metadata__name
    FROM kube.POD
    `,
)
if err != nil {
    log.Fatal(err)
}

rows := result.Rows()
defer rows.Close()

for rows.Next() {

    namespace, err := rows.String("metadata__namespace")
    if err != nil {
        log.Fatal(err)
    }

    name, err := rows.String("metadata__name")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%s/%s\n", namespace, name)

}
```

# Reading Values

Every Kubling type has a corresponding typed accessor.

```go
rows.String("name")
rows.Char("grade")

rows.Bool("enabled")

rows.Byte("priority")
rows.Short("port")
rows.Integer("replicas")
rows.Long("size")

rows.BigInteger("counter")

rows.Float("cpu")
rows.Double("memory")
rows.Decimal("price")

rows.Date("created_date")
rows.Time("created_time")
rows.Timestamp("created_at")

rows.JSON("metadata")
```

# Executing Statements

```go
result, err := exec.Exec(
    cli,
    `
    UPDATE kube1.DEPLOYMENT
    SET spec__template__metadata__labels='{
        "mgmt.kubling.com/app" : "rabbitmq",
        "mgmt.kubling.com/platform" : "true"
    }'
    WHERE identifier='11J9JXouFxqNmSr3BRj0YBadqohUn2519842071010';
    `,
)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.AffectedRows())
```

# Transactions

```go
transaction, err := tx.Begin(cli)
if err != nil {
    log.Fatal(err)
}
defer transaction.Rollback()

_, err = transaction.Exec(`
UPDATE ...
`)
if err != nil {
    log.Fatal(err)
}

rows, err := transaction.Query(`
SELECT ...
`)
if err != nil {
    log.Fatal(err)
}

if err := transaction.Commit(); err != nil {
    log.Fatal(err)
}
```

# Generated Keys

```go
result, err := transaction.Exec(`
INSERT ...
`)
if err != nil {
    log.Fatal(err)
}

keys := result.GeneratedKeys()

rows := keys.Rows()

for rows.Next() {

    id, _ := rows.String("identifier")

    fmt.Println(id)

}
```