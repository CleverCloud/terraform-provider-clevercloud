package client

// Using works of https://github.com/klaidas/go-oauth1/
// But using a HMAC-SHA512 algorithm

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// OAuth1Config own credentials to contact CleverCloud API.
type OAuth1Config struct {
	ConsumerKey    string `json:"-"`
	ConsumerSecret string `json:"-"`
	AccessToken    string `json:"token"`
	AccessSecret   string `json:"secret"`
}

// Sign an HTTP request with the given OAuth1 signature.
func (auth *OAuth1Config) Sign(req *http.Request) {
	if auth == nil {
		return
	}

	authHeader := auth.buildOAuth1Header(req.Method, req.URL.String(), map[string]string{})
	req.Header.Set("Authorization", authHeader)
}

// Params being any key-value url query parameter pairs.
func (auth OAuth1Config) buildOAuth1Header(method, path string, params map[string]string) string {
	vals := url.Values{}
	vals.Add("oauth_nonce", auth.generateNonce())
	vals.Add("oauth_consumer_key", auth.ConsumerKey)
	vals.Add("oauth_signature_method", "HMAC-SHA512")
	vals.Add("oauth_timestamp", strconv.Itoa(int(time.Now().Unix())))
	vals.Add("oauth_token", auth.AccessToken)
	vals.Add("oauth_version", "1.0")

	for k, v := range params {
		vals.Add(k, v)
	}
	// net/url package QueryEscape escapes " " into "+", this replaces it with the percentage encoding of " "
	parameterString := strings.ReplaceAll(vals.Encode(), "+", "%20")

	// Calculating Signature Base String and Signing Key
	signatureBase := strings.ToUpper(method) + "&" + url.QueryEscape(strings.Split(path, "?")[0]) + "&" + url.QueryEscape(parameterString)
	signingKey := url.QueryEscape(auth.ConsumerSecret) + "&" + url.QueryEscape(auth.AccessSecret)
	signature := auth.calculateSignature(signatureBase, signingKey)

	return "OAuth oauth_consumer_key=\"" + url.QueryEscape(vals.Get("oauth_consumer_key")) + "\", oauth_nonce=\"" + url.QueryEscape(vals.Get("oauth_nonce")) +
		"\", oauth_signature=\"" + url.QueryEscape(signature) + "\", oauth_signature_method=\"" + url.QueryEscape(vals.Get("oauth_signature_method")) +
		"\", oauth_timestamp=\"" + url.QueryEscape(vals.Get("oauth_timestamp")) + "\", oauth_token=\"" + url.QueryEscape(vals.Get("oauth_token")) +
		"\", oauth_version=\"" + url.QueryEscape(vals.Get("oauth_version")) + "\""
}

func (auth OAuth1Config) calculateSignature(base, key string) string {
	hash := hmac.New(sha512.New, []byte(key))
	hash.Write([]byte(base))
	signature := hash.Sum(nil)

	return base64.StdEncoding.EncodeToString(signature)
}

const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const NONCE_SIZE = 48

func (auth OAuth1Config) generateNonce() string {
	b := make([]byte, NONCE_SIZE)
	for i := range b {
		r, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allowed))))
		b[i] = allowed[r.Int64()]
	}

	return string(b)
}

func (auth OAuth1Config) Oauth1UserCredentials() (string, string) {
	return auth.AccessToken, auth.AccessSecret
}
