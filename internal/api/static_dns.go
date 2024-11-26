package api

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type createStaticDNSRecordsResponseSuccess StaticDNSRecord

type createStaticDNSRecordsResponseError struct {
	Code      string `json:"code"`
	Details   any    `json:"details"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

type deleteStaticDNSRecordsResponseSuccess struct{}

type deleteStaticDNSRecordsResponseError struct {
	Code      string `json:"code"`
	Details   any    `json:"details"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

type getStaticDNSRecordsResponseSuccess []StaticDNSRecord

type getStaticDNSRecordsResponseError struct {
	Code      string `json:"code"`
	Details   any    `json:"details"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

type updateStaticDNSRecordsResponseSuccess StaticDNSRecord

type updateStaticDNSRecordsResponseError struct {
	Code      string `json:"code"`
	Details   any    `json:"details"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

type StaticDNSRecord struct {
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

func (c *Client) CreateStaticDNSRecord(ctx context.Context, record StaticDNSRecord) (StaticDNSRecord, error) {
	ctx = c.addClientContext(ctx)
	ctx = tflog.SetField(ctx, "record_type", record.RecordType)
	ctx = tflog.SetField(ctx, "key", record.Key)
	ctx = tflog.SetField(ctx, "value", record.Value)
	ctx = tflog.SetField(ctx, "ttl", record.TTL)

	// POST /proxy/network/v2/api/site/:site/static-dns
	url, req := c.newAuthenticatedRequest(ctx, fmt.Sprintf("/proxy/network/v2/api/site/%s/static-dns", c.site))
	tflog.Debug(ctx, "creating static DNS record", map[string]any{
		"url":    url,
		"record": record,
	})
	apiResponseSuccess := createStaticDNSRecordsResponseSuccess{}
	apiResponseError := createStaticDNSRecordsResponseError{}
	resp, err := req.
		SetBody(record).
		SetResult(&apiResponseSuccess).
		SetError(&apiResponseError).
		Post(url)
	if err != nil {
		tflog.Error(ctx, "failed to execute POST request", map[string]any{
			"error_message": err.Error(),
		})
		return StaticDNSRecord{}, err
	}
	statusCode, _, _ := c.logResponse(ctx, resp)

	// if a non-200 error was returned by the API, something went wrong
	if statusCode != 200 {
		tflog.Error(ctx, "failed to create static DNS record", map[string]any{
			"message":    apiResponseError.Message,
			"error_code": apiResponseError.ErrorCode,
			"code":       apiResponseError.Code,
			"details":    apiResponseError.Details,
		})
		return StaticDNSRecord{}, fmt.Errorf("failed to create static DNS record: %s", apiResponseError.Message)
	}
	return StaticDNSRecord(apiResponseSuccess), nil
}

func (c *Client) DeleteStaticDNSRecord(ctx context.Context, id string) error {
	ctx = c.addClientContext(ctx)

	// DELETE /proxy/network/v2/api/site/:site/static-dns/:id
	url, req := c.newAuthenticatedRequest(ctx, fmt.Sprintf("/proxy/network/v2/api/site/%s/static-dns/%s", c.site, id))
	tflog.Debug(ctx, "deleting static DNS record", map[string]any{
		"url": url,
	})
	apiResponseSuccess := deleteStaticDNSRecordsResponseSuccess{}
	apiResponseError := deleteStaticDNSRecordsResponseError{}
	resp, err := req.
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
		tflog.Error(ctx, "failed to delete static DNS record", map[string]any{
			"message":    apiResponseError.Message,
			"error_code": apiResponseError.ErrorCode,
			"code":       apiResponseError.Code,
			"details":    apiResponseError.Details,
		})
		return fmt.Errorf("failed to delete static DNS record: %s", apiResponseError.Message)
	}
	return nil
}

func (c *Client) GetStaticDNSRecords(ctx context.Context) ([]StaticDNSRecord, error) {
	ctx = c.addClientContext(ctx)

	// GET /proxy/network/v2/api/site/:site/static-dns
	url, req := c.newAuthenticatedRequest(ctx, fmt.Sprintf("/proxy/network/v2/api/site/%s/static-dns", c.site))
	tflog.Debug(ctx, "retrieving static DNS records", map[string]any{
		"url": url,
	})
	apiResponseSuccess := getStaticDNSRecordsResponseSuccess{}
	apiResponseError := getStaticDNSRecordsResponseError{}
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
		tflog.Error(ctx, "failed to retrieve static DNS records", map[string]any{
			"body": resp.Body(),
		})
		return nil, fmt.Errorf("failed to retrieve static DNS records")
	}
	return []StaticDNSRecord(apiResponseSuccess), nil
}

func (c *Client) GetStaticDNSRecord(ctx context.Context, id string) (StaticDNSRecord, error) {
	ctx = tflog.SetField(ctx, "id", id)
	records, err := c.GetStaticDNSRecords(ctx)
	if err != nil {
		return StaticDNSRecord{}, err
	}
	ctx = c.addClientContext(ctx)

	// find the ID in question
	tflog.Debug(ctx, "searching for static DNS record")
	for _, record := range records {
		if record.ID == id {
			tflog.Debug(ctx, "static DNS record was located", map[string]any{"record": record})
			return record, nil
		}
	}
	tflog.Warn(ctx, "static DNS record not found")
	return StaticDNSRecord{}, fmt.Errorf("no static DNS record found with an ID of '%s'", id)
}

func (c *Client) UpdateStaticDNSRecord(ctx context.Context, id string, record StaticDNSRecord) (
	StaticDNSRecord, error) {

	ctx = c.addClientContext(ctx)
	ctx = tflog.SetField(ctx, "id", id)
	ctx = tflog.SetField(ctx, "record_type", record.RecordType)
	ctx = tflog.SetField(ctx, "key", record.Key)
	ctx = tflog.SetField(ctx, "value", record.Value)
	ctx = tflog.SetField(ctx, "ttl", record.TTL)

	// PUT /proxy/network/v2/api/site/:site/static-dns/:id
	url, req := c.newAuthenticatedRequest(ctx, fmt.Sprintf("/proxy/network/v2/api/site/%s/static-dns/%s", c.site, id))
	tflog.Debug(ctx, "updating static DNS record", map[string]any{
		"url":    url,
		"record": record,
	})
	apiResponseSuccess := updateStaticDNSRecordsResponseSuccess{}
	apiResponseError := updateStaticDNSRecordsResponseError{}
	resp, err := req.
		SetBody(record).
		SetResult(&apiResponseSuccess).
		SetError(&apiResponseError).
		Put(url)
	if err != nil {
		tflog.Error(ctx, "failed to execute PUT request", map[string]any{
			"error_message": err.Error(),
		})
		return StaticDNSRecord{}, err
	}
	statusCode, _, _ := c.logResponse(ctx, resp)

	// if a non-200 error was returned by the API, something went wrong
	if statusCode != 200 {
		tflog.Error(ctx, "failed to update static DNS record", map[string]any{
			"message":    apiResponseError.Message,
			"error_code": apiResponseError.ErrorCode,
			"code":       apiResponseError.Code,
			"details":    apiResponseError.Details,
		})
		return StaticDNSRecord{}, fmt.Errorf("failed to update static DNS record: %s", apiResponseError.Message)
	}
	return StaticDNSRecord(apiResponseSuccess), nil
}
