package client

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client is the eNom API client.
type Client struct {
	BaseURL    string
	UID        string
	PW         string
	HTTPClient *http.Client
}

// NewClient creates a new eNom API client.
func NewClient(baseURL, uid, pw string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		UID:        uid,
		PW:         pw,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ---------------------------------------------------------------------------
// XML response types
// ---------------------------------------------------------------------------

// interfaceResponse is the top-level XML response from the eNom API.
// GetDNS returns nameservers as repeated <dns> child elements.
type interfaceResponse struct {
	XMLName  xml.Name `xml:"interface-response"`
	ErrCount int      `xml:"ErrCount"`
	Err1     string   `xml:"Err1"`
	Err2     string   `xml:"Err2"`
	Err3     string   `xml:"Err3"`
	DNS      []string `xml:"dns"`
}

// ---------------------------------------------------------------------------
// Helper utilities
// ---------------------------------------------------------------------------

// splitDomain splits a domain name like "fileblaze.com" into SLD="fileblaze" and
// TLD="com", or "example.co.uk" into SLD="example" and TLD="co.uk".
// Everything after the first dot is treated as the TLD.
func splitDomain(domain string) (sld, tld string, err error) {
	idx := strings.Index(domain, ".")
	if idx < 1 || idx == len(domain)-1 {
		return "", "", fmt.Errorf("invalid domain name %q: expected at least one dot", domain)
	}
	return domain[:idx], domain[idx+1:], nil
}

// authParams returns the base query parameters required on every eNom API call.
func (c *Client) authParams() url.Values {
	params := url.Values{}
	params.Set("uid", c.UID)
	params.Set("pw", c.PW)
	params.Set("responsetype", "xml")
	return params
}

// debugHTTP reports whether ENOM_DEBUG_HTTP is set, enabling verbose
// request/response logging to stderr. The password is always redacted.
func debugHTTP() bool {
	return os.Getenv("ENOM_DEBUG_HTTP") != ""
}

func (c *Client) redact(s string) string {
	if c.PW == "" {
		return s
	}
	return strings.ReplaceAll(s, c.PW, "[redacted]")
}

// doRequest performs a GET request to the eNom API and decodes the XML response.
func (c *Client) doRequest(params url.Values) (*interfaceResponse, error) {
	reqURL := c.BaseURL + "?" + params.Encode()
	if debugHTTP() {
		fmt.Fprintf(os.Stderr, "ENOM_DEBUG_HTTP: GET %s\n", c.redact(reqURL))
	}

	resp, err := c.HTTPClient.Get(reqURL)
	if err != nil {
		if debugHTTP() {
			fmt.Fprintf(os.Stderr, "ENOM_DEBUG_HTTP: request failed: %v\n", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if debugHTTP() {
		fmt.Fprintf(os.Stderr, "ENOM_DEBUG_HTTP: response status=%s\n", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if debugHTTP() {
		fmt.Fprintf(os.Stderr, "ENOM_DEBUG_HTTP: response body=%q\n", c.redact(string(body)))
	}

	var result interfaceResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing XML response: %w (body: %s)", err, string(body))
	}

	if result.ErrCount > 0 {
		msg := result.Err1
		if msg == "" {
			msg = fmt.Sprintf("eNom API returned %d error(s)", result.ErrCount)
		}
		return nil, fmt.Errorf("eNom API error: %s", msg)
	}

	return &result, nil
}

// ---------------------------------------------------------------------------
// GetNameservers
// ---------------------------------------------------------------------------

// GetNameservers fetches the current nameservers for the given domain name
// using the eNom GetDNS command.
func (c *Client) GetNameservers(domain string) ([]string, error) {
	sld, tld, err := splitDomain(domain)
	if err != nil {
		return nil, err
	}

	params := c.authParams()
	params.Set("command", "GetDNS")
	params.Set("SLD", sld)
	params.Set("TLD", tld)

	result, err := c.doRequest(params)
	if err != nil {
		return nil, err
	}

	var nameservers []string
	for _, v := range result.DNS {
		if v != "" {
			nameservers = append(nameservers, v)
		}
	}
	return nameservers, nil
}

// ---------------------------------------------------------------------------
// ModifyNameservers
// ---------------------------------------------------------------------------

// ModifyNameservers updates the nameservers for the given domain name.
func (c *Client) ModifyNameservers(domain string, nameservers []string) error {
	sld, tld, err := splitDomain(domain)
	if err != nil {
		return err
	}

	params := c.authParams()
	params.Set("command", "ModifyNS")
	params.Set("SLD", sld)
	params.Set("TLD", tld)

	for i, ns := range nameservers {
		params.Set(fmt.Sprintf("NS%d", i+1), ns)
	}

	_, err = c.doRequest(params)
	return err
}
