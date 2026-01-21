package proxmox

import (
	"context"
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

	if len(pools) == 0 {
		// This will not hit as the API will return a 500 error if the pool does not exist
		return nil, fmt.Errorf("pool not found")
	} else if len(pools) == 1 {
		pool = pools[0]
	} else {
		// Should be impossible to have multiple pools with the same poolid
		return nil, fmt.Errorf("multiple pools found for poolid: %s", poolid)
	}

	pool.client = c

	return
}

func (p *Pool) Update(ctx context.Context, opt *PoolUpdateOption) error {
	return p.client.Put(ctx, fmt.Sprintf("/pools/%s", p.PoolID), opt, nil)
}

func (p *Pool) Delete(ctx context.Context) error {
	return p.client.Delete(ctx, fmt.Sprintf("/pools/%s", p.PoolID), nil)
}
