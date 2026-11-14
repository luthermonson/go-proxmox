package proxmox

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clusterResourcesJSON is a minimal /cluster/resources?type=vm payload with
// one QEMU guest on node2 and one LXC guest on node3.
const clusterResourcesJSON = `{"data": [
	{"id": "qemu/101", "type": "qemu", "vmid": 101, "node": "node2", "name": "vm101", "status": "running"},
	{"id": "lxc/100", "type": "lxc", "vmid": 100, "node": "node3", "name": "ct100", "status": "running"}
]}`

func TestFindVirtualMachine(t *testing.T) {
	ctx := context.Background()

	t.Run("resolves node in one request and binds the handle", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/cluster/resources$").Reply(200).JSON(clusterResourcesJSON)

		vm, err := client.FindVirtualMachine(ctx, 101)
		require.NoError(t, err)
		assert.Equal(t, []string{"GET /cluster/resources"}, rec.calls())

		// The returned handle is bound to the resolved node, request-free.
		assert.Equal(t, "node2", vm.Node)
		assert.Equal(t, StringOrUint64(101), vm.VMID)
		assert.Same(t, client, vm.client)
		assert.Nil(t, vm.VirtualMachineConfig)
	})

	t.Run("a subsequent action targets the resolved node", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/cluster/resources$").Reply(200).JSON(clusterResourcesJSON)
		gock.New(refsTestURI).Post("^/nodes/node2/qemu/101/status/stop$").Reply(200).JSON(upid("qmstop", "101"))

		vm, err := client.FindVirtualMachine(ctx, 101)
		require.NoError(t, err)
		_, err = vm.Stop(ctx)
		require.NoError(t, err)

		assert.Equal(t, []string{
			"GET /cluster/resources",
			"POST /nodes/node2/qemu/101/status/stop",
		}, rec.calls())
	})

	t.Run("unknown vmid is not found", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/cluster/resources$").Reply(200).JSON(clusterResourcesJSON)

		vm, err := client.FindVirtualMachine(ctx, 999)
		require.Error(t, err)
		assert.Nil(t, vm)
		assert.True(t, IsNotFound(err))
		assert.Equal(t, []string{"GET /cluster/resources"}, rec.calls())
	})

	t.Run("a container vmid does not resolve as a VM", func(t *testing.T) {
		client, _ := newRefsClient(t)
		gock.New(refsTestURI).Get("^/cluster/resources$").Reply(200).JSON(clusterResourcesJSON)

		// 100 is an LXC guest, so looking it up as a VM is not found.
		_, err := client.FindVirtualMachine(ctx, 100)
		require.Error(t, err)
		assert.True(t, IsNotFound(err))
	})

	t.Run("a resources error propagates", func(t *testing.T) {
		client, _ := newRefsClient(t)
		gock.New(refsTestURI).Get("^/cluster/resources$").Reply(500)

		_, err := client.FindVirtualMachine(ctx, 101)
		require.Error(t, err)
		assert.False(t, IsNotFound(err))
	})
}

func TestFindContainer(t *testing.T) {
	ctx := context.Background()

	t.Run("resolves node in one request and binds the handle", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/cluster/resources$").Reply(200).JSON(clusterResourcesJSON)

		ct, err := client.FindContainer(ctx, 100)
		require.NoError(t, err)
		assert.Equal(t, []string{"GET /cluster/resources"}, rec.calls())

		assert.Equal(t, "node3", ct.Node)
		assert.Equal(t, StringOrUint64(100), ct.VMID)
		assert.Same(t, client, ct.client)
		assert.Nil(t, ct.ContainerConfig)
	})

	t.Run("a subsequent action targets the resolved node", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/cluster/resources$").Reply(200).JSON(clusterResourcesJSON)
		gock.New(refsTestURI).Post("^/nodes/node3/lxc/100/status/stop$").Reply(200).JSON(upid("vzstop", "100"))

		ct, err := client.FindContainer(ctx, 100)
		require.NoError(t, err)
		_, err = ct.Stop(ctx)
		require.NoError(t, err)

		assert.Equal(t, []string{
			"GET /cluster/resources",
			"POST /nodes/node3/lxc/100/status/stop",
		}, rec.calls())
	})

	t.Run("unknown vmid is not found", func(t *testing.T) {
		client, _ := newRefsClient(t)
		gock.New(refsTestURI).Get("^/cluster/resources$").Reply(200).JSON(clusterResourcesJSON)

		ct, err := client.FindContainer(ctx, 999)
		require.Error(t, err)
		assert.Nil(t, ct)
		assert.True(t, IsNotFound(err))
	})
}
