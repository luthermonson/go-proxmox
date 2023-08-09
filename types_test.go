package proxmox

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringOrUint64(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected StringOrUint64
		err      error
	}{
		{
			"",
			StringOrUint64(0),
			nil,
		}, {
			"0",
			StringOrUint64(0),
			nil,
		}, {
			0,
			StringOrUint64(0),
			nil,
		}, {
			"00",
			StringOrUint64(0),
			nil,
		}, {
			"0.0",
			StringOrUint64(0),
			nil,
		}, {
			"1",
			StringOrUint64(1),
			nil,
		}, {
			1,
			StringOrUint64(1),
			nil,
		}, {
			"1.0",
			StringOrUint64(1),
			nil,
		}, {
			1.0,
			StringOrUint64(1),
			nil,
		}, {
			01,
			StringOrUint64(1),
			nil,
		}, {
			"01.0",
			StringOrUint64(1),
			nil,
		}, {
			01.0,
			StringOrUint64(1),
			nil,
		}, {
			0.1,
			StringOrUint64(0),
			nil,
		}, {
			"bad-parse-1234-value", // parse error
			StringOrUint64(0),
			errors.New("failed to match ^[0-9.]*$: bad-parse-1234-value"),
		},
	}

	type s struct {
		Value StringOrUint64
	}

	for _, test := range cases {
		var value string
		switch v := test.input.(type) {
		case string:
			value = fmt.Sprintf("\"%s\"", v)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			value = fmt.Sprintf("%d", v)
		default:
			value = fmt.Sprintf("%f", v)
		}
		m := `{
	"value": ` + value + `
}
`
		var unmarshall s
		assert.Equal(t, test.err, json.Unmarshal([]byte(m), &unmarshall))
		assert.Equal(t, test.expected, unmarshall.Value)
	}
}

func TestStringOrFloat64(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected StringOrFloat64
		err      error
	}{
		{
			"",
			StringOrFloat64(0),
			nil,
		}, {
			"0",
			StringOrFloat64(0),
			nil,
		}, {
			0,
			StringOrFloat64(0),
			nil,
		}, {
			"00",
			StringOrFloat64(0),
			nil,
		}, {
			"0.0",
			StringOrFloat64(0),
			nil,
		}, {
			"1",
			StringOrFloat64(1),
			nil,
		}, {
			1,
			StringOrFloat64(1),
			nil,
		}, {
			"1.0",
			StringOrFloat64(1),
			nil,
		}, {
			1.0,
			StringOrFloat64(1),
			nil,
		}, {
			01,
			StringOrFloat64(1),
			nil,
		}, {
			"01.0",
			StringOrFloat64(1),
			nil,
		}, {
			01.0,
			StringOrFloat64(1),
			nil,
		}, {
			0.1,
			StringOrFloat64(0.1),
			nil,
		}, {
			"bad-parse-1234-value", // parse error
			StringOrFloat64(0),
			errors.New("failed to match ^[0-9.]*$: bad-parse-1234-value"),
		},
	}

	type s struct {
		Value StringOrFloat64
	}

	for _, test := range cases {
		var value string
		switch v := test.input.(type) {
		case string:
			value = fmt.Sprintf("\"%s\"", v)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			value = fmt.Sprintf("%d", v)
		default:
			value = fmt.Sprintf("%f", v)
		}
		m := `{
	"value": ` + value + `
}
`
		var unmarshall s
		assert.Equal(t, test.err, json.Unmarshal([]byte(m), &unmarshall))
		assert.Equal(t, test.expected, unmarshall.Value)
	}
}
func TestStringOrInt(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected StringOrInt
		err      error
	}{
		{
			"",
			StringOrInt(0),
			nil,
		}, {
			"0",
			StringOrInt(0),
			nil,
		}, {
			0,
			StringOrInt(0),
			nil,
		}, {
			"00",
			StringOrInt(0),
			nil,
		}, {
			"0.0",
			StringOrInt(0),
			nil,
		}, {
			"1",
			StringOrInt(1),
			nil,
		}, {
			1,
			StringOrInt(1),
			nil,
		}, {
			"1.0",
			StringOrInt(1),
			nil,
		}, {
			1.0,
			StringOrInt(1),
			nil,
		}, {
			01,
			StringOrInt(1),
			nil,
		}, {
			"01.0",
			StringOrInt(1),
			nil,
		}, {
			01.0,
			StringOrInt(1),
			nil,
		}, {
			0.1,
			StringOrInt(0),
			nil,
		}, {
			"bad-parse-1234-value", // parse error
			StringOrInt(0),
			errors.New("failed to match ^[0-9.]*$: bad-parse-1234-value"),
		},
	}

	type s struct {
		Value StringOrInt
	}

	for _, test := range cases {
		var value string
		switch v := test.input.(type) {
		case string:
			value = fmt.Sprintf("\"%s\"", v)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			value = fmt.Sprintf("%d", v)
		default:
			value = fmt.Sprintf("%f", v)
		}
		m := `{
	"value": ` + value + `
}
`
		var unmarshall s
		assert.Equal(t, test.err, json.Unmarshal([]byte(m), &unmarshall))
		assert.Equal(t, test.expected, unmarshall.Value)
	}
}
