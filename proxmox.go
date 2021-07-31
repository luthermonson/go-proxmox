package proxmox

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/luthermonson/go-proxmox/types"
)

var (
	UserAgent = "go-proxmox/dev"
)

type Client struct {
	httpClient *http.Client
	BaseUrl    string
	Version    types.Version
	Session    types.Session
}

func NewClient(baseUrl string, opts ...Option) *Client {
	c := &Client{
		BaseUrl: baseUrl,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *Client) Login(un, pw, otf string) {

}

func (c *Client) Req(method, path string, data []byte) ([]byte, error) {
	req, err := http.NewRequest(method, path, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (c *Client) Get(p string) error {
	return c.Req("GET", p, nil)
}

func (c *Client) Post(p string, d []byte) error {
	return c.Req("POST", p, d)
}

func (c *Client) Put(p string, d []byte) error {
	return c.Req("PUT", p, d)
}

func (c *Client) Delete(p string) error {
	return c.Req("DELETE", p, nil)
}
