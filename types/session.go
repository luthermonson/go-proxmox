package types

import "net/http"

type Session struct {
	ApiUrl     string
	AuthTicket string
	CsrfToken  string
	AuthToken  string // Combination of user, realm, token ID and UUID
	Headers    http.Header
}
