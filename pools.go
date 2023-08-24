package proxmox

import "fmt"

func (c *Client) Pools() *PoolAPI {
	poolapi := &PoolAPI{
		client: c,
	}
	return poolapi
}

func (p *PoolAPI) List() (pools Pools, err error) {
	err = p.client.Get("/pools", &pools)
	for _, pool := range pools {
		pool.client = p.client
	}
	return
}

func (p *PoolAPI) Get(name string) (pool *Pool, err error) {
	if err = p.client.Get(fmt.Sprintf("/pools/%s", name), &pool); err != nil {
		return nil, err
	}
	pool.PoolID = name
	pool.client = p.client

	return
}

func (p *PoolAPI) Create(opt *PoolCreateOption) error {
	return p.client.Post("/pools", opt, nil)
}

func (p *Pool) Update(opt *PoolUpdateOption) error {
	return p.client.Put(fmt.Sprintf("/pools/%s", p.PoolID), opt, nil)
}

func (p *Pool) Delete() error {
	return p.client.Delete(fmt.Sprintf("/pools/%s", p.PoolID), nil)
}
