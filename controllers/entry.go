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

package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type EntryPayload struct {
	OccurredTime   string `json:"occurredTime"`
	HostName       string `json:"hostName"`
	Source         string `json:"source"`
	Category       string `json:"category"`
	EventType      string `json:"eventType"`
	Severity       string `json:"severity"`
	Status         string `json:"status"`
	Summary        string `json:"summary"`
	RawPayload     string `json:"rawPayload"`
	Labels         string `json:"labels"`
	TraceId        string `json:"traceId"`
	SessionId      string `json:"sessionId"`
	Pid            int    `json:"pid"`
	CorrelationKey string `json:"correlationKey"`
}

type AddEntriesRequest struct {
	Owner     string          `json:"owner"`
	AgentName string          `json:"agentName"`
	Entries   []*EntryPayload `json:"entries"`
}

// GetEntries
// @Title GetEntries
// @Tag Entry API
// @Description get entries
// @Param   owner     query    string  true        "The owner of entries"
// @Success 200 {array} object.Entry The Response object
// @router /get-entries [get]
func (c *ApiController) GetEntries() {
	owner := c.Ctx.Input.Query("owner")
	if owner == "admin" {
		owner = ""
	}

	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")
	filter := object.EntryFilter{
		Agent:     c.Ctx.Input.Query("agent"),
		Source:    c.Ctx.Input.Query("source"),
		Category:  c.Ctx.Input.Query("category"),
		EventType: c.Ctx.Input.Query("eventType"),
		Severity:  c.Ctx.Input.Query("severity"),
		TraceId:   c.Ctx.Input.Query("traceId"),
		SessionId: c.Ctx.Input.Query("sessionId"),
		TimeFrom:  c.Ctx.Input.Query("from"),
		TimeTo:    c.Ctx.Input.Query("to"),
		Field:     field,
		Value:     value,
	}

	if limit == "" || page == "" {
		entries, err := object.GetPaginationEntries(owner, 0, 0, filter, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(entries)
		return
	}

	limitInt := util.ParseInt(limit)
	count, err := object.GetEntryCount(owner, filter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	paginator := pagination.SetPaginator(c.Ctx, limitInt, count)
	entries, err := object.GetPaginationEntries(owner, paginator.Offset(), limitInt, filter, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(entries, paginator.Nums())
}

// GetEntry
// @Title GetEntry
// @Tag Entry API
// @Description get entry
// @Param   id     query    string  true        "The id ( owner/name ) of the entry"
// @Success 200 {object} object.Entry The Response object
// @router /get-entry [get]
func (c *ApiController) GetEntry() {
	id := c.Ctx.Input.Query("id")

	entry, err := object.GetEntry(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(entry)
}

// UpdateEntry
// @Title UpdateEntry
// @Tag Entry API
// @Description update entry
// @Param   id     query    string  true        "The id ( owner/name ) of the entry"
// @Param   body    body   object.Entry  true        "The details of the entry"
// @Success 200 {object} controllers.Response The Response object
// @router /update-entry [post]
func (c *ApiController) UpdateEntry() {
	id := c.Ctx.Input.Query("id")

	var entry object.Entry
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &entry)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateEntry(id, &entry))
	c.ServeJSON()
}

// AddEntry
// @Title AddEntry
// @Tag Entry API
// @Description add entry
// @Param   body    body   object.Entry  true        "The details of the entry"
// @Success 200 {object} controllers.Response The Response object
// @router /add-entry [post]
func (c *ApiController) AddEntry() {
	var entry object.Entry
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &entry)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddEntry(&entry))
	c.ServeJSON()
}

// AddEntries
// @Title AddEntries
// @Tag Entry API
// @Description add log or telemetry entries from probe
// @Param   body    body   controllers.AddEntriesRequest  true  "The entries payload"
// @Success 200 {object} controllers.Response The Response object
// @router /add-entries [post]
func (c *ApiController) AddEntries() {
	var req AddEntriesRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	agent, ok := c.requireProbeAgent(req.Owner, req.AgentName)
	if !ok {
		return
	}

	entries := make([]*object.Entry, 0, len(req.Entries))
	lastEventTime := ""
	for _, payload := range req.Entries {
		if payload == nil {
			continue
		}

		occurredTime := payload.OccurredTime
		if occurredTime == "" {
			occurredTime = util.GetCurrentTime()
		}
		if occurredTime > lastEventTime {
			lastEventTime = occurredTime
		}

		hostName := payload.HostName
		if hostName == "" {
			hostName = agent.HostName
		}

		entryName := fmt.Sprintf("entry_%s_%s", util.GenerateSimpleTimeId(), util.GetRandomName())
		entries = append(entries, &object.Entry{
			Owner:          agent.Owner,
			Name:           entryName,
			DisplayName:    entryName,
			OccurredTime:   occurredTime,
			Agent:          agent.Name,
			HostName:       hostName,
			Source:         payload.Source,
			Category:       payload.Category,
			EventType:      payload.EventType,
			Severity:       payload.Severity,
			Status:         payload.Status,
			Summary:        payload.Summary,
			RawPayload:     payload.RawPayload,
			Labels:         payload.Labels,
			TraceId:        payload.TraceId,
			SessionId:      payload.SessionId,
			Pid:            payload.Pid,
			CorrelationKey: payload.CorrelationKey,
		})
	}

	if len(entries) == 0 {
		c.ResponseError("entries is empty")
		return
	}

	affected, err := object.AddEntries(entries)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if lastEventTime != "" {
		agent.LastEventTime = lastEventTime
		_, _ = object.UpdateAgent(agent.GetId(), agent)
	}

	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

// DeleteEntry
// @Title DeleteEntry
// @Tag Entry API
// @Description delete entry
// @Param   body    body   object.Entry  true        "The details of the entry"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-entry [post]
func (c *ApiController) DeleteEntry() {
	var entry object.Entry
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &entry)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteEntry(&entry))
	c.ServeJSON()
}
