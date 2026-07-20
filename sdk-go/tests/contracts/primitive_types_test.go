package contracts

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"

	exec "github.com/kubling-community/kubling-grpc/sdk-go/exec"
	"github.com/kubling-community/kubling-grpc/sdk-go/internal/assert"
	res "github.com/kubling-community/kubling-grpc/sdk-go/result"
)

func TestPrimitiveTypes(
	t *testing.T,
) {

	client := runtime.Client(t)

	id := uuid.NewString()

	//
	// INSERT
	//

	_, err := exec.Exec(
		client,
		fmt.Sprintf(
			`INSERT INTO TYPE_COVERAGE (

				ID,

				STRING_VALUE,
				CHAR_VALUE,

				BOOLEAN_VALUE,

				BYTE_VALUE,
				SHORT_VALUE,
				INTEGER_VALUE,
				LONG_VALUE,

				BIGINTEGER_VALUE,

				FLOAT_VALUE,
				DOUBLE_VALUE,
				DECIMAL_VALUE,

				DATE_VALUE,
				TIME_VALUE,
				TIMESTAMP_VALUE,

				JSON_VALUE

			)
			VALUES (

				'%s',

				'hello',
				'A',

				TRUE,

				7,
				1234,
				123456,
				1234567890123,

				123456789012345678901234567890,

				12.5,
				1234.56789,
				42.75,

				{d '2026-01-01'},
				{t '12:34:56'},
				{ts '2026-01-01 12:34:56'},

				'{"status":"ok","count":5}'

			)`,
			id,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	//
	// QUERY
	//

	result, err := res.Query(
		client,
		fmt.Sprintf(
			`SELECT

				STRING_VALUE,
				CHAR_VALUE,

				BOOLEAN_VALUE,

				BYTE_VALUE,
				SHORT_VALUE,
				INTEGER_VALUE,
				LONG_VALUE,

				BIGINTEGER_VALUE,

				FLOAT_VALUE,
				DOUBLE_VALUE,
				DECIMAL_VALUE,

				DATE_VALUE,
				TIME_VALUE,
				TIMESTAMP_VALUE,

				JSON_VALUE

			FROM TYPE_COVERAGE

			WHERE ID='%s'`,
			id,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = assert.AssertEquals(
		15,
		result.ColumnCount(),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = assert.AssertEquals(
		1,
		result.RowCount(),
	)
	if err != nil {
		t.Fatal(err)
	}

	rows := result.Rows()

	if !rows.Next() {
		t.Fatal("expected one row")
	}

	//
	// ASSERTS
	//

	//
	// STRING
	//

	stringValue, err := rows.String("STRING_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals("hello", stringValue); err != nil {
		t.Fatal(err)
	}

	//
	// CHAR
	//

	charValue, err := rows.Char("CHAR_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals("A", charValue); err != nil {
		t.Fatal(err)
	}

	//
	// BOOLEAN
	//

	booleanValue, err := rows.Bool("BOOLEAN_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(true, booleanValue); err != nil {
		t.Fatal(err)
	}

	//
	// BYTE
	//

	byteValue, err := rows.Byte("BYTE_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(int8(7), byteValue); err != nil {
		t.Fatal(err)
	}

	//
	// SHORT
	//

	shortValue, err := rows.Short("SHORT_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(int16(1234), shortValue); err != nil {
		t.Fatal(err)
	}

	//
	// INTEGER
	//

	integerValue, err := rows.Integer("INTEGER_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(int32(123456), integerValue); err != nil {
		t.Fatal(err)
	}

	//
	// LONG
	//

	longValue, err := rows.Long("LONG_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(int64(1234567890123), longValue); err != nil {
		t.Fatal(err)
	}

	//
	// BIGINTEGER
	//

	bigIntegerValue, err := rows.BigInteger("BIGINTEGER_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	expectedBigInteger := new(big.Int)
	expectedBigInteger.SetString(
		"123456789012345678901234567890",
		10,
	)

	if bigIntegerValue.Cmp(expectedBigInteger) != 0 {
		t.Fatalf(
			"expected %s got %s",
			expectedBigInteger.String(),
			bigIntegerValue.String(),
		)
	}

	//
	// FLOAT
	//

	floatValue, err := rows.Float("FLOAT_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if math.Abs(float64(floatValue)-12.5) > 0.000001 {
		t.Fatalf("expected 12.5 got %f", floatValue)
	}

	//
	// DOUBLE
	//

	doubleValue, err := rows.Double("DOUBLE_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	if math.Abs(doubleValue-1234.56789) > 0.000001 {
		t.Fatalf("expected 1234.56789 got %f", doubleValue)
	}

	//
	// DECIMAL
	//

	decimalValue, err := rows.Decimal("DECIMAL_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	expectedDecimal := big.NewFloat(42.75)

	if decimalValue.Cmp(expectedDecimal) != 0 {
		t.Fatalf(
			"expected %v got %v",
			expectedDecimal,
			decimalValue,
		)
	}

	//
	// DATE
	//

	dateValue, err := rows.Date("DATE_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	expectedDate := time.Date(
		2026,
		time.January,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	if !dateValue.Equal(expectedDate) {
		t.Fatalf(
			"expected %v got %v",
			expectedDate,
			dateValue,
		)
	}

	//
	// TIME
	//

	timeValue, err := rows.Time("TIME_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	expectedTime := time.Date(
		0,
		time.January,
		1,
		12,
		34,
		56,
		0,
		time.UTC,
	)

	if timeValue.Hour() != expectedTime.Hour() ||
		timeValue.Minute() != expectedTime.Minute() ||
		timeValue.Second() != expectedTime.Second() {

		t.Fatalf(
			"expected %02d:%02d:%02d got %02d:%02d:%02d",
			expectedTime.Hour(),
			expectedTime.Minute(),
			expectedTime.Second(),
			timeValue.Hour(),
			timeValue.Minute(),
			timeValue.Second(),
		)

	}

	//
	// TIMESTAMP
	//

	timestampValue, err := rows.Timestamp("TIMESTAMP_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	expectedTimestamp := time.Date(
		2026,
		time.January,
		1,
		12,
		34,
		56,
		0,
		time.UTC,
	)

	if !timestampValue.Equal(expectedTimestamp) {
		t.Fatalf(
			"expected %v got %v",
			expectedTimestamp,
			timestampValue,
		)
	}

	//
	// JSON
	//

	jsonValue, err := rows.JSON("JSON_VALUE")
	if err != nil {
		t.Fatal(err)
	}

	var payload map[string]any

	err = json.Unmarshal(
		[]byte(jsonValue),
		&payload,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(
		"ok",
		payload["status"],
	); err != nil {
		t.Fatal(err)
	}

	if err := assert.AssertEquals(
		float64(5),
		payload["count"],
	); err != nil {
		t.Fatal(err)
	}

}
