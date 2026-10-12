package proxmox

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// refsTestURI is distinct from TestURI so the routes registered here never
// collide with the shared pve9x fixtures, and every request a test makes has
// to match a route the test itself registered.
const refsTestURI = "http://refs.test.localhost"

type recordedRequest struct {
	Method string
	Path   string
	Body   string
}

// requestRecorder is an http.RoundTripper that records every request the
// client sends, then hands it to whatever http.DefaultTransport is at that
// moment (gock's interceptor while a test runs). Reading DefaultTransport per
// request rather than at construction keeps the recorder in front of gock no
// matter when gock starts intercepting.
type requestRecorder struct {
	mu   sync.Mutex
	reqs []recordedRequest
}

func (r *requestRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := recordedRequest{Method: req.Method, Path: req.URL.Path}
	if req.Body != nil {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_ = req.Body.Close()
		rec.Body = string(raw)
		req.Body = io.NopCloser(bytes.NewReader(raw))
	}
	r.mu.Lock()
	r.reqs = append(r.reqs, rec)
	r.mu.Unlock()
	return http.DefaultTransport.RoundTrip(req)
}

func (r *requestRecorder) requests() []recordedRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]recordedRequest(nil), r.reqs...)
}

// calls renders the recorded requests as "METHOD /path" for compact asserts.
func (r *requestRecorder) calls() []string {
	var out []string
	for _, req := range r.requests() {
		out = append(out, req.Method+" "+req.Path)
	}
	return out
}

func (r *requestRecorder) reset() {
	r.mu.Lock()
	r.reqs = nil
	r.mu.Unlock()
}

// newRefsClient returns a client whose every request is recorded. gock is
// switched on before the client exists, so a request no route matches fails
// instead of leaving the process.
func newRefsClient(t *testing.T) (*Client, *requestRecorder) {
	t.Helper()
	gock.Intercept()
	gock.CleanUnmatchedRequest()
	t.Cleanup(gock.Off)
	rec := &requestRecorder{}
	return NewClient(refsTestURI, WithHTTPClient(&http.Client{Transport: rec})), rec
}

func upid(kind, id string) string {
	return `{"data": "UPID:node1:00001234:00005678:5A3B7C8D:` + kind + `:` + id + `:root@pam:"}`
}

func TestRefs_MakeNoRequests(t *testing.T) {
	client, rec := newRefsClient(t)

	node := client.NodeRef("node1")
	vm := node.VMRef(101)
	ct := node.ContainerRef(100)

	assert.Empty(t, rec.calls(), "constructing refs must not touch the API")
	assert.False(t, gock.HasUnmatchedRequest())

	assert.Equal(t, "node1", node.Name)
	assert.Same(t, client, node.client)

	assert.Equal(t, "node1", vm.Node)
	assert.Equal(t, StringOrUint64(101), vm.VMID)
	assert.Same(t, client, vm.client)
	assert.Nil(t, vm.VirtualMachineConfig)
	assert.Empty(t, vm.Status)

	assert.Equal(t, "node1", ct.Node)
	assert.Equal(t, StringOrUint64(100), ct.VMID)
	assert.Same(t, client, ct.client)
	assert.Nil(t, ct.ContainerConfig)
	assert.Empty(t, ct.Status)
}

// TestRefs_AccessorsOnUnfetchedRef pins the class-3 behaviour: the
// context-free accessors read handle fields, make no request, and report
// false (never a misleading true) until Ping has run.
func TestRefs_AccessorsOnUnfetchedRef(t *testing.T) {
	client, rec := newRefsClient(t)
	vm := client.NodeRef("node1").VMRef(101)
	ct := client.NodeRef("node1").ContainerRef(100)

	assert.False(t, vm.IsRunning())
	assert.False(t, vm.IsStopped())
	assert.False(t, vm.IsPaused())
	assert.False(t, vm.IsHibernated())
	assert.False(t, vm.HasTag("anything"))
	assert.NotPanics(t, vm.SplitTags)
	assert.Nil(t, vm.VirtualMachineConfig)

	assert.False(t, ct.HasTag("anything"))
	assert.NotPanics(t, ct.SplitTags)
	assert.Nil(t, ct.ContainerConfig)

	// Pure path builders need only node + vmid.
	assert.Equal(t, "/nodes/node1/qemu/101/mtunnelwebsocket?", vm.MigrationTunnelWebSocketPath(nil))
	assert.Equal(t, "/nodes/node1/lxc/100/mtunnelwebsocket?", ct.MigrationTunnelWebSocketPath(nil))
	assert.Equal(t, 101, vm.Snapshot("s1").VMID)
	assert.Equal(t, "node1", ct.FirewallRule(0).node)

	assert.Empty(t, rec.calls())
}

func TestVMRef_ActionsAreOneRequest(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name string
		call func(*VirtualMachine) error
		want string
	}{
		{"Start", func(v *VirtualMachine) error { _, err := v.Start(ctx); return err }, "POST /nodes/node1/qemu/101/status/start"},
		{"Stop", func(v *VirtualMachine) error { _, err := v.Stop(ctx); return err }, "POST /nodes/node1/qemu/101/status/stop"},
		{"Shutdown", func(v *VirtualMachine) error { _, err := v.Shutdown(ctx); return err }, "POST /nodes/node1/qemu/101/status/shutdown"},
		{"Reboot", func(v *VirtualMachine) error { _, err := v.Reboot(ctx); return err }, "POST /nodes/node1/qemu/101/status/reboot"},
		{"Reset", func(v *VirtualMachine) error { _, err := v.Reset(ctx); return err }, "POST /nodes/node1/qemu/101/status/reset"},
		{"Pause", func(v *VirtualMachine) error { _, err := v.Pause(ctx); return err }, "POST /nodes/node1/qemu/101/status/suspend"},
		{"Hibernate", func(v *VirtualMachine) error { _, err := v.Hibernate(ctx); return err }, "POST /nodes/node1/qemu/101/status/suspend"},
		{"Resume", func(v *VirtualMachine) error { _, err := v.Resume(ctx); return err }, "POST /nodes/node1/qemu/101/status/resume"},
		{"Config", func(v *VirtualMachine) error {
			_, err := v.Config(ctx, VirtualMachineOption{Name: "memory", Value: 2048})
			return err
		}, "POST /nodes/node1/qemu/101/config"},
		{"ConfigSync", func(v *VirtualMachine) error {
			return v.ConfigSync(ctx, VirtualMachineOption{Name: "memory", Value: 2048})
		}, "PUT /nodes/node1/qemu/101/config"},
		{"Migrate", func(v *VirtualMachine) error {
			_, err := v.Migrate(ctx, &VirtualMachineMigrateOptions{Target: "node2"})
			return err
		}, "POST /nodes/node1/qemu/101/migrate"},
		{"ResizeDisk", func(v *VirtualMachine) error { _, err := v.ResizeDisk(ctx, "scsi0", "+1G"); return err }, "PUT /nodes/node1/qemu/101/resize"},
		{"Clone", func(v *VirtualMachine) error {
			_, _, err := v.Clone(ctx, &VirtualMachineCloneOptions{NewID: 202})
			return err
		}, "POST /nodes/node1/qemu/101/clone"},
		{"ConvertToTemplate", func(v *VirtualMachine) error { _, err := v.ConvertToTemplate(ctx); return err }, "POST /nodes/node1/qemu/101/template"},
		{"NewSnapshot", func(v *VirtualMachine) error { _, err := v.NewSnapshot(ctx, "s1"); return err }, "POST /nodes/node1/qemu/101/snapshot"},
		{"AgentGetHostName", func(v *VirtualMachine) error { _, err := v.AgentGetHostName(ctx); return err }, "GET /nodes/node1/qemu/101/agent/get-host-name"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, rec := newRefsClient(t)
			gock.New(refsTestURI).Post("^/nodes/node1/qemu/101/").Persist().Reply(200).JSON(upid("qmtask", "101"))
			gock.New(refsTestURI).Put("^/nodes/node1/qemu/101/").Persist().Reply(200).JSON(upid("qmtask", "101"))
			gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/agent/get-host-name$").Persist().Reply(200).
				JSON(`{"data": {"result": {"host-name": "vm101"}}}`)

			vm := client.NodeRef("node1").VMRef(101)
			require.NoError(t, tc.call(vm))
			assert.Equal(t, []string{tc.want}, rec.calls())
		})
	}
}

func TestContainerRef_ActionsAreOneRequest(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name string
		call func(*Container) error
		want string
	}{
		{"Start", func(c *Container) error { _, err := c.Start(ctx); return err }, "POST /nodes/node1/lxc/100/status/start"},
		{"Stop", func(c *Container) error { _, err := c.Stop(ctx); return err }, "POST /nodes/node1/lxc/100/status/stop"},
		{"Shutdown", func(c *Container) error { _, err := c.Shutdown(ctx, false, 30); return err }, "POST /nodes/node1/lxc/100/status/shutdown"},
		{"Reboot", func(c *Container) error { _, err := c.Reboot(ctx); return err }, "POST /nodes/node1/lxc/100/status/reboot"},
		{"Suspend", func(c *Container) error { _, err := c.Suspend(ctx); return err }, "POST /nodes/node1/lxc/100/status/suspend"},
		{"Resume", func(c *Container) error { _, err := c.Resume(ctx); return err }, "POST /nodes/node1/lxc/100/status/resume"},
		{"Config", func(c *Container) error {
			_, err := c.Config(ctx, ContainerOption{Name: "memory", Value: 1024})
			return err
		}, "PUT /nodes/node1/lxc/100/config"},
		{"Migrate", func(c *Container) error {
			_, err := c.Migrate(ctx, &ContainerMigrateOptions{Target: "node2"})
			return err
		}, "POST /nodes/node1/lxc/100/migrate"},
		{"Resize", func(c *Container) error { _, err := c.Resize(ctx, "rootfs", "+1G"); return err }, "PUT /nodes/node1/lxc/100/resize"},
		{"Delete", func(c *Container) error { _, err := c.Delete(ctx, nil); return err }, "DELETE /nodes/node1/lxc/100"},
		{"Clone", func(c *Container) error {
			_, _, err := c.Clone(ctx, &ContainerCloneOptions{NewID: 200})
			return err
		}, "POST /nodes/node1/lxc/100/clone"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, rec := newRefsClient(t)
			gock.New(refsTestURI).Post("^/nodes/node1/lxc/100/").Persist().Reply(200).JSON(upid("vztask", "100"))
			gock.New(refsTestURI).Put("^/nodes/node1/lxc/100/").Persist().Reply(200).JSON(upid("vztask", "100"))
			gock.New(refsTestURI).Delete("^/nodes/node1/lxc/100$").Persist().Reply(200).JSON(upid("vzdestroy", "100"))

			ct := client.NodeRef("node1").ContainerRef(100)
			require.NoError(t, tc.call(ct))
			assert.Equal(t, []string{tc.want}, rec.calls())
		})
	}
}

func TestNodeRef_MethodsAreOneRequest(t *testing.T) {
	client, rec := newRefsClient(t)
	gock.New(refsTestURI).Get("^/nodes/node1/qemu$").Reply(200).
		JSON(`{"data": [{"vmid": 101, "name": "vm101", "status": "running"}]}`)

	vms, err := client.NodeRef("node1").VirtualMachines(context.Background())
	require.NoError(t, err)
	require.Len(t, vms, 1)
	assert.Equal(t, []string{"GET /nodes/node1/qemu"}, rec.calls())

	// List results come back wired to the same client and node.
	assert.Equal(t, "node1", vms[0].Node)
	assert.Same(t, client, vms[0].client)
}

// TestVMRef_Delete covers the one action that reads the config first: the
// cloud-init cleanup has to see the tag, and a ref doesn't have it.
func TestVMRef_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("no cloud-init tag", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).
			JSON(`{"data": {"name": "vm101", "tags": "prod"}}`)
		gock.New(refsTestURI).Delete("^/nodes/node1/qemu/101$").Reply(200).JSON(upid("qmdestroy", "101"))

		task, err := client.NodeRef("node1").VMRef(101).Delete(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, "qmdestroy", task.Type)
		assert.Equal(t, []string{
			"GET /nodes/node1/qemu/101/config",
			"DELETE /nodes/node1/qemu/101",
		}, rec.calls())
	})

	t.Run("cloud-init tag cleans up the ISO", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).
			JSON(`{"data": {"name": "vm101", "tags": "` + MakeTag(TagCloudInit) + `"}}`)
		// No storage accepts iso content, so the cleanup scan finds nothing
		// to delete; what matters is that it ran.
		gock.New(refsTestURI).Get("^/nodes/node1/storage$").Reply(200).
			JSON(`{"data": [{"storage": "local-lvm", "enabled": 1, "content": "images,rootdir"}]}`)
		gock.New(refsTestURI).Delete("^/nodes/node1/qemu/101$").Reply(200).JSON(upid("qmdestroy", "101"))

		_, err := client.NodeRef("node1").VMRef(101).Delete(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"GET /nodes/node1/qemu/101/config",
			"GET /nodes/node1/storage",
			"DELETE /nodes/node1/qemu/101",
		}, rec.calls())
	})

	t.Run("config error stops the delete", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(500)

		_, err := client.NodeRef("node1").VMRef(101).Delete(ctx, nil)
		require.Error(t, err)
		assert.Equal(t, []string{"GET /nodes/node1/qemu/101/config"}, rec.calls())
	})

	t.Run("fetched VM does not re-read config", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Delete("^/nodes/node1/qemu/101$").Reply(200).JSON(upid("qmdestroy", "101"))

		vm := client.NodeRef("node1").VMRef(101)
		vm.VirtualMachineConfig = &VirtualMachineConfig{Tags: "prod"}
		_, err := vm.Delete(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"DELETE /nodes/node1/qemu/101"}, rec.calls())
	})
}

func TestVMRef_Tags(t *testing.T) {
	ctx := context.Background()

	t.Run("AddTag keeps existing tags", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).
			JSON(`{"data": {"tags": "alpha;beta"}}`)
		gock.New(refsTestURI).Post("^/nodes/node1/qemu/101/config$").Reply(200).JSON(upid("qmconfig", "101"))

		vm := client.NodeRef("node1").VMRef(101)
		_, err := vm.AddTag(ctx, "gamma")
		require.NoError(t, err)
		reqs := rec.requests()
		require.Len(t, reqs, 2)
		assert.Equal(t, "GET", reqs[0].Method)
		assert.Equal(t, "POST", reqs[1].Method)
		assert.JSONEq(t, `{"tags": "alpha;beta;gamma"}`, reqs[1].Body)
		assert.True(t, vm.HasTag("gamma"))
	})

	t.Run("AddTag of a present tag is a no-op", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).
			JSON(`{"data": {"tags": "alpha;beta"}}`)

		_, err := client.NodeRef("node1").VMRef(101).AddTag(ctx, "beta")
		assert.True(t, IsErrNoop(err))
		assert.Equal(t, []string{"GET /nodes/node1/qemu/101/config"}, rec.calls())
	})

	t.Run("RemoveTag removes a tag the ref did not know about", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).
			JSON(`{"data": {"tags": "alpha;beta"}}`)
		gock.New(refsTestURI).Post("^/nodes/node1/qemu/101/config$").Reply(200).JSON(upid("qmconfig", "101"))

		_, err := client.NodeRef("node1").VMRef(101).RemoveTag(ctx, "alpha")
		require.NoError(t, err)
		reqs := rec.requests()
		require.Len(t, reqs, 2)
		assert.JSONEq(t, `{"tags": "beta"}`, reqs[1].Body)
	})

	t.Run("config error is returned, not a panic", func(t *testing.T) {
		client, _ := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(500)

		_, err := client.NodeRef("node1").VMRef(101).AddTag(ctx, "x")
		assert.Error(t, err)
		assert.False(t, IsErrNoop(err))
	})

	t.Run("UnmountCloudInitISO without the tag only reads config", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).
			JSON(`{"data": {"tags": "prod"}}`)

		require.NoError(t, client.NodeRef("node1").VMRef(101).UnmountCloudInitISO(ctx, "ide2"))
		assert.Equal(t, []string{"GET /nodes/node1/qemu/101/config"}, rec.calls())
	})

	t.Run("UnmountCloudInitISO with the tag detaches the ISO", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).
			JSON(`{"data": {"tags": "` + MakeTag(TagCloudInit) + `"}}`)
		gock.New(refsTestURI).Post("^/nodes/node1/qemu/101/config$").Reply(200).JSON(upid("qmconfig", "101"))
		gock.New(refsTestURI).Get("^/nodes/node1/storage$").Reply(200).JSON(`{"data": []}`)

		require.NoError(t, client.NodeRef("node1").VMRef(101).UnmountCloudInitISO(ctx, "ide2"))
		assert.Equal(t, []string{
			"GET /nodes/node1/qemu/101/config",
			"POST /nodes/node1/qemu/101/config",
			"GET /nodes/node1/storage",
		}, rec.calls())
	})
}

func TestContainerRef_Tags(t *testing.T) {
	ctx := context.Background()

	t.Run("AddTag keeps existing tags", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/lxc/100/config$").Reply(200).
			JSON(`{"data": {"tags": "alpha"}}`)
		gock.New(refsTestURI).Put("^/nodes/node1/lxc/100/config$").Reply(200).JSON(upid("vzconfig", "100"))

		ct := client.NodeRef("node1").ContainerRef(100)
		_, err := ct.AddTag(ctx, "beta")
		require.NoError(t, err)
		reqs := rec.requests()
		require.Len(t, reqs, 2)
		assert.Equal(t, "GET", reqs[0].Method)
		assert.Equal(t, "PUT", reqs[1].Method)
		assert.JSONEq(t, `{"tags": "alpha;beta"}`, reqs[1].Body)
		assert.Equal(t, "alpha;beta", ct.Tags)
	})

	t.Run("RemoveTag removes a tag the ref did not know about", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/lxc/100/config$").Reply(200).
			JSON(`{"data": {"tags": "alpha;beta"}}`)
		gock.New(refsTestURI).Put("^/nodes/node1/lxc/100/config$").Reply(200).JSON(upid("vzconfig", "100"))

		_, err := client.NodeRef("node1").ContainerRef(100).RemoveTag(ctx, "beta")
		require.NoError(t, err)
		reqs := rec.requests()
		require.Len(t, reqs, 2)
		assert.JSONEq(t, `{"tags": "alpha"}`, reqs[1].Body)
	})

	t.Run("config error is returned, not a panic", func(t *testing.T) {
		client, _ := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/lxc/100/config$").Reply(500)

		_, err := client.NodeRef("node1").ContainerRef(100).RemoveTag(ctx, "x")
		assert.Error(t, err)
		assert.False(t, IsErrNoop(err))
	})
}

func TestVMRef_Ping(t *testing.T) {
	ctx := context.Background()

	t.Run("fills status and config", func(t *testing.T) {
		client, rec := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/status/current$").Reply(200).
			JSON(`{"data": {"vmid": 101, "name": "vm101", "status": "running", "qmpstatus": "running", "cpus": 2}}`)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).
			JSON(`{"data": {"name": "vm101", "memory": 2048, "tags": "prod"}}`)

		vm := client.NodeRef("node1").VMRef(101)
		require.NoError(t, vm.Ping(ctx))
		assert.Equal(t, []string{
			"GET /nodes/node1/qemu/101/status/current",
			"GET /nodes/node1/qemu/101/config",
		}, rec.calls())

		assert.Equal(t, "vm101", vm.Name)
		assert.Equal(t, 2, vm.CPUs)
		assert.True(t, vm.IsRunning())
		require.NotNil(t, vm.VirtualMachineConfig)
		assert.True(t, vm.HasTag("prod"))
		// Identity and client survive the refresh.
		assert.Equal(t, "node1", vm.Node)
		assert.Equal(t, StringOrUint64(101), vm.VMID)
		assert.Same(t, client, vm.client)
	})

	t.Run("clears fields PVE stops sending", func(t *testing.T) {
		client, _ := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/status/current$").Reply(200).
			JSON(`{"data": {"vmid": 101, "status": "stopped", "lock": "suspended"}}`)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).JSON(`{"data": {}}`)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/status/current$").Reply(200).
			JSON(`{"data": {"vmid": 101, "status": "stopped"}}`)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).JSON(`{"data": {}}`)

		vm := client.NodeRef("node1").VMRef(101)
		require.NoError(t, vm.Ping(ctx))
		assert.True(t, vm.IsHibernated())
		require.NoError(t, vm.Ping(ctx))
		assert.Empty(t, vm.Lock)
		assert.False(t, vm.IsHibernated())
		assert.True(t, vm.IsStopped())
	})

	t.Run("a failed refresh leaves the handle unchanged", func(t *testing.T) {
		client, _ := newRefsClient(t)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/status/current$").Reply(200).
			JSON(`{"data": {"vmid": 101, "status": "running"}}`)
		gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(500)

		vm := client.NodeRef("node1").VMRef(101)
		require.Error(t, vm.Ping(ctx))
		assert.Empty(t, vm.Status)
		assert.Nil(t, vm.VirtualMachineConfig)
	})
}

func TestContainerRef_Ping(t *testing.T) {
	client, rec := newRefsClient(t)
	gock.New(refsTestURI).Get("^/nodes/node1/lxc/100/status/current$").Reply(200).
		JSON(`{"data": {"vmid": 100, "name": "ct100", "status": "running", "cpus": 1}}`)
	gock.New(refsTestURI).Get("^/nodes/node1/lxc/100/config$").Reply(200).
		JSON(`{"data": {"hostname": "ct100", "tags": "web"}}`)

	ct := client.NodeRef("node1").ContainerRef(100)
	require.NoError(t, ct.Ping(context.Background()))
	assert.Equal(t, []string{
		"GET /nodes/node1/lxc/100/status/current",
		"GET /nodes/node1/lxc/100/config",
	}, rec.calls())
	assert.Equal(t, "ct100", ct.Name)
	assert.Equal(t, "running", ct.Status)
	require.NotNil(t, ct.ContainerConfig)
	assert.Equal(t, "ct100", ct.ContainerConfig.Hostname)
	assert.True(t, ct.HasTag("web"))
	assert.Equal(t, "node1", ct.Node)
	assert.Equal(t, StringOrUint64(100), ct.VMID)
}

func TestNodeRef_Status(t *testing.T) {
	client, rec := newRefsClient(t)
	gock.New(refsTestURI).Get("^/nodes/node1/status$").Reply(200).
		JSON(`{"data": {"uptime": 1234, "pveversion": "pve-manager/9.0.0"}}`)

	node := client.NodeRef("node1")
	assert.Zero(t, node.Uptime)
	require.NoError(t, node.Status(context.Background()))
	assert.Equal(t, []string{"GET /nodes/node1/status"}, rec.calls())
	assert.Equal(t, uint64(1234), node.Uptime)
	assert.Equal(t, "pve-manager/9.0.0", node.PVEVersion)
	assert.Equal(t, "node1", node.Name)
}

// TestRefs_FetchAPIsUnchanged guards the other half of the contract: the
// existing fetching accessors still make the same requests they always did.
func TestRefs_FetchAPIsUnchanged(t *testing.T) {
	ctx := context.Background()
	client, rec := newRefsClient(t)
	gock.New(refsTestURI).Get("^/nodes/node1/status$").Reply(200).JSON(`{"data": {"uptime": 1}}`)
	gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/status/current$").Reply(200).
		JSON(`{"data": {"vmid": 101, "status": "running"}}`)
	gock.New(refsTestURI).Get("^/nodes/node1/qemu/101/config$").Reply(200).JSON(`{"data": {"memory": 512}}`)
	gock.New(refsTestURI).Post("^/nodes/node1/qemu/101/status/stop$").Times(2).Reply(200).JSON(upid("qmstop", "101"))

	node, err := client.Node(ctx, "node1")
	require.NoError(t, err)
	vm, err := node.VirtualMachine(ctx, 101)
	require.NoError(t, err)
	_, err = vm.Stop(ctx)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"GET /nodes/node1/status",
		"GET /nodes/node1/qemu/101/status/current",
		"GET /nodes/node1/qemu/101/config",
		"POST /nodes/node1/qemu/101/status/stop",
	}, rec.calls())
	rec.reset()

	// The same stop through refs is the POST alone.
	_, err = client.NodeRef("node1").VMRef(101).Stop(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"POST /nodes/node1/qemu/101/status/stop"}, rec.calls())
}
