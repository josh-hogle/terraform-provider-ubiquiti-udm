package api

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type createClientDeviceResponseSuccess struct {
	Meta struct {
		RC      string `json:"rc"`
		Message string `json:"msg"`
	} `json:"meta"`
	Data []ClientDevice `json:"data"`
}

type createClientDeviceResponseError struct {
	Meta struct {
		RC      string `json:"rc"`
		Message string `json:"msg"`
	}
	Data []ClientDevice `json:"data"`
}

type getClientDevicesResponseSuccess struct {
	Meta struct {
		RC      string `json:"rc"`
		Message string `json:"msg"`
	} `json:"meta"`
	Data []ClientDevice `json:"data"`
}

type getClientDevicesResponseError struct {
	Meta struct {
		RC      string `json:"rc"`
		Message string `json:"msg"`
	}
	Data []ClientDevice `json:"data"`
}

type ClientDevice struct {
	Blocked                       bool     `json:"blocked,omitempty"`
	Confidence                    int      `json:"confidence,omitempty"`
	DeviceCategory                int      `json:"dev_cat"`
	DeviceFamily                  int      `json:"dev_family"`
	DeviceID                      int      `json:"dev_id"`
	DeviceVendor                  int      `json:"dev_vendor"`
	DisconnectTimestamp           uint64   `json:"disconnect_timestamp,omitempty"`
	FingerprintEngineVersion      string   `json:"fingerprint_engine_version,omitempty"`
	FingerprintSource             int      `json:"fingerprint_source"`
	FirstSeen                     uint64   `json:"first_seen,omitempty"`
	FixedIP                       string   `json:"fixed_ip,omitempty"`
	HardwareAddress               string   `json:"mac,omitempty"`
	Hostname                      string   `json:"hostname,omitempty"`
	ID                            string   `json:"_id"`
	IsGuest                       bool     `json:"is_guest,omitempty"`
	IsWired                       bool     `json:"is_wired,omitempty"`
	LastConnectionNetworkID       string   `json:"last_connection_network_id,omitempty"`
	LastConnectionNetworkName     string   `json:"last_connection_network_name,omitempty"`
	LastIPV6                      []string `json:"last_ipv6,omitempty"`
	LastRadio                     string   `json:"last_radio,omitempty"`
	LastSeen                      uint64   `json:"last_seen,omitempty"`
	LastUplinkHardwareAddress     string   `json:"last_uplink_mac,omitempty"`
	LastUplinkName                string   `json:"last_uplink_name,omitempty"`
	LocalDNSRecord                string   `json:"local_dns_record,omitempty"`
	LocalDNSRecordEnabled         bool     `json:"local_dns_record_enabled,omitempty"`
	Name                          string   `json:"name,omitempty"`
	Noted                         bool     `json:"noted,omitempty"`
	OSClass                       int      `json:"os_class,omitempty"`
	OSName                        int      `json:"os_name,omitempty"`
	Oui                           string   `json:"oui,omitempty"`
	SiteID                        string   `json:"site_id,omitempty"`
	UseFixedIP                    bool     `json:"use_fixedip,omitempty"`
	UsergroupID                   string   `json:"usergroup_id,omitempty"`
	VirtualNetworkOverrideEnabled bool     `json:"virtual_network_override_enabled,omitempty"`
	VirtualNetworkOverrideID      string   `json:"virtual_network_override_id"`
	WLANConfigID                  string   `json:"wlanconf_id,omitempty"`
}

func (c *Client) CreateClientDevice(ctx context.Context, device ClientDevice) (ClientDevice, error) {
	ctx = c.addClientContext(ctx)
	ctx = tflog.SetField(ctx, "hardware_address", device.HardwareAddress)

	// POST /proxy/network/api/s/:site/rest/user
	url, req := c.newAuthenticatedRequest(ctx, fmt.Sprintf("/proxy/network/api/s/%s/rest/user", c.site))
	tflog.Debug(ctx, "creating client device record", map[string]any{
		"url":    url,
		"device": device,
	})
	apiResponseSuccess := createClientDeviceResponseSuccess{}
	apiResponseError := createClientDeviceResponseError{}
	resp, err := req.
		SetBody(device).
		SetResult(&apiResponseSuccess).
		SetError(&apiResponseError).
		Post(url)
	if err != nil {
		tflog.Error(ctx, "failed to execute POST request", map[string]any{
			"error_message": err.Error(),
		})
		return ClientDevice{}, err
	}
	statusCode, _, _ := c.logResponse(ctx, resp)

	// if a non-200 error was returned by the API, something went wrong
	if statusCode != 200 {
		tflog.Error(ctx, "failed to create client device", map[string]any{
			"message":     apiResponseError.Meta.Message,
			"return_code": apiResponseError.Meta.RC,
		})
		return ClientDevice{}, fmt.Errorf("failed to create client device: %s", apiResponseError.Meta.Message)
	}
	return ClientDevice(apiResponseSuccess.Data[0]), nil
}

func (c *Client) GetClientDevices(ctx context.Context) ([]ClientDevice, error) {
	ctx = c.addClientContext(ctx)

	// GET /proxy/network/api/s/:site/rest/user
	url, req := c.newAuthenticatedRequest(ctx, fmt.Sprintf("/proxy/network/api/s/%s/rest/user", c.site))
	tflog.Debug(ctx, "retrieving client devices", map[string]any{
		"url": url,
	})
	apiResponseSuccess := getClientDevicesResponseSuccess{}
	apiResponseError := getClientDevicesResponseError{}
	resp, err := req.
		SetResult(&apiResponseSuccess).
		SetError(&apiResponseError).
		Get(url)
	if err != nil {
		tflog.Error(ctx, "failed to execute GET request", map[string]any{
			"error_message": err.Error(),
		})
		return nil, err
	}
	statusCode, _, _ := c.logResponse(ctx, resp)

	// if a non-200 error was returned by the API, something went wrong
	if statusCode != 200 {
		tflog.Error(ctx, "failed to retrieve client devices", map[string]any{
			"body": resp.Body(),
		})
		return nil, fmt.Errorf("failed to retrieve client devices")
	}
	return []ClientDevice(apiResponseSuccess.Data), nil
}

func (c *Client) GetClientDevice(ctx context.Context, id string) (ClientDevice, error) {
	ctx = tflog.SetField(ctx, "id", id)
	devices, err := c.GetClientDevices(ctx)
	if err != nil {
		return ClientDevice{}, err
	}
	ctx = c.addClientContext(ctx)

	// find the ID in question
	tflog.Debug(ctx, "searching for client device")
	for _, device := range devices {
		if device.ID == id {
			tflog.Debug(ctx, "client device was located", map[string]any{"device": device})
			return device, nil
		}
	}
	tflog.Warn(ctx, "client device not found")
	return ClientDevice{}, fmt.Errorf("no client device found with an ID of '%s'", id)
}
