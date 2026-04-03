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
	"crypto/subtle"
	"encoding/json"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type OpenClawProbeHeartbeatRequest struct {
	Owner            string   `json:"owner"`
	AgentName        string   `json:"agentName"`
	HostName         string   `json:"hostName"`
	DeployType       string   `json:"deployType"`
	Version          string   `json:"version"`
	GatewayPort      int      `json:"gatewayPort"`
	StateDir         string   `json:"stateDir"`
	WorkspacePath    string   `json:"workspacePath"`
	CollectorVersion string   `json:"collectorVersion"`
	Sources          []string `json:"sources"`
	Status           string   `json:"status"`
}

func (c *ApiController) getProbeBearerToken() string {
	authHeader := c.Ctx.Input.Header("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	return strings.TrimSpace(c.Ctx.Input.Header("X-Access-Token"))
}

func (c *ApiController) requireProbeAgent(owner, name string) (*object.Agent, bool) {
	if owner == "" || name == "" {
		c.Ctx.Output.SetStatus(400)
		c.ResponseError("owner or agentName is empty")
		return nil, false
	}

	agent, err := object.GetAgent(util.GetId(owner, name))
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if agent == nil {
		c.Ctx.Output.SetStatus(404)
		c.ResponseError("agent not found")
		return nil, false
	}

	token := c.getProbeBearerToken()
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(agent.Token)) != 1 {
		c.Ctx.Output.SetStatus(401)
		c.ResponseError("invalid probe token")
		return nil, false
	}

	return agent, true
}

// UpdateOpenClawProbeHeartbeat
// @Title UpdateOpenClawProbeHeartbeat
// @Tag OpenClaw Probe API
// @Description update openclaw agent heartbeat from collector
// @Param   body    body   controllers.OpenClawProbeHeartbeatRequest  true  "The heartbeat payload"
// @Success 200 {object} controllers.Response The Response object
// @router /update-openclaw-probe-heartbeat [post]
func (c *ApiController) UpdateOpenClawProbeHeartbeat() {
	var req OpenClawProbeHeartbeatRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	agent, ok := c.requireProbeAgent(req.Owner, req.AgentName)
	if !ok {
		return
	}

	agent.HostName = req.HostName
	agent.DeployType = req.DeployType
	agent.Version = req.Version
	agent.GatewayPort = req.GatewayPort
	agent.StateDir = req.StateDir
	agent.WorkspacePath = req.WorkspacePath
	agent.CollectorVersion = req.CollectorVersion
	agent.Sources = req.Sources
	agent.Status = req.Status
	agent.LastHeartbeat = util.GetCurrentTime()

	c.Data["json"] = wrapActionResponse(object.UpdateAgent(agent.GetId(), agent))
	c.ServeJSON()
}
