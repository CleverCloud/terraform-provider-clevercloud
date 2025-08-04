package client

import (
	"fmt"
	"net/http"
)

type BearerConfig struct {
	Token string
}

func (auth *BearerConfig) Sign(req *http.Request) {
	value := fmt.Sprintf("Bearer %s", auth.Token)

	req.
		Header.
		Set("Authorization", value)
}
