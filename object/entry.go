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
	"fmt"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

var entryListColumns = []string{
	"owner",
	"name",
	"created_time",
	"updated_time",
	"display_name",
	"organization",
	"application",
	"agent",
	"source",
	"event_type",
	"channel",
	"action",
	"status",
	"level",
	"job_id",
	"run_id",
	"session_id",
	"session_key",
	"model",
	"provider",
	"input_tokens",
	"output_tokens",
	"total_tokens",
	"cache_read_tokens",
	"cache_write_tokens",
	"context_limit",
	"context_used",
	"cost_usd",
	"ts",
	"seq",
	"run_at_ms",
	"duration_ms",
	"next_run_at_ms",
	"delivery_status",
	"summary",
	"payload_bytes",
	"payload_truncated",
}

type Entry struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Organization string `xorm:"varchar(100) index" json:"organization"`
	Application  string `xorm:"varchar(100) index" json:"application"`
	Agent        string `xorm:"varchar(200) index" json:"agent"`

	Source    string `xorm:"varchar(100) index" json:"source"`
	EventType string `xorm:"varchar(100) index" json:"eventType"`
	Channel   string `xorm:"varchar(100)" json:"channel"`
	Action    string `xorm:"varchar(100)" json:"action"`
	Status    string `xorm:"varchar(100) index" json:"status"`
	Level     string `xorm:"varchar(50)" json:"level"`

	JobId      string `xorm:"varchar(200) index" json:"jobId"`
	RunId      string `xorm:"varchar(200) index" json:"runId"`
	SessionId  string `xorm:"varchar(200) index" json:"sessionId"`
	SessionKey string `xorm:"varchar(200) index" json:"sessionKey"`

	Model    string `xorm:"varchar(200)" json:"model"`
	Provider string `xorm:"varchar(100)" json:"provider"`

	InputTokens      int64 `xorm:"bigint" json:"inputTokens"`
	OutputTokens     int64 `xorm:"bigint" json:"outputTokens"`
	TotalTokens      int64 `xorm:"bigint" json:"totalTokens"`
	CacheReadTokens  int64 `xorm:"bigint" json:"cacheReadTokens"`
	CacheWriteTokens int64 `xorm:"bigint" json:"cacheWriteTokens"`
	ContextLimit     int64 `xorm:"bigint" json:"contextLimit"`
	ContextUsed      int64 `xorm:"bigint" json:"contextUsed"`

	CostUsd     float64 `xorm:"double" json:"costUsd"`
	Ts          int64   `xorm:"bigint index" json:"ts"`
	Seq         int64   `xorm:"bigint index" json:"seq"`
	RunAtMs     int64   `xorm:"bigint index" json:"runAtMs"`
	DurationMs  int64   `xorm:"bigint" json:"durationMs"`
	NextRunAtMs int64   `xorm:"bigint" json:"nextRunAtMs"`

	DeliveryStatus   string `xorm:"varchar(100)" json:"deliveryStatus"`
	DeliveryError    string `xorm:"mediumtext" json:"deliveryError"`
	Summary          string `xorm:"mediumtext" json:"summary"`
	Error            string `xorm:"mediumtext" json:"error"`
	Payload          string `xorm:"mediumtext" json:"payload"`
	PayloadBytes     int64  `xorm:"bigint" json:"payloadBytes"`
	PayloadTruncated bool   `json:"payloadTruncated"`
}

func GetEntries(owner string) ([]*Entry, error) {
	entries := []*Entry{}
	session := ormer.Engine.Desc("created_time").Cols(entryListColumns...)
	if owner != "" {
		session = session.Where("owner = ?", owner)
	}
	err := session.Find(&entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func getEntry(owner string, name string) (*Entry, error) {
	entry := Entry{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&entry)
	if err != nil {
		return nil, err
	}

	if existed {
		return &entry, nil
	}
	return nil, nil
}

func GetEntry(id string) (*Entry, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getEntry(owner, name)
}

func UpdateEntry(id string, entry *Entry) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	if e, err := getEntry(owner, name); err != nil {
		return false, err
	} else if e == nil {
		return false, nil
	}

	entry.UpdatedTime = util.GetCurrentTime()

	_, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(entry)
	if err != nil {
		return false, err
	}

	return true, nil
}

func AddEntry(entry *Entry) (bool, error) {
	affected, err := ormer.Engine.Insert(entry)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteEntry(entry *Entry) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{entry.Owner, entry.Name}).Delete(&Entry{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (entry *Entry) GetId() string {
	return fmt.Sprintf("%s/%s", entry.Owner, entry.Name)
}

func GetEntryCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Entry{})
}

func GetPaginationEntries(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Entry, error) {
	entries := []*Entry{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	session = session.Cols(entryListColumns...)
	err := session.Find(&entries)
	if err != nil {
		return entries, err
	}

	return entries, nil
}
