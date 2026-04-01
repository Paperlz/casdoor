// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/casdoor/casdoor/util"
)

const (
	MaxOpenClawWebhookPayloadBytes = 256 * 1024
	MaxStoredOpenClawPayloadBytes  = 64 * 1024
	maxEntryShortFieldBytes        = 100
	maxEntryMediumFieldBytes       = 200
	maxEntryLevelFieldBytes        = 50
)

func NewEntryFromOpenClawPayload(agent *Agent, body []byte) (*Entry, error) {
	payload := map[string]interface{}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	usage := getMapValue(payload, "usage")
	context := getMapValue(payload, "context")

	action := getStringValue(payload, "action")
	status := getStringValue(payload, "status")
	eventType := resolveOpenClawEventType(payload)
	if eventType == "" {
		eventType = "openclaw.event"
	}

	summary := getStringValue(payload, "summary")
	if summary == "" {
		summary = getStringValue(payload, "message")
	}

	level := getStringValue(payload, "level")
	if level == "" {
		if status == "error" || getStringValue(payload, "error") != "" {
			level = "error"
		} else {
			level = "info"
		}
	}

	ts := getInt64Value(payload, "ts")
	if ts == 0 {
		ts = getInt64Value(payload, "runAtMs")
	}

	storedPayload, payloadTruncated := truncateOpenClawPayload(body)
	entryName := buildOpenClawEntryName(agent, payload, body, eventType, action, status, ts)

	entry := &Entry{
		Owner:        agent.Owner,
		Name:         entryName,
		CreatedTime:  util.GetCurrentTime(),
		UpdatedTime:  util.GetCurrentTime(),
		DisplayName:  truncateStringForColumn(buildOpenClawEntryDisplayName(agent, eventType, action, status), maxEntryShortFieldBytes),
		Organization: truncateStringForColumn(agent.Owner, maxEntryShortFieldBytes),
		Application:  truncateStringForColumn(agent.Application, maxEntryShortFieldBytes),
		Agent:        truncateStringForColumn(agent.GetId(), maxEntryMediumFieldBytes),
		Source:       "openclaw",
		EventType:    truncateStringForColumn(eventType, maxEntryShortFieldBytes),
		Channel:      truncateStringForColumn(getStringValue(payload, "channel"), maxEntryShortFieldBytes),
		Action:       truncateStringForColumn(action, maxEntryShortFieldBytes),
		Status:       truncateStringForColumn(status, maxEntryShortFieldBytes),
		Level:        truncateStringForColumn(level, maxEntryLevelFieldBytes),
		JobId:        truncateStringForColumn(getStringValue(payload, "jobId"), maxEntryMediumFieldBytes),
		RunId:        truncateStringForColumn(getStringValue(payload, "runId"), maxEntryMediumFieldBytes),
		SessionId:    truncateStringForColumn(getStringValue(payload, "sessionId"), maxEntryMediumFieldBytes),
		SessionKey:   truncateStringForColumn(getStringValue(payload, "sessionKey"), maxEntryMediumFieldBytes),
		Model:        truncateStringForColumn(getStringValue(payload, "model"), maxEntryMediumFieldBytes),
		Provider:     truncateStringForColumn(getStringValue(payload, "provider"), maxEntryShortFieldBytes),
		InputTokens:  getInt64ValueWithFallbacks(usage, "input_tokens", "input"),
		OutputTokens: getInt64ValueWithFallbacks(usage, "output_tokens", "output"),
		TotalTokens:  getInt64ValueWithFallbacks(usage, "total_tokens", "total"),
		CacheReadTokens: getInt64ValueWithFallbacks(usage,
			"cache_read_tokens", "cacheRead"),
		CacheWriteTokens: getInt64ValueWithFallbacks(usage,
			"cache_write_tokens", "cacheWrite"),
		ContextLimit:     getInt64Value(context, "limit"),
		ContextUsed:      getInt64Value(context, "used"),
		CostUsd:          getFloat64Value(payload, "costUsd"),
		Ts:               ts,
		Seq:              getInt64Value(payload, "seq"),
		RunAtMs:          getInt64Value(payload, "runAtMs"),
		DurationMs:       getInt64Value(payload, "durationMs"),
		NextRunAtMs:      getInt64Value(payload, "nextRunAtMs"),
		DeliveryStatus:   truncateStringForColumn(getStringValue(payload, "deliveryStatus"), maxEntryShortFieldBytes),
		DeliveryError:    getStringValue(payload, "deliveryError"),
		Summary:          summary,
		Error:            getStringValue(payload, "error"),
		Payload:          storedPayload,
		PayloadBytes:     int64(len(body)),
		PayloadTruncated: payloadTruncated,
	}

	return entry, nil
}

func truncateStringForColumn(value string, maxBytes int) string {
	if maxBytes <= 0 || len(value) <= maxBytes {
		return value
	}

	return truncateUtf8StringByBytes(value, maxBytes)
}

func truncateUtf8StringByBytes(value string, maxBytes int) string {
	if maxBytes <= 0 || len(value) <= maxBytes {
		return value
	}

	lastSafeIndex := 0
	for index := range value {
		if index > maxBytes {
			break
		}
		lastSafeIndex = index
	}

	if lastSafeIndex == 0 && len(value) > maxBytes {
		return ""
	}

	if lastSafeIndex == 0 || lastSafeIndex > len(value) {
		return value
	}

	return value[:lastSafeIndex]
}

func buildOpenClawEntryName(
	agent *Agent,
	payload map[string]interface{},
	body []byte,
	eventType, action, status string,
	ts int64,
) string {
	runId := getStringValue(payload, "runId")
	seq := getInt64Value(payload, "seq")
	jobId := getStringValue(payload, "jobId")
	runAtMs := getInt64Value(payload, "runAtMs")

	identity := ""
	switch {
	case runId != "":
		identity = fmt.Sprintf("agent=%s|event=%s|run=%s|seq=%d", agent.GetId(), eventType, runId, seq)
	case jobId != "":
		identity = fmt.Sprintf(
			"agent=%s|event=%s|job=%s|action=%s|status=%s|runAtMs=%d|ts=%d",
			agent.GetId(), eventType, jobId, action, status, runAtMs, ts,
		)
	default:
		digest := sha256.Sum256(body)
		identity = fmt.Sprintf("agent=%s|event=%s|body=%x", agent.GetId(), eventType, digest)
	}

	sum := sha256.Sum256([]byte(identity))
	return fmt.Sprintf("openclaw_%x", sum[:16])
}

func truncateOpenClawPayload(body []byte) (string, bool) {
	if len(body) <= MaxStoredOpenClawPayloadBytes {
		return string(body), false
	}

	return truncateUtf8StringByBytes(string(body), MaxStoredOpenClawPayloadBytes), true
}

func resolveOpenClawEventType(payload map[string]interface{}) string {
	if eventType := getStringValue(payload, "type"); eventType != "" {
		return eventType
	}

	action := getStringValue(payload, "action")
	if action != "" {
		return fmt.Sprintf("cron.%s", action)
	}

	if getStringValue(payload, "jobId") != "" && getStringValue(payload, "message") != "" {
		return "cron.failure"
	}

	return ""
}

func buildOpenClawEntryDisplayName(agent *Agent, eventType, action, status string) string {
	if action != "" && status != "" {
		return fmt.Sprintf("%s %s (%s)", agent.Name, action, status)
	}
	if eventType != "" {
		return fmt.Sprintf("%s %s", agent.Name, eventType)
	}
	return fmt.Sprintf("%s openclaw event", agent.Name)
}

func getMapValue(payload map[string]interface{}, key string) map[string]interface{} {
	value, ok := payload[key]
	if !ok {
		return nil
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}

	return m
}

func getStringValue(payload map[string]interface{}, key string) string {
	if payload == nil {
		return ""
	}

	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", value)
	}
}

func getInt64Value(payload map[string]interface{}, key string) int64 {
	if payload == nil {
		return 0
	}

	value, ok := payload[key]
	if !ok || value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case json.Number:
		i, _ := v.Int64()
		return i
	default:
		return 0
	}
}

func getInt64ValueWithFallbacks(payload map[string]interface{}, keys ...string) int64 {
	for _, key := range keys {
		if value := getInt64Value(payload, key); value != 0 {
			return value
		}
	}

	return 0
}

func getFloat64Value(payload map[string]interface{}, key string) float64 {
	if payload == nil {
		return 0
	}

	value, ok := payload[key]
	if !ok || value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		f, _ := v.Float64()
		return f
	default:
		return 0
	}
}
