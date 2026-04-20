package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("api request failed with status %d", e.StatusCode)
	}
	return fmt.Sprintf("api request failed with status %d: %s", e.StatusCode, e.Body)
}

func IsNotFoundError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

type oauthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type SecureAccessClient struct {
	clientID     string
	clientSecret string
	organization string
	baseURL      string
	tokenURL     string
	scope        string
	httpClient   *http.Client

	tokenMu     sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

func NewClient(clientID, clientSecret, organization, baseURL, tokenURL, scope string) *SecureAccessClient {
	return &SecureAccessClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		organization: organization,
		baseURL:      strings.TrimRight(baseURL, "/"),
		tokenURL:     tokenURL,
		scope:        scope,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *SecureAccessClient) token(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-1*time.Minute)) {
		return c.accessToken, nil
	}

	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	if c.scope != "" {
		values.Set("scope", c.scope)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if c.organization != "" {
		req.Header.Set("X-Umbrella-OrgId", c.organization)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var t oauthToken
	if err := json.Unmarshal(body, &t); err != nil {
		return "", fmt.Errorf("decoding token response: %w", err)
	}
	if t.AccessToken == "" {
		return "", fmt.Errorf("token response did not include access_token")
	}

	expiresIn := t.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}

	c.accessToken = t.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return c.accessToken, nil
}

func (c *SecureAccessClient) request(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	token, err := c.token(ctx)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(payload)
	}

	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

type routingPayload struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type networkTunnelGroupRequest struct {
	Name         string          `json:"name"`
	Region       string          `json:"region"`
	DeviceType   string          `json:"deviceType,omitempty"`
	AuthIDPrefix interface{}     `json:"authIdPrefix"`
	Passphrase   string          `json:"passphrase"`
	Routing      *routingPayload `json:"routing,omitempty"`
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type networkTunnelGroup struct {
	ID             int64            `json:"id"`
	Name           string           `json:"name"`
	OrganizationID int64            `json:"organizationId"`
	DeviceType     string           `json:"deviceType"`
	Region         string           `json:"region"`
	Status         string           `json:"status"`
	Routing        *routingResponse `json:"routing"`
	CreatedAt      string           `json:"createdAt"`
	ModifiedAt     string           `json:"modifiedAt"`
}

type routingResponse struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (c *SecureAccessClient) CreateNetworkTunnelGroup(ctx context.Context, reqBody networkTunnelGroupRequest) (*networkTunnelGroup, error) {
	var out networkTunnelGroup
	if err := c.request(ctx, http.MethodPost, "/networktunnelgroups", reqBody, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SecureAccessClient) GetNetworkTunnelGroup(ctx context.Context, id string) (*networkTunnelGroup, error) {
	var out networkTunnelGroup
	if err := c.request(ctx, http.MethodGet, "/networktunnelgroups/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SecureAccessClient) UpdateNetworkTunnelGroup(ctx context.Context, id string, operations []patchOperation) (*networkTunnelGroup, error) {
	var out networkTunnelGroup
	if err := c.request(ctx, http.MethodPatch, "/networktunnelgroups/"+url.PathEscape(id), operations, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SecureAccessClient) DeleteNetworkTunnelGroup(ctx context.Context, id string) error {
	return c.request(ctx, http.MethodDelete, "/networktunnelgroups/"+url.PathEscape(id), nil, nil)
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}
