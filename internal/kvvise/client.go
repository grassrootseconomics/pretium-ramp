package kvvise

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type (
	KVViseOpts struct {
		Endpoint string
		APIKey   string
	}

	KVVise struct {
		endpoint   string
		apiKey     string
		httpClient *http.Client
	}

	OKResponse struct {
		Ok          bool           `json:"ok"`
		Description string         `json:"description"`
		Result      map[string]any `json:"result"`
	}

	ReverseLookupResponse struct {
		Ok          bool                `json:"ok"`
		Description string              `json:"description"`
		Result      ReverseLookupResult `json:"result"`
	}

	ReverseLookupResult struct {
		Phone string `json:"phone"`
	}

	ErrResponse struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
)

const (
	userAgent   = "pretium-ramp-go"
	contentType = "application/json"

	versionPath = "/v1/"
)

// New returns an instance of a Pretium client reusable across different products
func New(endpoint string, apiKey string) *KVVise {
	KVVise := &KVVise{
		endpoint: endpoint,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	return KVVise
}

// SetHTTPClient can be used to override the default client with a custom set one
func (kv *KVVise) SetHTTPClient(httpClient *http.Client) *KVVise {
	kv.httpClient = httpClient

	return kv
}

// setHeaders sets the headers required by the Fonbnk API
func (kv *KVVise) setHeaders(req *http.Request) (*http.Request, error) {
	if err := kv.setAuthHeaders(req); err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", contentType)
	req.Header.Set("Content-Type", contentType)

	return req, nil
}

// requestWithCtx builds the HTTP request
func (kv *KVVise) requestWithCtx(ctx context.Context, method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	return kv.do(req)
}

// do executes the built http request, setting appropriate headers
func (kv *KVVise) do(req *http.Request) (*http.Response, error) {
	builtRequest, err := kv.setHeaders(req)
	if err != nil {
		return nil, err
	}

	return kv.httpClient.Do(builtRequest)
}

// parseResponse is a general utility to decode JSON responses correctly
func parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr APIError
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&apiErr); err == nil && apiErr.Message != "" {
			if apiErr.Code == 0 {
				// Fallback to HTTP status code when API didn't set it.
				apiErr.Code = resp.StatusCode
			}
			return &apiErr
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("KVVise server error: status=%s", resp.Status)
		}
		return fmt.Errorf("KVVise server error: status=%s body=%s", resp.Status, string(b))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (kv *KVVise) ReverseLookup(ctx context.Context, address string) (*ReverseLookupResponse, error) {
	url := fmt.Sprintf("%s%slookup/reverse/%s", kv.endpoint, versionPath, address)

	resp, err := kv.requestWithCtx(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var lookupResp ReverseLookupResponse
	if err := parseResponse(resp, &lookupResp); err != nil {
		return nil, err
	}

	return &lookupResp, nil
}
