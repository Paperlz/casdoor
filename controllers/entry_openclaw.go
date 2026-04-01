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
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// ReceiveOpenClawEntry
// @Title ReceiveOpenClawEntry
// @Tag Entry API
// @Description receive OpenClaw logs and persist them as entries
// @router /openclaw-webhook/:owner/:name [post]
func (c *ApiController) ReceiveOpenClawEntry() {
	owner := c.Ctx.Input.Param(":owner")
	name := c.Ctx.Input.Param(":name")
	if owner == "" || name == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.ResponseError("invalid OpenClaw agent path")
		return
	}

	agent, err := object.GetAgent(util.GetId(owner, name))
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.ResponseError(err.Error())
		return
	}
	if agent == nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.ResponseError("OpenClaw agent not found")
		return
	}

	token := strings.TrimSpace(strings.TrimPrefix(c.Ctx.Input.Header("Authorization"), "Bearer "))
	if agent.Token == "" || token == "" || token != agent.Token {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.ResponseError("invalid OpenClaw webhook token")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(c.Ctx.ResponseWriter.ResponseWriter, c.Ctx.Request.Body, object.MaxOpenClawWebhookPayloadBytes))
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.Ctx.Output.SetStatus(http.StatusRequestEntityTooLarge)
			c.ResponseError("OpenClaw payload too large")
			return
		}
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.ResponseError(err.Error())
		return
	}

	if len(body) == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.ResponseError("empty OpenClaw payload")
		return
	}

	entry, err := object.NewEntryFromOpenClawPayload(agent, body)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.ResponseError(err.Error())
		return
	}

	existedEntry, err := object.GetEntry(entry.GetId())
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.ResponseError(err.Error())
		return
	}
	if existedEntry != nil {
		c.ResponseOk(map[string]string{
			"entryId": existedEntry.GetId(),
			"status":  "received",
		})
		return
	}

	_, err = object.AddEntry(entry)
	if err != nil {
		if object.IsDuplicateKeyError(err) {
			c.ResponseOk(map[string]string{
				"entryId": entry.GetId(),
				"status":  "received",
			})
			return
		}
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(map[string]string{
		"entryId": entry.GetId(),
		"status":  "received",
	})
}
