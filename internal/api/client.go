package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Client struct {
	authData       authData
	cli            *resty.Client
	hostname       string
	networkAppInfo NetworkAppInfo
	site           string
}

func NewClient(hostname, site string, ignoreUntrustedSSLCert bool) *Client {
	cli := resty.New().SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: ignoreUntrustedSSLCert,
	})
	return &Client{
		cli:      cli,
		hostname: hostname,
		site:     site,
	}
}

func (c *Client) addClientContext(ctx context.Context) context.Context {
	ctx = tflog.SetField(ctx, "hostname", c.hostname)
	ctx = tflog.SetField(ctx, "site", c.site)
	ctx = tflog.SetField(ctx, "username", c.authData.User.Username)
	ctx = tflog.SetField(ctx, "system_name", c.networkAppInfo.System.Name)
	ctx = tflog.SetField(ctx, "unifi_console_version", c.networkAppInfo.System.UnifiConsole.Version)
	ctx = tflog.SetField(ctx, "network_app_version", c.networkAppInfo.System.Version)
	return ctx
}

func (c *Client) logResponse(ctx context.Context, resp *resty.Response) (int, http.Header, []byte) {
	statusCode := resp.StatusCode()
	httpHeaders := resp.Header()
	body := resp.Body()

	if strings.EqualFold(os.Getenv("TF_LOG"), "DEBUG") || strings.EqualFold(os.Getenv("TF_LOG"), "TRACE") {
		headers := map[string]any{}
		cookies := map[string]map[string]any{}
		for k, v := range resp.Header() {
			if k == "Set-Cookie" {
				continue
			}
			headers[k] = v
		}
		for _, cookie := range resp.Cookies() {
			cookies[cookie.Name] = map[string]any{
				"Value":    cookie.Value,
				"Path":     cookie.Path,
				"Domain":   cookie.Domain,
				"Expires":  cookie.Expires.Format(time.RFC3339),
				"HttpOnly": cookie.HttpOnly,
				"Secure":   cookie.Secure,
				"SameSite": cookie.SameSite,
			}
		}
		tflog.Debug(ctx, fmt.Sprintf("response received from: %s", resp.Request.URL), map[string]any{
			"status_code": statusCode,
			"headers":     headers,
			"cookies":     cookies,
			"body":        string(body),
		})
	}
	return statusCode, httpHeaders, body
}

func (c *Client) newAuthenticatedRequest(ctx context.Context, uri string) (string, *resty.Request) {
	url, req := c.newRequest(ctx, uri)
	return url,
		req.
			SetHeader("X-Csrf-Token", c.authData.CSRFToken).
			SetCookie(&http.Cookie{
				Name:     JWTCookieName,
				Value:    c.authData.JWT,
				Path:     "/",
				Domain:   c.hostname,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteNoneMode,
			})
}

func (c *Client) newRequest(_ context.Context, uri string) (string, *resty.Request) {
	return fmt.Sprintf("https://%s%s", c.hostname, uri),
		c.cli.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json")
}
