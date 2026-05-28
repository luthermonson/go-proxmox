package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// SDNPrefixLists lists configured prefix-lists. pending/running toggle which
// configuration is returned; verbose=false returns just IDs.
//
// GET /cluster/sdn/prefix-lists
func (cl *Cluster) SDNPrefixLists(ctx context.Context, pending, running, verbose bool) (lists []*SDNPrefixList, err error) {
	path := "/cluster/sdn/prefix-lists"
	q := url.Values{}
	if pending {
		q.Set("pending", "1")
	}
	if running {
		q.Set("running", "1")
	}
	if verbose {
		q.Set("verbose", "1")
	}
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	if err = cl.client.Get(ctx, path, &lists); err != nil {
		return nil, err
	}
	for _, l := range lists {
		l.client = cl.client
	}
	return
}

// SDNPrefixList returns a handle for a single prefix-list. No API call is made.
//
// GET /cluster/sdn/prefix-lists/{id}
func (cl *Cluster) SDNPrefixList(id string) *SDNPrefixList {
	return &SDNPrefixList{client: cl.client, ID: id}
}

// NewSDNPrefixList creates a new prefix-list. opts.ID is required.
//
// POST /cluster/sdn/prefix-lists
func (cl *Cluster) NewSDNPrefixList(ctx context.Context, opts *SDNPrefixListOptions) error {
	if opts == nil || opts.ID == "" {
		return errors.New("sdn prefix-list id is required")
	}
	return cl.client.Post(ctx, "/cluster/sdn/prefix-lists", opts, nil)
}

// Read populates the receiver with the prefix-list configuration including
// entries.
//
// GET /cluster/sdn/prefix-lists/{id}
func (l *SDNPrefixList) Read(ctx context.Context) error {
	if l.ID == "" {
		return errors.New("sdn prefix-list id is required")
	}
	return l.client.Get(ctx, fmt.Sprintf("/cluster/sdn/prefix-lists/%s", l.ID), l)
}

// Update mutates the prefix-list. Pass a fresh Entries slice to replace the
// current set.
//
// PUT /cluster/sdn/prefix-lists/{id}
func (l *SDNPrefixList) Update(ctx context.Context, opts *SDNPrefixListOptions) error {
	if l.ID == "" {
		return errors.New("sdn prefix-list id is required")
	}
	if opts == nil {
		opts = &SDNPrefixListOptions{}
	}
	return l.client.Put(ctx, fmt.Sprintf("/cluster/sdn/prefix-lists/%s", l.ID), opts, nil)
}

// Delete removes the prefix-list.
//
// DELETE /cluster/sdn/prefix-lists/{id}
func (l *SDNPrefixList) Delete(ctx context.Context) error {
	if l.ID == "" {
		return errors.New("sdn prefix-list id is required")
	}
	return l.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/prefix-lists/%s", l.ID), nil)
}

// ListEntries lists the entries within this prefix-list.
//
// GET /cluster/sdn/prefix-lists/{id}/entries
func (l *SDNPrefixList) ListEntries(ctx context.Context) (entries []*SDNPrefixListEntry, err error) {
	if l.ID == "" {
		return nil, errors.New("sdn prefix-list id is required")
	}
	if err = l.client.Get(ctx, fmt.Sprintf("/cluster/sdn/prefix-lists/%s/entries", l.ID), &entries); err != nil {
		return nil, err
	}
	for _, e := range entries {
		e.client = l.client
		e.ID = l.ID
	}
	return
}

// Entry returns a handle for a single entry in this prefix-list keyed by seq.
// No API call is made.
//
// GET /cluster/sdn/prefix-lists/{id}/entries/{url_seq}
func (l *SDNPrefixList) Entry(seq uint32) *SDNPrefixListEntry {
	return &SDNPrefixListEntry{client: l.client, ID: l.ID, Seq: seq}
}

// AddEntry creates a new entry in this prefix-list. opts.Action and opts.Prefix
// are required.
//
// POST /cluster/sdn/prefix-lists/{id}/entries
func (l *SDNPrefixList) AddEntry(ctx context.Context, opts *SDNPrefixListEntryOptions) error {
	if l.ID == "" {
		return errors.New("sdn prefix-list id is required")
	}
	if opts == nil || opts.Action == "" || opts.Prefix == "" {
		return errors.New("sdn prefix-list entry action and prefix are required")
	}
	return l.client.Post(ctx, fmt.Sprintf("/cluster/sdn/prefix-lists/%s/entries", l.ID), opts, nil)
}

// Read populates the receiver with the entry configuration.
//
// GET /cluster/sdn/prefix-lists/{id}/entries/{url_seq}
func (e *SDNPrefixListEntry) Read(ctx context.Context) error {
	if e.ID == "" {
		return errors.New("sdn prefix-list id is required")
	}
	return e.client.Get(ctx, fmt.Sprintf("/cluster/sdn/prefix-lists/%s/entries/%s", e.ID, strconv.FormatUint(uint64(e.Seq), 10)), e)
}

// Update mutates the prefix-list entry.
//
// PUT /cluster/sdn/prefix-lists/{id}/entries/{url_seq}
func (e *SDNPrefixListEntry) Update(ctx context.Context, opts *SDNPrefixListEntryOptions) error {
	if e.ID == "" {
		return errors.New("sdn prefix-list id is required")
	}
	if opts == nil {
		opts = &SDNPrefixListEntryOptions{}
	}
	return e.client.Put(ctx, fmt.Sprintf("/cluster/sdn/prefix-lists/%s/entries/%s", e.ID, strconv.FormatUint(uint64(e.Seq), 10)), opts, nil)
}

// Delete removes the prefix-list entry.
//
// DELETE /cluster/sdn/prefix-lists/{id}/entries/{url_seq}
func (e *SDNPrefixListEntry) Delete(ctx context.Context) error {
	if e.ID == "" {
		return errors.New("sdn prefix-list id is required")
	}
	return e.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/prefix-lists/%s/entries/%s", e.ID, strconv.FormatUint(uint64(e.Seq), 10)), nil)
}
