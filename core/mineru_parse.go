package core

import (
	"context"
	"encoding/json"
	"strings"
)

func parseMinerUInInputOrder(ctx context.Context, client MinerUClient, files []InputFile) (MinerUParseResult, error) {
	var combined MinerUParseResult
	var rawRequests []any
	var rawResponses []any
	pageOffset := 0
	for _, file := range files {
		result, err := client.Parse(ctx, file)
		rawRequests = appendRawPayload(rawRequests, result.RawRequest)
		rawResponses = appendRawPayload(rawResponses, result.RawResponse)
		if err != nil {
			combined.RawRequest = marshalRawPayloads(rawRequests)
			combined.RawResponse = marshalRawPayloads(rawResponses)
			return combined, err
		}
		maxPage := -1
		for _, item := range result.ContentList {
			if item.PageIndex != nil {
				adjusted := *item.PageIndex + pageOffset
				item.PageIndex = &adjusted
				if adjusted > maxPage {
					maxPage = adjusted
				}
			}
			combined.ContentList = append(combined.ContentList, item)
		}
		if maxPage >= pageOffset {
			pageOffset = maxPage + 1
		} else {
			pageOffset++
		}
	}
	combined.RawRequest = marshalRawPayloads(rawRequests)
	combined.RawResponse = marshalRawPayloads(rawResponses)
	return combined, nil
}

func appendRawPayload(values []any, raw string) []any {
	if strings.TrimSpace(raw) == "" {
		return values
	}
	if json.Valid([]byte(raw)) {
		return append(values, json.RawMessage(raw))
	}
	return append(values, raw)
}

func marshalRawPayloads(values []any) string {
	if len(values) == 0 {
		return ""
	}
	data, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}
