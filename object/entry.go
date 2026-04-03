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
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm"
)

type Entry struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	OccurredTime   string `xorm:"varchar(100) index" json:"occurredTime"`
	Agent          string `xorm:"varchar(100) index" json:"agent"`
	HostName       string `xorm:"varchar(100)" json:"hostName"`
	Source         string `xorm:"varchar(100) index" json:"source"`
	Category       string `xorm:"varchar(100) index" json:"category"`
	EventType      string `xorm:"varchar(100) index" json:"eventType"`
	Severity       string `xorm:"varchar(100) index" json:"severity"`
	Status         string `xorm:"varchar(100)" json:"status"`
	Summary        string `xorm:"varchar(500)" json:"summary"`
	RawPayload     string `xorm:"mediumtext" json:"rawPayload"`
	Labels         string `xorm:"text" json:"labels"`
	TraceId        string `xorm:"varchar(100) index" json:"traceId"`
	SessionId      string `xorm:"varchar(100) index" json:"sessionId"`
	Pid            int    `json:"pid"`
	CorrelationKey string `xorm:"varchar(100) index" json:"correlationKey"`

	Url         string `xorm:"varchar(500)" json:"url"`
	Token       string `xorm:"varchar(500)" json:"token"`
	Application string `xorm:"varchar(100)" json:"application"`
	Message     string `xorm:"mediumtext" json:"message"`
}

type EntryFilter struct {
	Agent     string
	Source    string
	Category  string
	EventType string
	Severity  string
	TraceId   string
	SessionId string
	TimeFrom  string
	TimeTo    string
	Field     string
	Value     string
}

func NewTraceEntry(message []byte) *Entry {
	return NewOtelEntry("otel.trace", "trace", "OTLP trace ingest", message)
}

func NewOtelEntry(source, category, summary string, message []byte) *Entry {
	currentTime := util.GetCurrentTime()
	entryId := fmt.Sprintf("entry_%s_%s", util.GenerateSimpleTimeId(), util.GetRandomName())

	return &Entry{
		Owner:        CasdoorOrganization,
		Name:         entryId,
		CreatedTime:  currentTime,
		UpdatedTime:  currentTime,
		OccurredTime: currentTime,
		DisplayName:  entryId,
		Source:       source,
		Category:     category,
		Summary:      summary,
		RawPayload:   string(message),
	}
}

func GetEntries(owner string) ([]*Entry, error) {
	entries := []*Entry{}
	session := getEntrySession(owner, EntryFilter{}, 0, 0, "", "")
	defer session.Close()
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
	if entry.CreatedTime == "" {
		entry.CreatedTime = util.GetCurrentTime()
	}
	if entry.UpdatedTime == "" {
		entry.UpdatedTime = entry.CreatedTime
	}
	if entry.OccurredTime == "" {
		entry.OccurredTime = entry.CreatedTime
	}

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

func GetEntryCount(owner string, filter EntryFilter) (int64, error) {
	session := getEntrySession(owner, filter, 0, 0, "", "")
	defer session.Close()
	return session.Count(&Entry{})
}

func GetPaginationEntries(owner string, offset, limit int, filter EntryFilter, sortField, sortOrder string) ([]*Entry, error) {
	entries := []*Entry{}
	session := getEntrySession(owner, filter, offset, limit, sortField, sortOrder)
	defer session.Close()
	err := session.Find(&entries)
	if err != nil {
		return entries, err
	}

	return entries, nil
}

func getEntrySession(owner string, filter EntryFilter, offset, limit int, sortField, sortOrder string) *xorm.Session {
	session := ormer.Engine.NewSession()
	if offset > 0 && limit > 0 {
		session = session.Limit(limit, offset)
	} else if offset == 0 && limit > 0 {
		session = session.Limit(limit, 0)
	}

	if owner != "" {
		session = session.And("owner=?", owner)
	}
	if filter.Agent != "" {
		session = session.And("agent=?", filter.Agent)
	}
	if filter.Source != "" {
		session = session.And("source=?", filter.Source)
	}
	if filter.Category != "" {
		session = session.And("category=?", filter.Category)
	}
	if filter.EventType != "" {
		session = session.And("event_type=?", filter.EventType)
	}
	if filter.Severity != "" {
		session = session.And("severity=?", filter.Severity)
	}
	if filter.TraceId != "" {
		session = session.And("trace_id=?", filter.TraceId)
	}
	if filter.SessionId != "" {
		session = session.And("session_id=?", filter.SessionId)
	}
	if filter.TimeFrom != "" {
		session = session.And("occurred_time>=?", filter.TimeFrom)
	}
	if filter.TimeTo != "" {
		session = session.And("occurred_time<=?", filter.TimeTo)
	}
	if filter.Field != "" && filter.Value != "" && util.FilterField(filter.Field) {
		session = session.And(fmt.Sprintf("%s like ?", util.SnakeString(filter.Field)), fmt.Sprintf("%%%s%%", filter.Value))
	}

	sortField = normalizeEntrySortField(sortField)
	if sortField == "" {
		if strings.EqualFold(sortOrder, "ascend") {
			session = session.Asc("occurred_time").Asc("created_time")
		} else {
			session = session.Desc("occurred_time").Desc("created_time")
		}
		return session
	}

	if strings.EqualFold(sortOrder, "ascend") {
		session = session.Asc(sortField)
	} else {
		session = session.Desc(sortField)
	}
	return session
}

func normalizeEntrySortField(sortField string) string {
	switch util.SnakeString(sortField) {
	case "created_time", "updated_time", "occurred_time", "agent", "source", "category", "event_type", "severity", "trace_id", "session_id", "name":
		return util.SnakeString(sortField)
	default:
		return ""
	}
}
