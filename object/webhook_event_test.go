// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"testing"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/xorm/names"
)

func TestWebhookEvent(t *testing.T) {
	initWebhookEventTestOrmer(t)

	// Test creating a webhook event
	webhook := &Webhook{
		Owner:                 "admin",
		Name:                  "test-webhook",
		Organization:          "test-org",
		Url:                   "http://localhost:8080/webhook",
		Method:                "POST",
		ContentType:           "application/json",
		IsEnabled:             true,
		MaxRetries:            5,
		RetryInterval:         60,
		UseExponentialBackoff: true,
	}

	record := &Record{
		Organization: "test-org",
		User:         "test-user",
		Action:       "test-action",
		Object:       `{"test": "data"}`,
	}

	event, err := CreateWebhookEventFromRecord(webhook, record, nil)
	if err != nil {
		t.Fatalf("Failed to create webhook event: %v", err)
	}

	if event == nil {
		t.Fatal("Event is nil")
	}

	if event.Status != WebhookEventStatusPending {
		t.Errorf("Expected status %s, got %s", WebhookEventStatusPending, event.Status)
	}

	if event.WebhookName != webhook.GetId() {
		t.Errorf("Expected webhook name %s, got %s", webhook.GetId(), event.WebhookName)
	}

	if event.Organization != record.Organization {
		t.Errorf("Expected organization %s, got %s", record.Organization, event.Organization)
	}

	if event.EventType != record.Action {
		t.Errorf("Expected event type %s, got %s", record.Action, event.EventType)
	}

	if event.MaxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", event.MaxRetries)
	}

	storedEvent, err := GetWebhookEvent(event.GetId())
	if err != nil {
		t.Fatalf("Failed to load persisted webhook event: %v", err)
	}

	if storedEvent == nil {
		t.Fatal("Persisted event is nil")
	}

	if storedEvent.MaxRetries != webhook.MaxRetries {
		t.Errorf("Expected persisted max retries %d, got %d", webhook.MaxRetries, storedEvent.MaxRetries)
	}
}

func TestCalculateNextRetryTime(t *testing.T) {
	// Test fixed interval
	nextTime := calculateNextRetryTime(1, 60, false)
	if nextTime == "" {
		t.Error("Next retry time should not be empty")
	}

	// Test exponential backoff
	nextTime = calculateNextRetryTime(1, 60, true)
	if nextTime == "" {
		t.Error("Next retry time should not be empty")
	}

	nextTime = calculateNextRetryTime(2, 60, true)
	if nextTime == "" {
		t.Error("Next retry time should not be empty")
	}
}

func TestWebhookEventStatus(t *testing.T) {
	initWebhookEventTestOrmer(t)

	event := &WebhookEvent{
		Owner:        "admin",
		Name:         util.GenerateId(),
		Status:       WebhookEventStatusPending,
		AttemptCount: 0,
		MaxRetries:   4,
	}

	_, err := AddWebhookEvent(event)
	if err != nil {
		t.Fatalf("Failed to persist webhook event: %v", err)
	}

	event.AttemptCount = 1
	_, err = UpdateWebhookEventStatus(event, WebhookEventStatusSuccess, 200, "OK", nil)
	if err != nil {
		t.Fatalf("Failed to update webhook event status: %v", err)
	}

	storedEvent, err := GetWebhookEvent(event.GetId())
	if err != nil {
		t.Fatalf("Failed to reload webhook event: %v", err)
	}

	if storedEvent == nil {
		t.Fatal("Stored event is nil")
	}

	if storedEvent.Status != WebhookEventStatusSuccess {
		t.Errorf("Expected status %s, got %s", WebhookEventStatusSuccess, storedEvent.Status)
	}

	if storedEvent.LastStatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", storedEvent.LastStatusCode)
	}

	if storedEvent.LastResponse != "OK" {
		t.Errorf("Expected response 'OK', got %s", storedEvent.LastResponse)
	}

	if storedEvent.MaxRetries != event.MaxRetries {
		t.Errorf("Expected max retries %d, got %d", event.MaxRetries, storedEvent.MaxRetries)
	}
}

func initWebhookEventTestOrmer(t *testing.T) {
	t.Helper()

	previousOrmer := ormer
	dsn := fmt.Sprintf("file:webhook-event-test-%d?mode=memory&cache=shared", time.Now().UnixNano())

	testOrmer, err := NewAdapter("sqlite", dsn, "")
	if err != nil {
		t.Fatalf("Failed to initialize test ormer: %v", err)
	}

	testOrmer.Engine.SetTableMapper(names.NewPrefixMapper(names.SnakeMapper{}, ""))

	err = testOrmer.Engine.Sync2(new(WebhookEvent))
	if err != nil {
		_ = testOrmer.Engine.Close()
		t.Fatalf("Failed to sync webhook event table: %v", err)
	}

	ormer = testOrmer

	t.Cleanup(func() {
		ormer = previousOrmer
		_ = testOrmer.Engine.Close()
	})
}
