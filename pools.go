package proxmox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) NewPool(ctx context.Context, poolid, comment string) error {
	return c.Post(ctx, "/pools", map[string]string{
		"poolid":  poolid,
		"comment": comment,
	}, nil)
}

func (c *Client) Pools(ctx context.Context) (pools Pools, err error) {
	err = c.Get(ctx, "/pools", &pools)
	for _, pool := range pools {
		pool.client = c
	}
	return
}

// Pool optional filter of cluster resources by type, enum can be "qemu", "lxc", "storage".
func (c *Client) Pool(ctx context.Context, poolid string, filters ...string) (pool *Pool, err error) {
	u := url.URL{Path: "/pools/"}

	// /pools/{poolid} is deprecated as it does not support nested pools
	// so we're using the /pools/ endpoint with a query parameter as recommended by the API documentation
	params := url.Values{}
	params.Add("poolid", poolid)

	// filters are variadic because they're optional, munging everything passed into one big string to make
	// a good request and the api will error out if there's an issue
	if f := strings.ReplaceAll(strings.Join(filters, ""), " ", ""); f != "" {
		params.Add("type", f)
	}

	u.RawQuery = params.Encode()

	var pools Pools
	if err = c.Get(ctx, u.String(), &pools); err != nil {
		return nil, err
	}

	switch len(pools) {
	case 0:
		// This will not hit as the API will return a 500 error if the pool does not exist
		return nil, fmt.Errorf("pool not found")
	case 1:
		pool = pools[0]
	default:
		// Should be impossible to have multiple pools with the same poolid
		return nil, fmt.Errorf("multiple pools found for poolid: %s", poolid)
	}

	pool.client = c

	return
}

// Update modifies the pool via the non-deprecated PUT /pools endpoint
// (poolid travels in the body alongside the other params). Use this in
// preference to UpdateDeprecated so nested pools work correctly.
func (p *Pool) Update(ctx context.Context, opt *PoolUpdateOption) error {
	data, err := poolUpdatePayload(p.PoolID, opt)
	if err != nil {
		return err
	}
	return p.client.Put(ctx, "/pools", data, nil)
}

// Delete removes the pool via the non-deprecated DELETE /pools endpoint.
// Use this in preference to DeleteDeprecated.
func (p *Pool) Delete(ctx context.Context) error {
	u := url.URL{Path: "/pools"}
	q := url.Values{}
	q.Set("poolid", p.PoolID)
	u.RawQuery = q.Encode()
	return p.client.Delete(ctx, u.String(), nil)
}

// poolUpdatePayload merges opt's set fields with poolid for the PUT body.
// Going through JSON round-trip preserves opt's omitempty rules so we don't
// duplicate that logic here.
func poolUpdatePayload(poolid string, opt *PoolUpdateOption) (map[string]interface{}, error) {
	data := map[string]interface{}{"poolid": poolid}
	if opt == nil {
		return data, nil
	}
	raw, err := json.Marshal(opt)
	if err != nil {
		return nil, err
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	for k, v := range fields {
		data[k] = v
	}
	return data, nil
}

// GetDeprecated reads the pool via the deprecated GET /pools/{poolid}
// endpoint. It does not support nested pools — prefer Client.Pool, which
// uses the non-deprecated /pools?poolid= form.
//
// Deprecated: use Client.Pool.
func (p *Pool) GetDeprecated(ctx context.Context, filters ...string) (*Pool, error) {
	u := url.URL{Path: fmt.Sprintf("/pools/%s", p.PoolID)}
	if f := strings.ReplaceAll(strings.Join(filters, ""), " ", ""); f != "" {
		q := url.Values{}
		q.Set("type", f)
		u.RawQuery = q.Encode()
	}

	// The deprecated endpoint returns the pool body directly (no poolid in
	// the response payload), so we seed PoolID from the receiver.
	pool := &Pool{client: p.client, PoolID: p.PoolID}
	if err := p.client.Get(ctx, u.String(), pool); err != nil {
		return nil, err
	}
	return pool, nil
}

// UpdateDeprecated writes to the deprecated PUT /pools/{poolid} endpoint.
// It does not support nested pools.
//
// Deprecated: use Pool.Update.
func (p *Pool) UpdateDeprecated(ctx context.Context, opt *PoolUpdateOption) error {
	return p.client.Put(ctx, fmt.Sprintf("/pools/%s", p.PoolID), opt, nil)
}

// DeleteDeprecated removes the pool via the deprecated DELETE /pools/{poolid}
// endpoint. It does not support nested pools.
//
// Deprecated: use Pool.Delete.
func (p *Pool) DeleteDeprecated(ctx context.Context) error {
	return p.client.Delete(ctx, fmt.Sprintf("/pools/%s", p.PoolID), nil)
}
