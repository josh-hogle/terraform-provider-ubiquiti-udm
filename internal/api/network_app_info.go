package api

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type getNetworkAppInfoResponseError struct{}

type getNetworkAppInfoResponseSuccess struct {
	NetworkAppInfo
}

type NetworkAppInfo struct {
	// Self `json:"self"`
	// Sites `json:"sites"`
	System SystemInfo `json:"system"`
}

type SystemInfo struct {
	DeviceID     string `json:"device_id"`
	HostMetadata struct {
		ID                string `json:"id"`
		ModelAbbreviation string `json:"model_abbreviation"`
		ModelFullName     string `json:"model_fullName"`
		ModelName         string `json:"model_name"`
		ModelSysID        string `json:"model_sysid"`
		SKU               string `json:"sku"`
	} `json:"host_meta"`
	Hostname     string `json:"hostname"`
	Name         string `json:"name"`
	UnifiConsole struct {
		Type    string `json:"type"`
		Version string `json:"version"`
	} `json:"unifi_console"`
	Uptime  uint64 `json:"uptime"`
	Version string `json:"version"`
}

func (c *Client) GetNetworkAppInfo(ctx context.Context) (NetworkAppInfo, error) {
	ctx = c.addClientContext(ctx)

	// GET /proxy/network/v2/api/info
	url, req := c.newAuthenticatedRequest(ctx, "/proxy/network/v2/api/info")
	tflog.Debug(ctx, "retrieving network app info", map[string]any{
		"url": url,
	})
	apiResponseSuccess := getNetworkAppInfoResponseSuccess{}
	apiResponseError := getNetworkAppInfoResponseError{}
	resp, err := req.
		SetResult(&apiResponseSuccess).
		SetError(&apiResponseError).
		Get(url)
	if err != nil {
		tflog.Error(ctx, "failed to execute GET request", map[string]any{
			"error_message": err.Error(),
		})
		return NetworkAppInfo{}, err
	}
	statusCode, _, _ := c.logResponse(ctx, resp)

	// if a non-200 error was returned by the API, something went wrong
	if statusCode != 200 {
		tflog.Error(ctx, "failed to retrieve network info", map[string]any{
			"body": resp.Body(),
		})
		return NetworkAppInfo{}, fmt.Errorf("failed to retrieve network info")
	}

	return apiResponseSuccess.NetworkAppInfo, nil
}
