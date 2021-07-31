package proxmox

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	UserAgent = "go-proxmox/dev"
)

type Client struct {
	httpClient *http.Client
	BaseURL    string
	Version    Version
	Session    Session
}

func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		BaseURL: baseURL,
	}

	for _, o := range opts {
		o(c)
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	return c
}

func (c *Client) Req(method, path string, data []byte, v interface{}) error {
	if strings.HasPrefix(path, "/") {
		path = c.BaseURL + path
	}
	var body io.Reader
	if data != nil {
		body = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	strBody := string(resBody)
	if strings.HasPrefix(strBody, "{\"data\":") && strings.HasSuffix(strBody, "}") {
		strBody = strings.TrimPrefix(strBody, "{\"data\":")
		strBody = strings.TrimSuffix(strBody, "}")
	}

	return json.Unmarshal([]byte(strBody), &v)
}

func (c *Client) Get(p string, v interface{}) error {
	return c.Req(http.MethodGet, p, nil, v)
}

func (c *Client) Post(p string, d []byte, v interface{}) error {
	return c.Req(http.MethodPost, p, d, v)
}

func (c *Client) Put(p string, d []byte, v interface{}) error {
	return c.Req(http.MethodPut, p, d, v)
}

func (c *Client) Delete(p string, v interface{}) error {
	return c.Req(http.MethodDelete, p, nil, v)
}
