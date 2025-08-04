package client

import "net/http"

// Authenticator is called to authenticate an HTTP request.
type Authenticator interface {
	// Add authentication stuff on an HTTP request
	Sign(req *http.Request)

	// Return current user credentials (oauth1 user token, oauth1 user secret)
	//Oauth1UserCredentials() (string, string)
}
