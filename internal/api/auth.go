package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type authData struct {
	CSRFToken string
	JWT       string
	User      User
}

type jwtClaims struct {
	jwt.RegisteredClaims

	UserID           string `json:"userId"`
	PasswordRevision int64  `json:"passwordRevision"`
	IsRemembered     bool   `json:"isRemembered"`
	CsrfToken        string `json:"csrfToken"`
}

type loginRequest struct {
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
	Token      string `json:"token"`
	Username   string `json:"username"`
}

type loginResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Level   string `json:"level"`
}

type loginResponseSuccess struct {
	User
}

func (c *Client) Login(ctx context.Context, username, password string) error {
	ctx = c.addClientContext(ctx)
	ctx = tflog.SetField(ctx, "username", username)

	// POST /api/auth/login
	url, req := c.newRequest(ctx, "/api/auth/login")
	tflog.Debug(ctx, "authenticating user", map[string]any{
		"url": url,
	})
	apiResponseSuccess := loginResponseSuccess{}
	apiResponseError := loginResponseError{}
	attempt := 1
	var err error
	var resp *resty.Response
	for {
		resp, err = req.
			SetBody(loginRequest{
				Username: username,
				Password: password,
			}).
			SetResult(&apiResponseSuccess).
			SetError(&apiResponseError).
			Post(url)
		if err != nil {
			tflog.Error(ctx, "failed to execute POST request", map[string]any{
				"error_message": err.Error(),
			})
			return err
		}
		statusCode, _, _ := c.logResponse(ctx, resp)

		// 200: authentication successful - we can break
		// AUTHENTICATION_FAILED_LIMIT_REACHED - might be a transient error with rate limits - retry for 30s
		// anything else is an error
		if statusCode == 200 {
			break
		}
		if apiResponseError.Code == "AUTHENTICATION_FAILED_LIMIT_REACHED" && attempt < 6 {
			time.Sleep(5 * time.Second)
			continue
		}
		tflog.Error(ctx, "failed to authenticate user with UDM server", map[string]any{
			"status_code":   statusCode,
			"error_code":    apiResponseError.Code,
			"error_message": apiResponseError.Message,
		})
		return fmt.Errorf("authentication failed: %s", apiResponseError.Message)
	}
	tflog.Debug(ctx, "authentication successful")

	// save the user's information and save the JWT token cookie
	c.authData.User = apiResponseSuccess.User
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == JWTCookieName {
			c.authData.JWT = cookie.Value
			tflog.Debug(ctx, "located JWT cookie")

			// parse CSRF token from JWT
			var claims jwtClaims
			_, err := jwt.ParseWithClaims(c.authData.JWT, &claims, nil)
			if err != nil && !errors.Is(err, jwt.ErrTokenUnverifiable) {
				tflog.Error(ctx, "failed to parse JWT", map[string]any{
					"error_message": err.Error(),
				})
				return err
			}
			c.authData.CSRFToken = claims.CsrfToken
			tflog.Debug(ctx, "extracted CSRF token from JWT")
			break
		}
	}
	if c.authData.JWT == "" {
		tflog.Error(ctx, "failed to locate JWT cookie")
		return errors.New("no JWT was returned by the server")
	}

	// collect application info
	netInfo, err := c.GetNetworkAppInfo(ctx)
	if err != nil {
		return err
	}
	c.networkAppInfo = netInfo
	return nil
}
