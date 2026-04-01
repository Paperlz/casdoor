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

import React from "react";
import {Button, Card, Col, Input, Row, Tag} from "antd";
import * as EntryBackend from "./backend/EntryBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

const {TextArea} = Input;

class EntryEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      entryName: props.match.params.entryName,
      owner: props.match.params.organizationName,
      entry: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.getEntry();
  }

  getEntry() {
    EntryBackend.getEntry(this.state.owner, this.state.entryName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        if (res.status === "ok") {
          this.setState({
            entry: res.data,
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      });
  }

  deleteEntry() {
    EntryBackend.deleteEntry(this.state.entry)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.props.history.push("/entries");
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderStatus(status) {
    if (!status) {
      return null;
    }

    let color = "default";
    if (status === "ok" || status === "success" || status === "completed") {
      color = "success";
    } else if (status === "error" || status === "failed") {
      color = "error";
    } else if (status === "pending" || status === "retrying") {
      color = "processing";
    }

    return <Tag color={color}>{status}</Tag>;
  }

  formatJson(payload) {
    if (!payload) {
      return "";
    }

    try {
      return JSON.stringify(JSON.parse(payload), null, 2);
    } catch {
      return payload;
    }
  }

  renderReadonlyField(label, value, options = {}) {
    const {span = 22, textArea = false, rows = 4} = options;

    return (
      <Row style={{marginTop: "20px"}}>
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {label}:
        </Col>
        <Col span={span}>
          {textArea ? (
            <TextArea value={value || ""} autoSize={{minRows: rows, maxRows: rows + 8}} readOnly />
          ) : (
            <Input value={value || ""} readOnly />
          )}
        </Col>
      </Row>
    );
  }

  renderEntry() {
    const {entry} = this.state;

    return (
      <Card size="small" title={
        <div>
          {i18next.t("general:Entries")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.getEntry()}>{i18next.t("general:Refresh")}</Button>
          <Button style={{marginLeft: "20px"}} danger onClick={() => this.deleteEntry()}>{i18next.t("general:Delete")}</Button>
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        {this.renderReadonlyField(i18next.t("general:Name"), entry.name)}
        {this.renderReadonlyField(i18next.t("general:Display name"), entry.displayName)}
        {this.renderReadonlyField(i18next.t("general:Organization"), entry.organization || entry.owner)}
        {this.renderReadonlyField(i18next.t("general:Application"), entry.application)}
        {this.renderReadonlyField("Agent", entry.agent)}
        {this.renderReadonlyField("Source", entry.source)}
        {this.renderReadonlyField("Event type", entry.eventType)}
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Status")}:
          </Col>
          <Col span={22}>
            {this.renderStatus(entry.status)}
          </Col>
        </Row>
        {this.renderReadonlyField("Action", entry.action)}
        {this.renderReadonlyField("Level", entry.level)}
        {this.renderReadonlyField("Job ID", entry.jobId)}
        {this.renderReadonlyField("Run ID", entry.runId)}
        {this.renderReadonlyField("Session ID", entry.sessionId)}
        {this.renderReadonlyField("Session key", entry.sessionKey)}
        {this.renderReadonlyField("Model", entry.model)}
        {this.renderReadonlyField("Provider", entry.provider)}
        {this.renderReadonlyField("Created time", Setting.getFormattedDate(entry.createdTime))}
        {this.renderReadonlyField("Run at", entry.runAtMs ? String(entry.runAtMs) : "")}
        {this.renderReadonlyField("Duration (ms)", entry.durationMs ? String(entry.durationMs) : "")}
        {this.renderReadonlyField("Next run at", entry.nextRunAtMs ? String(entry.nextRunAtMs) : "")}
        {this.renderReadonlyField("Input tokens", entry.inputTokens ? String(entry.inputTokens) : "")}
        {this.renderReadonlyField("Output tokens", entry.outputTokens ? String(entry.outputTokens) : "")}
        {this.renderReadonlyField("Total tokens", entry.totalTokens ? String(entry.totalTokens) : "")}
        {this.renderReadonlyField("Cache read tokens", entry.cacheReadTokens ? String(entry.cacheReadTokens) : "")}
        {this.renderReadonlyField("Cache write tokens", entry.cacheWriteTokens ? String(entry.cacheWriteTokens) : "")}
        {this.renderReadonlyField("Context limit", entry.contextLimit ? String(entry.contextLimit) : "")}
        {this.renderReadonlyField("Context used", entry.contextUsed ? String(entry.contextUsed) : "")}
        {this.renderReadonlyField("Cost (USD)", entry.costUsd ? String(entry.costUsd) : "")}
        {this.renderReadonlyField("Delivery status", entry.deliveryStatus)}
        {this.renderReadonlyField("Summary", entry.summary, {textArea: true, rows: 3})}
        {this.renderReadonlyField("Error", entry.error, {textArea: true, rows: 3})}
        {this.renderReadonlyField("Delivery error", entry.deliveryError, {textArea: true, rows: 3})}
        {this.renderReadonlyField("Payload", this.formatJson(entry.payload), {textArea: true, rows: 10})}
      </Card>
    );
  }

  render() {
    if (this.state.entry === null) {
      return null;
    }

    return (
      <div>
        {this.renderEntry()}
      </div>
    );
  }
}

export default EntryEditPage;
