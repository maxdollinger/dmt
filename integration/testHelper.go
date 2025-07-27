package integration

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
)

const (
	testAPIKey = "test-api-key-123"
)

var EncodedAPIKey = base64.StdEncoding.EncodeToString([]byte(testAPIKey))

func JSONRequestWithApiKey(mehtod string, url string, json []byte) *http.Request {
	req := httptest.NewRequest(mehtod, url, bytes.NewBuffer(json))
	SetAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	return req
}

func SetAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+EncodedAPIKey)
}

func stringPtr(s string) *string {
	return &s
}
