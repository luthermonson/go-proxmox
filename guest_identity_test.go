package proxmox

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGuestMeta(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want *GuestMeta
	}{
		{"empty", "", &GuestMeta{}},
		{"vm not yet written by 9.3", "creation-qemu=11.0.3,ctime=1794062726",
			&GuestMeta{CreationQEMU: "11.0.3", CTime: 1794062726}},
		{"9.3 vm", "creation-qemu=11.1.3,ctime=1795762726,rtime=1795762726284,uid=abcdefghij234567klmn",
			&GuestMeta{CreationQEMU: "11.1.3", CTime: 1795762726, RTime: 1795762726284, UID: "abcdefghij234567klmn"}},
		{"9.3 container", "ctime=1795763078,rtime=1795763078193,uid=abcdefghij234567klmn",
			&GuestMeta{CTime: 1795763078, RTime: 1795763078193, UID: "abcdefghij234567klmn"}},
		{"unknown key kept", "uid=abcdefghij234567klmn,future-key=x",
			&GuestMeta{UID: "abcdefghij234567klmn", Extra: map[string]string{"future-key": "x"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGuestMeta(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseGuestMeta_Errors(t *testing.T) {
	for _, in := range []string{"uid", "ctime=yesterday", "rtime=1.5"} {
		t.Run(in, func(t *testing.T) {
			_, err := ParseGuestMeta(in)
			assert.Error(t, err)
		})
	}
}

func TestGuestMeta_Times(t *testing.T) {
	m := &GuestMeta{CTime: 1700000000, RTime: 1764150000123}
	assert.Equal(t, time.Unix(1700000000, 0), m.CreatedAt())
	assert.Equal(t, time.UnixMilli(1764150000123), m.RegisteredAt())

	empty := &GuestMeta{}
	assert.True(t, empty.CreatedAt().IsZero())
	assert.True(t, empty.RegisteredAt().IsZero())
}

func TestCluster_UnmarshalJSON_ClusterID(t *testing.T) {
	t.Run("standalone node carries the id on its node entry", func(t *testing.T) {
		var cl Cluster
		require.NoError(t, json.Unmarshal([]byte(`[
			{"type":"node","id":"node/pve1","name":"pve1","online":1,"local":1,"cluster-id":"q2w3e4r5t6y7u2i3o4pa"}
		]`), &cl))
		assert.Equal(t, "q2w3e4r5t6y7u2i3o4pa", cl.ClusterID)
		require.Len(t, cl.Nodes, 1)
		assert.Equal(t, "q2w3e4r5t6y7u2i3o4pa", cl.Nodes[0].ClusterID)
	})

	t.Run("cluster without an id yet", func(t *testing.T) {
		var cl Cluster
		require.NoError(t, json.Unmarshal([]byte(`[
			{"type":"cluster","id":"cluster","name":"c","version":4,"quorate":1},
			{"type":"node","id":"node/pve1","name":"pve1","online":1}
		]`), &cl))
		assert.Empty(t, cl.ClusterID)
		assert.Empty(t, cl.Nodes[0].ClusterID)
	})
}

func TestContainerConfig_MetaOmittedWhenUnset(t *testing.T) {
	b, err := json.Marshal(&ContainerConfig{})
	require.NoError(t, err)
	assert.NotContains(t, string(b), `"meta"`)
}
