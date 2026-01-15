package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"nextui-game-manager/shared"

	converthtmltabletodata "github.com/activcoding/HTML-Table-to-JSON"
	"qlova.tech/sum"
)

type HostType = sum.Int[shared.HostType]

type HttpTableClient struct {
	RootURL            string
	HostType           HostType
	TableColumns       shared.TableColumns
	SourceReplacements map[string]string
	Filters            []string
}

func NewHttpTableClient(rootURL string, hostType HostType, tableColumns shared.TableColumns,
	sourceReplacements map[string]string, filters []string) *HttpTableClient {
	return &HttpTableClient{
		RootURL:            rootURL,
		HostType:           hostType,
		TableColumns:       tableColumns,
		SourceReplacements: sourceReplacements,
		Filters:            filters,
	}
}

func (c *HttpTableClient) Close() error {
	return nil
}

func (c *HttpTableClient) ListDirectory(subdirectory string) (shared.Items, error) {
	logger := GetLoggerInstance()
	params := url.Values{}

	switch c.HostType {
	case shared.HostTypes.APACHE:
		params.Add("F", "2") // To enable table mode for mod_autoindex
	}

	combinedUrl, err := url.JoinPath(c.RootURL, subdirectory)
	if err != nil {
		return nil, fmt.Errorf("unable to build download url: %w", err)
	}

	logger.Debug("Listing Directory", "combined_url", combinedUrl)

	u, err := url.Parse(combinedUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse table URL: %v", err)
	}
	u.RawQuery = params.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch table:, %v", err)
	}
	defer resp.Body.Close()

	jsonBytes, err := converthtmltabletodata.ConvertReaderToJSON(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse table into json: %v", err)
	}

	rawJson := string(jsonBytes)

	cleaned := rawJson

	switch c.HostType {
	case shared.HostTypes.APACHE:
		cleaned = strings.ReplaceAll(cleaned, "[[", "[")
		cleaned = strings.ReplaceAll(cleaned, "]]", "]")
		cleaned = strings.ReplaceAll(cleaned, "Name", "filename")
		cleaned = strings.ReplaceAll(cleaned, "Size", "file_size")
		cleaned = strings.ReplaceAll(cleaned, "Last modified", "date")
	case shared.HostTypes.MEGATHREAD:
		{
			cleaned = strings.ReplaceAll(cleaned, "  ↓", "")
			cleaned = strings.ReplaceAll(cleaned, "[[", "[")
			cleaned = strings.ReplaceAll(cleaned, "]]", "]")
			cleaned = strings.ReplaceAll(cleaned, "File Name", "filename")
			cleaned = strings.ReplaceAll(cleaned, "File Size", "file_size")
			cleaned = strings.ReplaceAll(cleaned, "Date", "date")
		}
	}

	var items shared.Items
	err = json.Unmarshal([]byte(cleaned), &items)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json: %v", err)
	}

	// Skip the header row(s)
	switch c.HostType {
	case shared.HostTypes.APACHE,
		shared.HostTypes.MEGATHREAD:
		{
			if len(items) > 1 {
				return items[1:], nil
			}
		}
	}

	return nil, nil
}

func (c *HttpTableClient) BuildDownloadHeaders() map[string]string {
	headers := make(map[string]string)
	return headers
}
