package proxmox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/buger/goterm"
	"github.com/gorilla/websocket"
)

const (
	DefaultUserAgent = "go-proxmox/dev"
	TagFormat        = "go-proxmox+%s"
)

var (
	ErrNotAuthorized = errors.New("not authorized to access endpoint")

	ErrSessionExists = errors.New("session already exists")
)

func IsNotAuthorized(err error) bool {
	return errors.Is(err, ErrNotAuthorized)
}

var ErrTimeout = errors.New("the operation has timed out")

func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

var ErrNotFound = errors.New("unable to find the item you are looking for")

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

var ErrNoop = errors.New("nothing to do")

func IsErrNoop(err error) bool {
	return errors.Is(err, ErrNoop)
}

func MakeTag(v string) string {
	return fmt.Sprintf(TagFormat, v)
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

	sessionMux sync.Mutex
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

func (c *Client) Version(ctx context.Context) (*Version, error) {
	return c.version, c.Get(ctx, "/version", &c.version)
}

func (c *Client) Req(ctx context.Context, method, path string, data []byte, v interface{}) error {
	if strings.HasPrefix(path, "/") {
		path = c.baseURL + path
	}

	c.log.Debugf("SEND: %s - %s", method, path)

	var body io.Reader
	if data != nil {
		if path != (c.baseURL + "/access/ticket") {
			// don't show passwords in the logs
			if len(data) < 2048 {
				c.log.Debugf("DATA: %s", string(data))
			} else {
				c.log.Debugf("DATA: %s", "truncated due to length")
			}
		}

		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	c.authHeaders(&req.Header)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		if path == (c.baseURL + "/access/ticket") {
			// received an unauthorized while trying to create a session
			return ErrNotAuthorized
		}

		if c.credentials != nil {
			// credentials passed but we need to check/create session
			err = c.CreateSession(ctx)
			if err != nil {
				if errors.Is(err, ErrSessionExists) {
					return ErrNotAuthorized
				}
				return err
			}
			return c.Req(ctx, method, path, data, v)
		}
		return ErrNotAuthorized
	}

	return c.handleResponse(res, v)

}

func (c *Client) Get(ctx context.Context, p string, v interface{}) error {
	return c.Req(ctx, http.MethodGet, p, nil, v)
}

// GetWithParams is a helper function to append query parameters to the URL
func (c *Client) GetWithParams(ctx context.Context, p string, d interface{}, v interface{}) error {
	// Parse data and append to URL
	if d != nil {
		queryString, err := dataParserForURL(d)
		if err != nil {
			return err
		}
		p = p + "?" + queryString
	}
	return c.Req(ctx, http.MethodGet, p, nil, v)
}

// dataParserForUrl parses the data and appends it to the URL as a query string
func dataParserForURL(d interface{}) (string, error) {
	jsonBytes, err := json.Marshal(d)
	if err != nil {
		return "", err
	}

	var m map[string]interface{}
	err = json.Unmarshal(jsonBytes, &m)
	if err != nil {
		return "", err
	}

	values := url.Values{}
	for key, value := range m {
		strValue := fmt.Sprintf("%v", value)
		values.Set(key, strValue)
	}

	return values.Encode(), nil

}

func (c *Client) Post(ctx context.Context, p string, d interface{}, v interface{}) error {
	var data []byte
	if d != nil {
		var err error
		data, err = json.Marshal(d)
		if err != nil {
			return err
		}
	}

	return c.Req(ctx, http.MethodPost, p, data, v)
}

func (c *Client) Put(ctx context.Context, p string, d interface{}, v interface{}) error {
	var data []byte
	if d != nil {
		var err error
		data, err = json.Marshal(d)
		if err != nil {
			return err
		}
	}

	return c.Req(ctx, http.MethodPut, p, data, v)
}

func (c *Client) Delete(ctx context.Context, p string, v interface{}) error {
	return c.Req(ctx, http.MethodDelete, p, nil, v)
}

// Upload - There is some weird 16kb limit hardcoded in proxmox for the max POST size, hopefully in the future we make
// a func to scp the file to the node directly as this API endpoint is kind of janky. For now big ISOs/vztmpl should
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
	c.authHeaders(&req.Header)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	return c.handleResponse(res, &v)
}

func (c *Client) authHeaders(header *http.Header) {
	header.Add("User-Agent", c.userAgent)
	header.Add("Accept", "application/json")
	if c.token != "" {
		header.Add("Authorization", "PVEAPIToken="+c.token)
	} else if c.session != nil {
		header.Add("Cookie", "PVEAuthCookie="+c.session.Ticket)
		header.Add("CSRFPreventionToken", c.session.CSRFPreventionToken)
	}
}

func (c *Client) handleResponse(res *http.Response, v interface{}) error {
	if res.StatusCode == http.StatusInternalServerError ||
		res.StatusCode == http.StatusNotImplemented {
		return errors.New(res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	c.log.Debugf("RECV: %d - %s", res.StatusCode, res.Status)

	if res.Request != nil && res.Request.URL != nil {
		if res.Request.URL.String() != (c.baseURL + "/access/ticket") {
			// don't show tokens out of the logs
			c.log.Debugf("BODY: %s", string(body))
		}
	}

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

	// if nil passed don't bother to do any unmarshalling
	if nil == v {
		return nil
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

func (c *Client) TermWebSocket(path string, term *Term) (chan []byte, chan []byte, chan error, func() error, error) {
	if strings.HasPrefix(path, "/") {
		path = strings.Replace(c.baseURL, "https://", "wss://", 1) + path
	}

	var tlsConfig *tls.Config
	transport := c.httpClient.Transport.(*http.Transport)
	if transport != nil {
		tlsConfig = transport.TLSClientConfig
	}
	c.log.Debugf("connecting to websocket: %s", path)
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 30 * time.Second,
		TLSClientConfig:  tlsConfig,
	}

	dialerHeaders := http.Header{}
	c.authHeaders(&dialerHeaders)

	conn, _, err := dialer.Dial(path, dialerHeaders)

	if err != nil {
		return nil, nil, nil, nil, err
	}

	// start the session by sending user@realm:ticket
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte(term.User+":"+term.Ticket+"\n")); err != nil {
		return nil, nil, nil, nil, err
	}

	// it sends back the same thing you just sent so catch it drop it
	_, msg, err := conn.ReadMessage()
	if err != nil || string(msg) != "OK" {
		if err := conn.Close(); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("error closing websocket: %s", err.Error())
		}
		return nil, nil, nil, nil, fmt.Errorf("unable to establish websocket: %s", err.Error())
	}

	type size struct {
		height int
		width  int
	}
	// start the session by sending user@realm:ticket
	tsize := size{
		height: goterm.Height(),
		width:  goterm.Width(),
	}

	c.log.Debugf("sending terminal size: %d x %d", tsize.height, tsize.width)
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte(fmt.Sprintf("1:%d:%d:", tsize.height, tsize.width))); err != nil {
		return nil, nil, nil, nil, err
	}

	send := make(chan []byte)
	recv := make(chan []byte)
	errs := make(chan error)
	done := make(chan struct{})
	ticker := time.NewTicker(30 * time.Second)
	resize := make(chan size)

	go func(tsize size) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				resized := size{
					height: goterm.Height(),
					width:  goterm.Width(),
				}
				if tsize.height != resized.height ||
					tsize.width != resized.width {
					tsize = resized
					resize <- resized
				}
			}
		}
	}(tsize)

	closer := func() error {
		close(done)
		time.Sleep(1 * time.Second)
		close(send)
		close(recv)
		close(errs)
		ticker.Stop()

		return conn.Close()
	}

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				_, msg, err := conn.ReadMessage()
				if err != nil {
					if strings.Contains(err.Error(), "use of closed network connection") {
						return
					}
					if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						return
					}
					errs <- err
				}
				recv <- msg
			}
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				if err := conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					errs <- err
				}
				return
			case <-ticker.C:
				c.log.Debugf("sending wss keep alive")
				if err := conn.WriteMessage(websocket.BinaryMessage, []byte("2")); err != nil {
					errs <- err
				}
			case resized := <-resize:
				c.log.Debugf("resizing terminal window: %d x %d", resized.height, resized.width)
				if err := conn.WriteMessage(websocket.BinaryMessage, []byte(fmt.Sprintf("1:%d:%d:", resized.height, resized.width))); err != nil {
					errs <- err
				}
			case msg := <-send:
				c.log.Debugf("sending: %s", msg)
				send := append([]byte(fmt.Sprintf("0:%d:", len(msg))), msg...)
				if err := conn.WriteMessage(websocket.BinaryMessage, send); err != nil {
					errs <- err
				}
			}
		}
	}()

	return send, recv, errs, closer, nil
}

func (c *Client) VNCWebSocket(path string, vnc *VNC) (chan []byte, chan []byte, chan error, func() error, error) {
	if strings.HasPrefix(path, "/") {
		path = strings.Replace(c.baseURL, "https://", "wss://", 1) + path
	}

	var tlsConfig *tls.Config
	transport := c.httpClient.Transport.(*http.Transport)
	if transport != nil {
		tlsConfig = transport.TLSClientConfig
	}
	c.log.Debugf("connecting to websocket: %s", path)
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 30 * time.Second,
		TLSClientConfig:  tlsConfig,
	}

	dialerHeaders := http.Header{}
	c.authHeaders(&dialerHeaders)

	conn, _, err := dialer.Dial(path, dialerHeaders)

	if err != nil {
		return nil, nil, nil, nil, err
	}

	send := make(chan []byte)
	recv := make(chan []byte)
	errs := make(chan error)
	done := make(chan struct{})

	closer := func() error {
		close(done)
		time.Sleep(1 * time.Second)
		close(send)
		close(recv)
		close(errs)

		return conn.Close()
	}

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				_, msg, err := conn.ReadMessage()
				if err != nil {
					if strings.Contains(err.Error(), "use of closed network connection") {
						return
					}
					if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						return
					}
					errs <- err
				}
				recv <- msg
			}
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				if err := conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					errs <- err
				}
				return
			case msg := <-send:
				c.log.Debugf("sending: %s", msg)
				if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
					errs <- err
				}
			}
		}
	}()

	return send, recv, errs, closer, nil
}
