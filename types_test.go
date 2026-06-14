package proxmox

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			nil, // JSON null — issue #198 (e.g. VM template returns pid: null)
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
		case nil:
			value = "null"
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
			nil, // JSON null
			StringOrFloat64(0),
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
		case nil:
			value = "null"
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
func TestIntOrBool_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		want    IntOrBool
		wantErr bool
	}{
		{"int 1 -> true", `1`, IntOrBool(true), false},
		{"int 0 -> false", `0`, IntOrBool(false), false},
		{"bool true", `true`, IntOrBool(true), false},
		{"bool false", `false`, IntOrBool(false), false},
		{"string -> error", `"x"`, IntOrBool(false), true},
		{"json null -> error", `null`, IntOrBool(false), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var v IntOrBool
			err := json.Unmarshal([]byte(tc.body), &v)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, v)
		})
	}
}

func TestIntOrBool_MarshalJSON(t *testing.T) {
	tr := IntOrBool(true)
	out, err := json.Marshal(&tr)
	require.NoError(t, err)
	assert.Equal(t, "1", string(out))

	fl := IntOrBool(false)
	out, err = json.Marshal(&fl)
	require.NoError(t, err)
	assert.Equal(t, "0", string(out))
}

func TestCSV_MarshalJSON_Nil(t *testing.T) {
	var c CSV
	out, err := json.Marshal(c)
	require.NoError(t, err)
	assert.Equal(t, "null", string(out))
}

func TestCSV_UnmarshalJSON_Error(t *testing.T) {
	var c CSV
	err := json.Unmarshal([]byte(`{"not":"valid"}`), &c)
	assert.Error(t, err)
}

func TestIsTemplate_UnmarshalJSON(t *testing.T) {
	var falsy IsTemplate
	require.NoError(t, json.Unmarshal([]byte(`""`), &falsy))
	assert.Equal(t, IsTemplate(false), falsy)

	var truthy IsTemplate
	require.NoError(t, json.Unmarshal([]byte(`"1"`), &truthy))
	assert.Equal(t, IsTemplate(true), truthy)

	var alsoTruthy IsTemplate
	require.NoError(t, json.Unmarshal([]byte(`1`), &alsoTruthy))
	assert.Equal(t, IsTemplate(true), alsoTruthy)
}

func TestFirewallLogEntry_UnmarshalJSON(t *testing.T) {
	var tuple FirewallLogEntry
	require.NoError(t, json.Unmarshal([]byte(`[42,"hello"]`), &tuple))
	assert.Equal(t, 42, tuple.LineNum)
	assert.Equal(t, "hello", tuple.Text)

	var obj FirewallLogEntry
	require.NoError(t, json.Unmarshal([]byte(`{"n":7,"t":"goodbye"}`), &obj))
	assert.Equal(t, 7, obj.LineNum)
	assert.Equal(t, "goodbye", obj.Text)

	var bad FirewallLogEntry
	assert.Error(t, json.Unmarshal([]byte(`"not json"`), &bad))
}

func TestLog_UnmarshalJSON(t *testing.T) {
	var l Log
	require.NoError(t, json.Unmarshal([]byte(`[{"n":1,"t":"first"},{"n":2,"t":"second"}]`), &l))
	assert.Equal(t, "first", l[0])
	assert.Equal(t, "second", l[1])

	var bad Log
	assert.Error(t, json.Unmarshal([]byte(`"not json"`), &bad))
}

func TestTask_UnmarshalJSON(t *testing.T) {
	body := `{
		"upid": "UPID:node1:00000001:00000001:64C0A800:test:1:root@pam:",
		"node": "node1",
		"type": "test",
		"id": "1",
		"user": "root@pam",
		"status": "stopped",
		"exitstatus": "OK",
		"starttime": 1700000000,
		"endtime": 1700000060
	}`
	var task Task
	require.NoError(t, json.Unmarshal([]byte(body), &task))
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "stopped", task.Status)
	assert.Equal(t, int64(60), int64(task.Duration.Seconds()))
	assert.False(t, task.StartTime.IsZero())
	assert.False(t, task.EndTime.IsZero())

	var bad Task
	assert.Error(t, json.Unmarshal([]byte(`"not json"`), &bad))
}

func TestCluster_UnmarshalJSON_Direct(t *testing.T) {
	body := `[
		{"type": "cluster", "id": "cluster", "name": "pve", "version": 3, "quorate": 1},
		{"type": "node", "name": "node1", "level": "", "online": 1, "id": "node/node1"},
		{"type": "node", "name": "node2", "level": "x", "online": 0, "id": "node/node2"}
	]`
	var cl Cluster
	require.NoError(t, json.Unmarshal([]byte(body), &cl))
	assert.Equal(t, "cluster", cl.ID)
	assert.Equal(t, "pve", cl.Name)
	assert.Equal(t, 3, cl.Version)
	assert.Equal(t, 1, cl.Quorate)
	assert.Len(t, cl.Nodes, 2)

	var bad Cluster
	assert.Error(t, json.Unmarshal([]byte(`"not json"`), &bad))
}

func TestStorage_UnmarshalJSON(t *testing.T) {
	body := `{
		"storage": "local",
		"node": "node1",
		"type": "dir",
		"content": "iso,vztmpl",
		"enabled": 1,
		"active": 1,
		"shared": 0,
		"used_fraction": 0.42,
		"avail": 1234567890,
		"used": 100,
		"total": 1234567990
	}`
	var s Storage
	require.NoError(t, json.Unmarshal([]byte(body), &s))
	assert.Equal(t, "local", s.Name)
	assert.Equal(t, "local", s.Storage)
	assert.Equal(t, "node1", s.Node)
	assert.Equal(t, 1, s.Enabled)
	assert.Equal(t, 0.42, s.UsedFraction)
	assert.Equal(t, uint64(100), s.Used)
	assert.Equal(t, uint64(1234567890), s.Avail)
	assert.Equal(t, uint64(1234567990), s.Total)

	// Scientific notation values — exercises the Float64 fallback branches.
	bodyBig := `{"storage":"big","total":1.0e18,"avail":1.5e17,"used":2e17}`
	var big Storage
	require.NoError(t, json.Unmarshal([]byte(bodyBig), &big))
	assert.Equal(t, uint64(1e18), big.Total)
	assert.Equal(t, uint64(1.5e17), big.Avail)
	assert.Equal(t, uint64(2e17), big.Used)

	var bad Storage
	assert.Error(t, json.Unmarshal([]byte(`"not json"`), &bad))
}

func TestFirewallRule_IsEnable(t *testing.T) {
	r := &FirewallRule{Enable: 1}
	assert.True(t, r.IsEnable())
	r.Enable = 0
	assert.False(t, r.IsEnable())
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
		}, {
			nil, // JSON null
			StringOrInt(0),
			nil,
		},
	}

	type s struct {
		Value StringOrInt
	}

	for _, test := range cases {
		var value string
		switch v := test.input.(type) {
		case nil:
			value = "null"
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
