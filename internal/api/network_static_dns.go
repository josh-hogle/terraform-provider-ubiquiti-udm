package api

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type createStaticDNSEntriesResponseSuccess StaticDNSEntry

type createStaticDNSEntriesResponseError struct {
	Code      string `json:"code"`
	Details   any    `json:"details"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

type deleteStaticDNSEntriesResponseSuccess struct{}

type deleteStaticDNSEntriesResponseError struct {
	Code      string `json:"code"`
	Details   any    `json:"details"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

type getStaticDNSEntriesResponseSuccess []StaticDNSEntry

type getStaticDNSEntriesResponseError struct{}

type updateStaticDNSEntriesResponseSuccess StaticDNSEntry

type updateStaticDNSEntriesResponseError struct {
	Code      string `json:"code"`
	Details   any    `json:"details"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

type StaticDNSEntry struct {
	ID         string `json:"_id,omitempty"`
	Enabled    bool   `json:"enabled"`
	Key        string `json:"key"`
	Port       int    `json:"port,omitempty"`
	Priority   int    `json:"priority,omitempty"`
	RecordType string `json:"record_type"`
	TTL        int    `json:"ttl"`
	Value      string `json:"value"`
	Weight     int    `json:"weight,omitempty"`
}

func (c *Client) CreateStaticDNSEntry(ctx context.Context, entry StaticDNSEntry) (StaticDNSEntry, error) {
	ctx = c.addClientContext(ctx)
	ctx = tflog.SetField(ctx, "record_type", entry.RecordType)
	ctx = tflog.SetField(ctx, "key", entry.Key)
	ctx = tflog.SetField(ctx, "value", entry.Value)
	ctx = tflog.SetField(ctx, "ttl", entry.TTL)

	// POST /proxy/network/v2/api/site/:site/static-dns
	url, req := c.newAuthenticatedRequest(ctx, fmt.Sprintf("/proxy/network/v2/api/site/%s/static-dns", c.site))
	tflog.Debug(ctx, "creating static DNS entry", map[string]any{
		"url":   url,
		"entry": entry,
	})
	apiResponseSuccess := createStaticDNSEntriesResponseSuccess{}
	apiResponseError := createStaticDNSEntriesResponseError{}
	resp, err := req.
		SetBody(entry).
		SetResult(&apiResponseSuccess).
		SetError(&apiResponseError).
		Post(url)
	if err != nil {
		tflog.Error(ctx, "failed to execute POST request", map[string]any{
			"error_message": err.Error(),
		})
		return StaticDNSEntry{}, err
	}
	statusCode, _, _ := c.logResponse(ctx, resp)

	// if a non-200 error was returned by the API, something went wrong
	if statusCode != 200 {
		tflog.Error(ctx, "failed to create static DNS entry", map[string]any{
			"message":    apiResponseError.Message,
			"error_code": apiResponseError.ErrorCode,
			"code":       apiResponseError.Code,
			"details":    apiResponseError.Details,
		})
		return StaticDNSEntry{}, fmt.Errorf("failed to create static DNS entry: %s", apiResponseError.Message)
	}
	return StaticDNSEntry(apiResponseSuccess), nil
}

func (c *Client) DeleteStaticDNSEntry(ctx context.Context, id string) error {
	ctx = c.addClientContext(ctx)

	// DELETE /proxy/network/v2/api/site/:site/static-dns/:id
	url, req := c.newAuthenticatedRequest(ctx, fmt.Sprintf("/proxy/network/v2/api/site/%s/static-dns/%s", c.site, id))
	tflog.Debug(ctx, "deleting static DNS entry", map[string]any{
		"url": url,
	})
	apiResponseSuccess := deleteStaticDNSEntriesResponseSuccess{}
	apiResponseError := deleteStaticDNSEntriesResponseError{}
	resp, err := req.
		SetBody("{}").
		SetResult(&apiResponseSuccess).
		SetError(&apiResponseError).
		Delete(url)
	if err != nil {
		tflog.Error(ctx, "failed to execute DELETE request", map[string]any{
			"error_message": err.Error(),
		})
		return err
	}
	statusCode, _, _ := c.logResponse(ctx, resp)

	// if a non-200 error was returned by the API, something went wrong
	if statusCode != 200 {
		tflog.Error(ctx, "failed to delete static DNS entry", map[string]any{
			"message":    apiResponseError.Message,
			"error_code": apiResponseError.ErrorCode,
			"code":       apiResponseError.Code,
			"details":    apiResponseError.Details,
		})
		return fmt.Errorf("failed to delete static DNS entry: %s", apiResponseError.Message)
	}
	return nil
}

func (c *Client) GetStaticDNSEntries(ctx context.Context) ([]StaticDNSEntry, error) {
	ctx = c.addClientContext(ctx)

	// GET /proxy/network/v2/api/site/:site/static-dns
	url, req := c.newAuthenticatedRequest(ctx, fmt.Sprintf("/proxy/network/v2/api/site/%s/static-dns", c.site))
	tflog.Debug(ctx, "retrieving static DNS entries", map[string]any{
		"url": url,
	})
	apiResponseSuccess := getStaticDNSEntriesResponseSuccess{}
	apiResponseError := getStaticDNSEntriesResponseError{}
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
		tflog.Error(ctx, "failed to retrieve static DNS entries", map[string]any{
			"body": resp.Body(),
		})
		return nil, fmt.Errorf("failed to retrieve static DNS entries")
	}
	return []StaticDNSEntry(apiResponseSuccess), nil
}

func (c *Client) GetStaticDNSEntry(ctx context.Context, id string) (StaticDNSEntry, error) {
	ctx = tflog.SetField(ctx, "id", id)
	entries, err := c.GetStaticDNSEntries(ctx)
	if err != nil {
		return StaticDNSEntry{}, err
	}
	ctx = c.addClientContext(ctx)

	// find the ID in question
	tflog.Debug(ctx, "searching for static DNS entry")
	for _, entry := range entries {
		if entry.ID == id {
			tflog.Debug(ctx, "static DNS entry was located", map[string]any{"entry": entry})
			return entry, nil
		}
	}
	tflog.Warn(ctx, "static DNS entry not found")
	return StaticDNSEntry{}, fmt.Errorf("no static DNS entry found with an ID of '%s'", id)
}

func (c *Client) UpdateStaticDNSEntry(ctx context.Context, id string, entry StaticDNSEntry) (StaticDNSEntry, error) {
	ctx = c.addClientContext(ctx)
	ctx = tflog.SetField(ctx, "id", id)
	ctx = tflog.SetField(ctx, "record_type", entry.RecordType)
	ctx = tflog.SetField(ctx, "key", entry.Key)
	ctx = tflog.SetField(ctx, "value", entry.Value)
	ctx = tflog.SetField(ctx, "ttl", entry.TTL)

	// PUT /proxy/network/v2/api/site/:site/static-dns/:id
	url, req := c.newAuthenticatedRequest(ctx, fmt.Sprintf("/proxy/network/v2/api/site/%s/static-dns/%s", c.site, id))
	tflog.Debug(ctx, "updating static DNS entry", map[string]any{
		"url":   url,
		"entry": entry,
	})
	apiResponseSuccess := updateStaticDNSEntriesResponseSuccess{}
	apiResponseError := updateStaticDNSEntriesResponseError{}
	resp, err := req.
		SetBody(entry).
		SetResult(&apiResponseSuccess).
		SetError(&apiResponseError).
		Put(url)
	if err != nil {
		tflog.Error(ctx, "failed to execute PUT request", map[string]any{
			"error_message": err.Error(),
		})
		return StaticDNSEntry{}, err
	}
	statusCode, _, _ := c.logResponse(ctx, resp)

	// if a non-200 error was returned by the API, something went wrong
	if statusCode != 200 {
		tflog.Error(ctx, "failed to update static DNS entry", map[string]any{
			"message":    apiResponseError.Message,
			"error_code": apiResponseError.ErrorCode,
			"code":       apiResponseError.Code,
			"details":    apiResponseError.Details,
		})
		return StaticDNSEntry{}, fmt.Errorf("failed to update static DNS entry: %s", apiResponseError.Message)
	}
	return StaticDNSEntry(apiResponseSuccess), nil
}
