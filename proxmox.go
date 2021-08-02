package proxmox

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	DefaultUserAgent = "go-proxmox/dev"
)

var ErrNotAuthorized = errors.New("not authorized to access endpoint")

func IsNotAuthorized(err error) bool {
	return err == ErrNotAuthorized
}

type Client struct {
	httpClient  *http.Client
	userAgent   string
	baseURL     string
	token       string
	credentials *Credentials
	version     *Version
	session     *Session
}

func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:   baseURL,
		userAgent: DefaultUserAgent,
	}

	for _, o := range opts {
		o(c)
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	return c
}

func (c *Client) Version() (*Version, error) {
	return c.version, c.Get("/version", &c.version)
}

func (c *Client) Req(method, path string, data []byte, v interface{}) error {
	if strings.HasPrefix(path, "/") {
		path = c.baseURL + path
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

	if c.token != "" {
		req.Header.Add("Authorization", "PVEAPIToken="+c.token)
	} else if c.session != nil {
		req.Header.Add("Cookie", "PVEAuthCookie="+c.session.Ticket)
		req.Header.Add("CSRFPreventionToken", c.session.CsrfPreventionToken)
	}

	req.Header.Add("User-Agent", c.userAgent)
	req.Header.Add("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		if c.credentials != nil && c.session == nil {
			// credentials passed but no session started, try a login and retry the request
			if _, err = c.Ticket(c.credentials); err != nil {
				return err
			}
			return c.Req(method, path, data, v)
		}
		return ErrNotAuthorized
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return errors.New(resp.Status)
	}

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
