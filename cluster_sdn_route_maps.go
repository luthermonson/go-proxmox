package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// SDNRouteMaps lists configured SDN route-maps. running=true returns the
// running config rather than pending.
//
// GET /cluster/sdn/route-maps
func (cl *Cluster) SDNRouteMaps(ctx context.Context, running bool) (maps []*SDNRouteMapID, err error) {
	path := "/cluster/sdn/route-maps"
	if running {
		path += "?running=1"
	}
	err = cl.client.Get(ctx, path, &maps)
	return
}

// SDNRouteMapEntries lists every route-map entry across all route-maps.
// pending/running toggle the returned configuration view.
//
// GET /cluster/sdn/route-maps/entries
func (cl *Cluster) SDNRouteMapEntries(ctx context.Context, pending, running bool) (entries []*SDNRouteMapEntry, err error) {
	path := "/cluster/sdn/route-maps/entries"
	q := url.Values{}
	if pending {
		q.Set("pending", "1")
	}
	if running {
		q.Set("running", "1")
	}
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	if err = cl.client.Get(ctx, path, &entries); err != nil {
		return nil, err
	}
	for _, e := range entries {
		e.client = cl.client
	}
	return
}

// SDNRouteMapEntriesFor lists the entries belonging to a single named
// route-map.
//
// GET /cluster/sdn/route-maps/entries/{route-map-id}
func (cl *Cluster) SDNRouteMapEntriesFor(ctx context.Context, routeMapID string, pending, running bool) (entries []*SDNRouteMapEntry, err error) {
	if routeMapID == "" {
		return nil, errors.New("route-map id is required")
	}
	path := fmt.Sprintf("/cluster/sdn/route-maps/entries/%s", routeMapID)
	q := url.Values{}
	if pending {
		q.Set("pending", "1")
	}
	if running {
		q.Set("running", "1")
	}
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	if err = cl.client.Get(ctx, path, &entries); err != nil {
		return nil, err
	}
	for _, e := range entries {
		e.client = cl.client
		e.RouteMapID = routeMapID
	}
	return
}

// SDNRouteMapEntry returns a handle for one entry in a route-map keyed by
// (route-map-id, order). No API call is made.
//
// GET /cluster/sdn/route-maps/entries/{route-map-id}/entry/{order}
func (cl *Cluster) SDNRouteMapEntry(routeMapID string, order uint16) *SDNRouteMapEntry {
	return &SDNRouteMapEntry{client: cl.client, RouteMapID: routeMapID, Order: order}
}

// NewSDNRouteMapEntry creates a new entry in a route-map. opts.RouteMapID,
// opts.Order, and opts.Action are required.
//
// POST /cluster/sdn/route-maps/entries
func (cl *Cluster) NewSDNRouteMapEntry(ctx context.Context, opts *SDNRouteMapEntryOptions) error {
	if opts == nil || opts.RouteMapID == "" {
		return errors.New("route-map id is required")
	}
	if opts.Action == "" {
		return errors.New("route-map entry action is required")
	}
	return cl.client.Post(ctx, "/cluster/sdn/route-maps/entries", opts, nil)
}

// Read populates the receiver with the route-map entry configuration.
//
// GET /cluster/sdn/route-maps/entries/{route-map-id}/entry/{order}
func (e *SDNRouteMapEntry) Read(ctx context.Context) error {
	if e.RouteMapID == "" {
		return errors.New("route-map id is required")
	}
	return e.client.Get(ctx, fmt.Sprintf("/cluster/sdn/route-maps/entries/%s/entry/%s", e.RouteMapID, strconv.FormatUint(uint64(e.Order), 10)), e)
}

// Update mutates the route-map entry.
//
// PUT /cluster/sdn/route-maps/entries/{route-map-id}/entry/{order}
func (e *SDNRouteMapEntry) Update(ctx context.Context, opts *SDNRouteMapEntryOptions) error {
	if e.RouteMapID == "" {
		return errors.New("route-map id is required")
	}
	if opts == nil {
		opts = &SDNRouteMapEntryOptions{}
	}
	return e.client.Put(ctx, fmt.Sprintf("/cluster/sdn/route-maps/entries/%s/entry/%s", e.RouteMapID, strconv.FormatUint(uint64(e.Order), 10)), opts, nil)
}

// Delete removes the route-map entry.
//
// DELETE /cluster/sdn/route-maps/entries/{route-map-id}/entry/{order}
func (e *SDNRouteMapEntry) Delete(ctx context.Context) error {
	if e.RouteMapID == "" {
		return errors.New("route-map id is required")
	}
	return e.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/route-maps/entries/%s/entry/%s", e.RouteMapID, strconv.FormatUint(uint64(e.Order), 10)), nil)
}
