package proxmox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultUserAgent = "go-proxmox/dev"
)

var ErrNotAuthorized = errors.New("not authorized to access endpoint")

func IsNotAuthorized(err error) bool {
	return err == ErrNotAuthorized
}

var ErrTimeout = errors.New("the operation has timed out")

func IsTimeout(err error) bool {
	return err == ErrTimeout
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
		if len(data) < 2048 {
			c.log.Debugf("DATA: %s", string(data))
		} else {
			c.log.Debugf("DATA: %s", "truncated due to length")
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	c.authHeaders(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		if c.credentials != nil && c.session == nil {
			// credentials passed but no session started, try a login and retry the request
			if _, err := c.Ticket(c.credentials); err != nil {
				return err
			}
			return c.Req(method, path, data, v)
		}
		return ErrNotAuthorized
	}

	return c.handleResponse(res, v)

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

// Upload - There is some weird 16kb limit hardcoded in proxmox for the max POST size, hopefully in the future we make
// a func to scp the file to the node directly as this API endopitn is kind of janky. For now big ISOs/vztmpl should
// be put somewhere and a use DownloadUrl. code link for posterity, I think they meant to do 16mb and got the bit math wrong
// https://git.proxmox.com/?p=pve-manager.git;a=blob;f=PVE/HTTPServer.pm;h=8a0c308ea6d6601b886b0dec2bada3d4c3da65d0;hb=HEAD#l36
// the task returned is the imgcopy from the tmp file to where the node actually wants the iso and you should wait for that
// to complete before using the iso
func (c *Client) Upload(path string, fields map[string]string, file *os.File, v interface{}) error {
	if strings.HasPrefix(path, "/") {
		path = c.baseURL + path
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for name, val := range fields {
		if err := w.WriteField(name, val); err != nil {
			return err
		}
	}

	if _, err := w.CreateFormFile("filename", filepath.Base(file.Name())); err != nil {
		return err
	}

	header := b.Len()
	if err := w.Close(); err != nil {
		return err
	}

	body := io.MultiReader(bytes.NewReader(b.Bytes()[:header]),
		file,
		bytes.NewReader(b.Bytes()[header:]))

	req, err := http.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return err
	}

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.ContentLength = int64(b.Len()) + fi.Size()
	c.authHeaders(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return c.handleResponse(res, &v)
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

func (c *Client) authHeaders(req *http.Request) {
	req.Header.Add("User-Agent", c.userAgent)
	req.Header.Add("Accept", "application/json")
	if c.token != "" {
		req.Header.Add("Authorization", "PVEAPIToken="+c.token)
	} else if c.session != nil {
		req.Header.Add("Cookie", "PVEAuthCookie="+c.session.Ticket)
		req.Header.Add("CSRFPreventionToken", c.session.CsrfPreventionToken)
	}
}

func (c *Client) handleResponse(res *http.Response, v interface{}) error {
	if res.StatusCode == http.StatusInternalServerError ||
		res.StatusCode == http.StatusNotImplemented {
		return errors.New(res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	c.log.Infof("RECV: %d - %s", res.StatusCode, res.Status)
	c.log.Debugf("BODY: %s", string(body))
	if res.StatusCode == http.StatusBadRequest {
		var errorskey map[string]json.RawMessage
		if err := json.Unmarshal(body, &errorskey); err != nil {
			return err
		}

		if body, ok := errorskey["errors"]; ok {
			return fmt.Errorf("bad request: %s - %s", res.Status, body)
		}

		return fmt.Errorf("bad request: %s - %s", res.Status, string(body))
	}

	// account for everything being in a data key
	var datakey map[string]json.RawMessage
	if err := json.Unmarshal(body, &datakey); err != nil {
		return err
	}

	if body, ok := datakey["data"]; ok {
		return json.Unmarshal(body, &v)
	}

	return json.Unmarshal(body, &v) // assume passed in type fully supports response
}
