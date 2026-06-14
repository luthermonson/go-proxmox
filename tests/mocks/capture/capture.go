// Package capture wires gock matchers that record details of intercepted
// requests so tests can assert on the body of multipart uploads (which gock
// can't match natively).
package capture

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	"github.com/h2non/gock"
)

// Upload holds the multipart fields of the most recent upload request.
type Upload struct {
	// Fields contains the non-file form fields keyed by field name.
	Fields map[string]string
	// Filename is the upload's "filename" form file name (the file the user
	// chose), not the form field name.
	Filename string
	// Body is the contents of the uploaded file.
	Body string
}

// LastUpload is populated by UploadMatcher each time the upload endpoint is
// hit. Reset() clears it. Tests should read it after the call under test.
var LastUpload *Upload

// Reset clears any captured state. Called from mocks.On so prior tests
// don't bleed into subsequent ones.
func Reset() { LastUpload = nil }

// UploadMatcher returns a gock matcher that drains the request body, parses
// it as multipart/form-data, and stores the result in LastUpload. It always
// reports a match so it can be combined with the URL/method matchers on a
// persistent mock.
func UploadMatcher() gock.MatchFunc {
	return func(req *http.Request, _ *gock.Request) (bool, error) {
		_, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if err != nil {
			return false, err
		}

		raw, err := io.ReadAll(req.Body)
		if err != nil {
			return false, err
		}
		// Restore the body so gock and any subsequent consumers see it.
		req.Body = io.NopCloser(bytes.NewReader(raw))

		captured := &Upload{Fields: map[string]string{}}
		mr := multipart.NewReader(bytes.NewReader(raw), params["boundary"])
		for {
			part, err := mr.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return false, err
			}
			data, err := io.ReadAll(part)
			if err != nil {
				return false, err
			}
			if part.FileName() != "" {
				captured.Filename = part.FileName()
				captured.Body = string(data)
			} else {
				captured.Fields[part.FormName()] = string(data)
			}
		}
		LastUpload = captured
		return true, nil
	}
}
