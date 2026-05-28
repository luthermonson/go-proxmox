package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// ---- /cluster/ha/groups -------------------------------------------------------
// PVE marks HA groups deprecated in favor of HA rules, but they still work and
// many existing clusters use them — wrap them anyway, with the deprecation
// noted on the type. Callers building new clusters should prefer HARules.

func (cl *Cluster) HAGroups(ctx context.Context) (groups []*HAGroup, err error) {
	err = cl.client.Get(ctx, "/cluster/ha/groups", &groups)
	return
}

func (cl *Cluster) HAGroup(ctx context.Context, name string) (group *HAGroup, err error) {
	group = &HAGroup{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/ha/groups/%s", name), group)
	return
}

func (cl *Cluster) NewHAGroup(ctx context.Context, opts *HAGroupCreateOption) error {
	return cl.client.Post(ctx, "/cluster/ha/groups", opts, nil)
}

func (cl *Cluster) HAGroupUpdate(ctx context.Context, name string, opts *HAGroupUpdateOption) error {
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/ha/groups/%s", name), opts, nil)
}

func (cl *Cluster) HAGroupDelete(ctx context.Context, name string) error {
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/ha/groups/%s", name), nil)
}

// ---- /cluster/ha/resources ---------------------------------------------------

// HAResources lists managed HA resources. typ filters by resource type
// (e.g. "vm", "ct"); pass "" for all.
func (cl *Cluster) HAResources(ctx context.Context, typ string) (resources []*HAResource, err error) {
	path := "/cluster/ha/resources"
	if typ != "" {
		path += "?type=" + typ
	}
	err = cl.client.Get(ctx, path, &resources)
	return
}

func (cl *Cluster) HAResource(ctx context.Context, sid string) (resource *HAResource, err error) {
	resource = &HAResource{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/ha/resources/%s", sid), resource)
	return
}

func (cl *Cluster) NewHAResource(ctx context.Context, opts *HAResourceCreateOption) error {
	return cl.client.Post(ctx, "/cluster/ha/resources", opts, nil)
}

func (cl *Cluster) HAResourceUpdate(ctx context.Context, sid string, opts *HAResourceUpdateOption) error {
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/ha/resources/%s", sid), opts, nil)
}

// HAResourceDelete removes the resource from HA management. purge also removes
// it from any HA rules that reference it (and deletes the rule if it's the only
// resource — that's PVE's documented behavior, not a client-side decision).
func (cl *Cluster) HAResourceDelete(ctx context.Context, sid string, purge bool) error {
	path := fmt.Sprintf("/cluster/ha/resources/%s", sid)
	if purge {
		path += "?purge=1"
	}
	return cl.client.Delete(ctx, path, nil)
}

func (cl *Cluster) HAResourceMigrate(ctx context.Context, sid, node string) error {
	return cl.client.Post(ctx, fmt.Sprintf("/cluster/ha/resources/%s/migrate", sid), map[string]string{"node": node}, nil)
}

// HAResourceRelocate is the harder cousin of Migrate — it stops the service on
// the old node and restarts it on the target, rather than doing an online
// migration. Use when online migration isn't supported by the guest type.
func (cl *Cluster) HAResourceRelocate(ctx context.Context, sid, node string) error {
	return cl.client.Post(ctx, fmt.Sprintf("/cluster/ha/resources/%s/relocate", sid), map[string]string{"node": node}, nil)
}

// ---- /cluster/ha/rules -------------------------------------------------------

// HARules lists HA rules. resource filters to rules affecting that resource ID;
// typ filters by rule type (e.g. "node-affinity", "resource-affinity"). Pass
// "" for either to skip that filter.
func (cl *Cluster) HARules(ctx context.Context, resource, typ string) (rules []*HARule, err error) {
	path := "/cluster/ha/rules"
	query := ""
	if resource != "" {
		query += "resource=" + resource
	}
	if typ != "" {
		if query != "" {
			query += "&"
		}
		query += "type=" + typ
	}
	if query != "" {
		path += "?" + query
	}
	err = cl.client.Get(ctx, path, &rules)
	return
}

func (cl *Cluster) HARule(ctx context.Context, name string) (rule *HARule, err error) {
	rule = &HARule{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/ha/rules/%s", name), rule)
	return
}

func (cl *Cluster) NewHARule(ctx context.Context, opts *HARuleCreateOption) error {
	return cl.client.Post(ctx, "/cluster/ha/rules", opts, nil)
}

func (cl *Cluster) HARuleUpdate(ctx context.Context, name string, opts *HARuleUpdateOption) error {
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/ha/rules/%s", name), opts, nil)
}

func (cl *Cluster) HARuleDelete(ctx context.Context, name string) error {
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/ha/rules/%s", name), nil)
}

// ---- /cluster/ha/status ------------------------------------------------------

func (cl *Cluster) HAStatus(ctx context.Context) (status []*HAStatusEntry, err error) {
	err = cl.client.Get(ctx, "/cluster/ha/status/current", &status)
	return
}

func (cl *Cluster) HAManagerStatus(ctx context.Context) (status *HAManagerStatus, err error) {
	status = &HAManagerStatus{}
	err = cl.client.Get(ctx, "/cluster/ha/status/manager_status", status)
	return
}

// HAArm re-arms the HA stack after it was previously disarmed. Manual
// quorum-override action — requires Sys.Console on /.
//
// POST /cluster/ha/status/arm-ha
func (cl *Cluster) HAArm(ctx context.Context) error {
	return cl.client.Post(ctx, "/cluster/ha/status/arm-ha", nil, nil)
}

// HADisarm requests disarming the HA stack and releases watchdogs cluster-wide.
// resourceMode is required by PVE: "freeze" preserves HA-tracking state but
// holds commands, "ignore" removes resources from HA tracking entirely.
//
// POST /cluster/ha/status/disarm-ha
func (cl *Cluster) HADisarm(ctx context.Context, resourceMode string) error {
	if resourceMode == "" {
		return errors.New("ha disarm: resource-mode is required (\"freeze\" or \"ignore\")")
	}
	return cl.client.Post(ctx, "/cluster/ha/status/disarm-ha", map[string]string{"resource-mode": resourceMode}, nil)
}
