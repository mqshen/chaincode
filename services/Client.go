package services

import (
	"net/http"
	"fmt"
	"bytes"
	"net/url"
	"strings"
	"net"
	"strconv"
	"github.com/hyperledger/fabric-cop/util"
	"github.com/hyperledger/fabric-cop/lib/tls"

	"path/filepath"
	"io/ioutil"
	"encoding/json"
	cfsslapi "github.com/cloudflare/cfssl/api"

	"github.com/cloudflare/cfssl/csr"
	"os"
	"encoding/base64"
)

const (
	// defaultServerPort is the default CFSSL listening port
	defaultServerPort = "8888"
	clientConfigFile  = "cop_client.json"
)

// CSRInfo is Certificate Signing Request information
type CSRInfo struct {
	CN           string               `json:"CN"`
	Names        []csr.Name           `json:"names,omitempty"`
	Hosts        []string             `json:"hosts,omitempty"`
	KeyRequest   *csr.BasicKeyRequest `json:"key,omitempty"`
	CA           *csr.CAConfig        `json:"ca,omitempty"`
	SerialNumber string               `json:"serial_number,omitempty"`
}

// Client is the COP client object
type Client struct {
	// ServerURL is the URL of the server
	ServerURL string `json:"serverURL,omitempty"`
	// HomeDir is the home directory
	HomeDir string `json:"homeDir,omitempty"`
}

// NewPost create a new post request
func (c *Client) NewPost(endpoint string, reqBody []byte) (*http.Request, error) {
	curl, cerr := c.getURL(endpoint)
	if cerr != nil {
		return nil, cerr
	}
	req, err := http.NewRequest("POST", curl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("Failed posting to %s: %s", curl, err)
	}
	return req, nil
}

// SendPost sends a request to the LDAP server and returns a response
func (c *Client) SendPost(req *http.Request) (interface{}, error) {
	reqStr := util.HTTPRequestToString(req)

	configFile, err := c.getClientConfig(c.HomeDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to load client config file [%s]; not sending\n%s", err, reqStr)
	}

	var cfg = new(tls.ClientTLSConfig)

	err = json.Unmarshal(configFile, cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse client config file [%s]; not sending\n%s", err, reqStr)
	}

	tlsConfig, err := tls.GetClientTLSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to get client TLS config [%s]; not sending\n%s", err, reqStr)
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{Transport: tr}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST failure [%s]; not sending\n%s", err, reqStr)
	}
	var respBody []byte
	if resp.Body != nil {
		respBody, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("Failed to read response [%s] of request:\n%s", err, reqStr)
		}
	}
	var body *cfsslapi.Response
	if respBody != nil && len(respBody) > 0 {
		body = new(cfsslapi.Response)
		err = json.Unmarshal(respBody, body)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse response [%s] for request:\n%s", err, reqStr)
		}
		if len(body.Errors) > 0 {
			msg := body.Errors[0].Message
			return nil, fmt.Errorf("Error response from server was '%s' for request:\n%s", msg, reqStr)
		}
	}
	scode := resp.StatusCode
	if scode >= 400 {
		return nil, fmt.Errorf("Failed with server status code %d for request:\n%s", scode, reqStr)
	}
	if body == nil {
		return nil, nil
	}
	if !body.Success {
		return nil, fmt.Errorf("Server returned failure for request:\n%s", reqStr)
	}
	return body.Result, nil
}

func (c *Client) getClientConfig(path string) ([]byte, error) {
	copClient := filepath.Join(path, clientConfigFile)
	fileBytes, err := ioutil.ReadFile(copClient)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}

func normalizeURL(addr string) (*url.URL, error) {
	addr = strings.TrimSpace(addr)
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	if u.Opaque != "" {
		u.Host = net.JoinHostPort(u.Scheme, u.Opaque)
		u.Opaque = ""
	} else if u.Path != "" && !strings.Contains(u.Path, ":") {
		u.Host = net.JoinHostPort(u.Path, defaultServerPort)
		u.Path = ""
	} else if u.Scheme == "" {
		u.Host = u.Path
		u.Path = ""
	}
	if u.Scheme != "https" {
		u.Scheme = "http"
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		_, port, err = net.SplitHostPort(u.Host + ":" + defaultServerPort)
		if err != nil {
			return nil, err
		}
	}
	if port != "" {
		_, err = strconv.Atoi(port)
		if err != nil {
			return nil, err
		}
	}
	return u, nil
}

func (c *Client) getURL(endpoint string) (string, error) {
	nurl, err := normalizeURL(c.ServerURL)
	if err != nil {
		return "", err
	}
	rtn := fmt.Sprintf("%s/api/v1/cfssl/%s", nurl, endpoint)
	return rtn, nil
}

// GenCSR generates a CSR (Certificate Signing Request)
func (c *Client) GenCSR(req *CSRInfo, id string) ([]byte, []byte, error) {

	cr := c.newCertificateRequest(req)
	cr.CN = id

	csrPEM, key, err := csr.ParseRequest(cr)
	if err != nil {
		return nil, nil, err
	}

	return csrPEM, key, nil
}

// newCertificateRequest creates a certificate request which is used to generate
// a CSR (Certificate Signing Request)
func (c *Client) newCertificateRequest(req *CSRInfo) *csr.CertificateRequest {
	cr := csr.CertificateRequest{}
	if req != nil && req.Names != nil {
		cr.Names = req.Names
	}
	if req != nil && req.Hosts != nil {
		cr.Hosts = req.Hosts
	} else {
		// Default requested hosts are local hostname
		hostname, _ := os.Hostname()
		if hostname != "" {
			cr.Hosts = make([]string, 1)
			cr.Hosts[0] = hostname
		}
	}
	if req != nil && req.KeyRequest != nil {
		cr.KeyRequest = req.KeyRequest
	}
	if req != nil {
		cr.CA = req.CA
		cr.SerialNumber = req.SerialNumber
	}
	return &cr
}

// newIdentityFromResponse returns an Identity for enroll and reenroll responses
// @param result The result from server
// @param id Name of identity being enrolled or reenrolled
// @param key The private key which was used to sign the request
func (c *Client) newIdentityFromResponse(result interface{}, id string, key []byte) (*Identity, error) {
	certByte, err := base64.StdEncoding.DecodeString(result.(string))
	if err != nil {
		return nil, fmt.Errorf("Invalid response format from server: %s", err)
	}
	return newIdentity(c, id, key, certByte), nil
}