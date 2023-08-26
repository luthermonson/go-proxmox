package proxmox

import (
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) NewPool(poolid, comment string) error {
	return c.Post("/pools", map[string]string{
		"poolid":  poolid,
		"comment": comment,
	}, nil)
}

func (c *Client) Pools() (pools Pools, err error) {
	err = c.Get("/pools", &pools)
	for _, pool := range pools {
		pool.client = c
	}
	return
}

// Pool optional filter of cluster resources by type, enum can be "qemu", "lxc", "storage".
func (c *Client) Pool(poolid string, filters ...string) (pool *Pool, err error) {
	u := url.URL{Path: fmt.Sprintf("/pools/%s", poolid)}

	// filters are variadic because they're optional, munging everything passed into one big string to make
	// a good request and the api will error out if there's an issue
	if f := strings.Replace(strings.Join(filters, ""), " ", "", -1); f != "" {
		params := url.Values{}
		params.Add("type", f)
		u.RawQuery = params.Encode()
	}

	if err = c.Get(u.String(), &pool); err != nil {
		return nil, err
	}
	pool.PoolID = poolid
	pool.client = c

	return
}

func (p *Pool) Update(opt *PoolUpdateOption) error {
	return p.client.Put(fmt.Sprintf("/pools/%s", p.PoolID), opt, nil)
}

func (p *Pool) Delete() error {
	return p.client.Delete(fmt.Sprintf("/pools/%s", p.PoolID), nil)
}
