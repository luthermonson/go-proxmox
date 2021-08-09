package proxmox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	log         LeveledLoggerInterface
}

func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:   baseURL,
		userAgent: DefaultUserAgent,
		log:       &LeveledLogger{Level: LevelError},
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

	c.log.Infof("SEND: %s - %s", method, path)

	var body io.Reader
	if data != nil {
		c.log.Debugf("DATA: %s", string(data))
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

	if resp.StatusCode == http.StatusInternalServerError ||
		resp.StatusCode == http.StatusNotImplemented {
		return errors.New(resp.Status)
	}

	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	c.log.Infof("RECV: %d - %s", resp.StatusCode, resp.Status)
	c.log.Debugf("BODY: %s", string(r))
	if resp.StatusCode == http.StatusBadRequest {
		var errorskey map[string]json.RawMessage
		if err := json.Unmarshal(r, &errorskey); err != nil {
			return err
		}

		if body, ok := errorskey["errors"]; ok {
			return fmt.Errorf("bad request: %s - %s", resp.Status, body)
		}

		return fmt.Errorf("bad request: %s - %s", resp.Status, string(r))
	}

	// account for everything being in a data key
	if strings.HasPrefix(string(r), "{\"data\":") {
		var datakey map[string]json.RawMessage
		if err := json.Unmarshal(r, &datakey); err != nil {
			return err
		}

		if body, ok := datakey["data"]; ok {
			return json.Unmarshal(body, &v)
		}
	}

	return json.Unmarshal(r, &v) // assume passed in type fully supports response
}

func (c *Client) Get(p string, v interface{}) error {
	return c.Req(http.MethodGet, p, nil, v)
}

func (c *Client) Post(p string, d interface{}, v interface{}) error {
	var data []byte
	if d != nil {
		var err error
		data, err = json.Marshal(d)
		if err != nil {
			return err
		}
	}

	return c.Req(http.MethodPost, p, data, v)
}

func (c *Client) Put(p string, d interface{}, v interface{}) error {
	var data []byte
	if d != nil {
		var err error
		data, err = json.Marshal(d)
		if err != nil {
			return err
		}
	}

	return c.Req(http.MethodPut, p, data, v)
}

func (c *Client) Delete(p string, v interface{}) error {
	return c.Req(http.MethodDelete, p, nil, v)
}
