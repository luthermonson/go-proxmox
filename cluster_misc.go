package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// Miscellaneous /cluster/* wrappers: cluster-wide log, the
// notifications/endpoints diridx, and the GET single-rule helper on a
// firewall security group.

// Log returns the cluster-wide task log. max caps the number of entries; 0
// uses the PVE default.
//
// GET /cluster/log
func (cl *Cluster) Log(ctx context.Context, max int) (entries []*ClusterLogEntry, err error) {
	path := "/cluster/log"
	if max > 0 {
		q := url.Values{}
		q.Set("max", strconv.Itoa(max))
		path = path + "?" + q.Encode()
	}
	err = cl.client.Get(ctx, path, &entries)
	return
}

// NotificationEndpointsSubdirs enumerates the children of
// /cluster/notifications/endpoints ("sendmail", "gotify", "smtp", "webhook").
// ACL-filtered. Each typed sub-resource is already covered by the per-type
// methods on *Cluster (NotificationSendmail, NotificationGotify, …).
//
// GET /cluster/notifications/endpoints
func (cl *Cluster) NotificationEndpointsSubdirs(ctx context.Context) ([]string, error) {
	return cl.notificationsDiridx(ctx, "/cluster/notifications/endpoints")
}

// FWGroupRule returns a single firewall rule in this security group by
// position. Companion to (*FirewallSecurityGroup).RuleCreate /
// RuleUpdate / RuleDelete.
//
// GET /cluster/firewall/groups/{group}/{pos}
func (g *FirewallSecurityGroup) GetRule(ctx context.Context, pos int) (rule *FirewallRule, err error) {
	if g.Group == "" {
		return nil, errors.New("firewall security group: name is required")
	}
	rule = &FirewallRule{}
	err = g.client.Get(ctx, fmt.Sprintf("/cluster/firewall/groups/%s/%d", g.Group, pos), rule)
	return
}
