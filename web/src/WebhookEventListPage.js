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

import React from "react";
import {Link} from "react-router-dom";
import {Button, Descriptions, Drawer, Result, Table, Tag, Tooltip} from "antd";
import i18next from "i18next";
import * as Setting from "./Setting";
import * as WebhookBackend from "./backend/WebhookBackend";
import Editor from "./common/Editor";

class WebhookEventListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      data: [],
      loading: false,
      replayingId: "",
      isAuthorized: true,
      detailShow: false,
      detailRecord: null,
      pagination: {
        current: 1,
        pageSize: 10,
        showQuickJumper: true,
        showSizeChanger: true,
      },
    };
  }

  componentDidMount() {
    window.addEventListener("storageOrganizationChanged", this.handleOrganizationChange);
    this.fetchWebhookEvents();
  }

  componentWillUnmount() {
    window.removeEventListener("storageOrganizationChanged", this.handleOrganizationChange);
  }

  handleOrganizationChange = () => {
    this.fetchWebhookEvents();
  };

  getStatusTag = (status) => {
    const statusConfig = {
      pending: {color: "gold", text: i18next.t("webhook:Pending")},
      success: {color: "green", text: i18next.t("webhook:Success")},
      failed: {color: "red", text: i18next.t("webhook:Failed")},
      retrying: {color: "blue", text: i18next.t("webhook:Retrying")},
    };

    const config = statusConfig[status] || {color: "default", text: status || i18next.t("webhook:Unknown")};

    return <Tag color={config.color}>{config.text}</Tag>;
  };

  getWebhookLink = (webhookName) => {
    if (!webhookName) {
      return "-";
    }

    const shortName = Setting.getShortName(webhookName);

    return (
      <Tooltip title={webhookName}>
        <Link to={`/webhooks/${encodeURIComponent(shortName)}`}>
          {shortName}
        </Link>
      </Tooltip>
    );
  };

  getOrganizationFilter = () => {
    if (!this.props.account) {
      return "";
    }

    return Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account);
  };

  fetchWebhookEvents = () => {
    this.setState({loading: true});

    WebhookBackend.getWebhookEvents("", this.getOrganizationFilter())
      .then((res) => {
        this.setState({loading: false});

        if (res.status === "ok") {
          this.setState((prevState) => ({
            data: res.data || [],
            pagination: {
              ...prevState.pagination,
              total: res.data?.length || 0,
            },
          }));
        } else if (Setting.isResponseDenied(res)) {
          this.setState({isAuthorized: false});
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error) => {
        this.setState({loading: false});
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  replayWebhookEvent = (event) => {
    const eventId = `${event.owner}/${event.name}`;
    this.setState({replayingId: eventId});

    WebhookBackend.replayWebhookEvent(eventId)
      .then((res) => {
        this.setState({replayingId: ""});

        if (res.status === "ok") {
          Setting.showMessage("success", typeof res.data === "string" ? res.data : "Webhook event replayed successfully");
          this.fetchWebhookEvents();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .catch((error) => {
        this.setState({replayingId: ""});
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  handleTableChange = (pagination) => {
    this.setState({pagination});
  };

  openDetailDrawer = (record) => {
    this.setState({
      detailRecord: record,
      detailShow: true,
    });
  };

  closeDetailDrawer = () => {
    this.setState({
      detailShow: false,
      detailRecord: null,
    });
  };

  getEditorMaxWidth = () => {
    return Setting.isMobile() ? window.innerWidth - 80 : 520;
  };

  jsonStrFormatter = (str) => {
    if (!str) {
      return "";
    }

    try {
      return JSON.stringify(JSON.parse(str), null, 2);
    } catch (e) {
      return str;
    }
  };

  getDetailField = (field) => {
    return this.state.detailRecord ? this.state.detailRecord[field] ?? "" : "";
  };

  renderTable = () => {
    const columns = [
      {
        title: "Webhook Name",
        dataIndex: "webhookName",
        key: "webhookName",
        width: 220,
        render: (text) => this.getWebhookLink(text),
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "organization",
        key: "organization",
        width: 160,
        render: (text) => text ? <Link to={`/organizations/${text}`}>{text}</Link> : "-",
      },
      {
        title: "Status",
        dataIndex: "status",
        key: "status",
        width: 140,
        filters: [
          {text: "Pending", value: "pending"},
          {text: "Success", value: "success"},
          {text: "Failed", value: "failed"},
          {text: "Retrying", value: "retrying"},
        ],
        onFilter: (value, record) => record.status === value,
        render: (text) => this.getStatusTag(text),
      },
      {
        title: "Attempt Count",
        dataIndex: "attemptCount",
        key: "attemptCount",
        width: 140,
        sorter: (a, b) => (a.attemptCount || 0) - (b.attemptCount || 0),
      },
      {
        title: "Next Retry Time",
        dataIndex: "nextRetryTime",
        key: "nextRetryTime",
        width: 180,
        sorter: (a, b) => {
          const timeA = a.nextRetryTime ? new Date(a.nextRetryTime).getTime() : 0;
          const timeB = b.nextRetryTime ? new Date(b.nextRetryTime).getTime() : 0;
          return timeA - timeB;
        },
        render: (text) => text ? Setting.getFormattedDate(text) : "-",
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "action",
        key: "action",
        width: 180,
        fixed: Setting.isMobile() ? false : "right",
        render: (_, record) => {
          const eventId = `${record.owner}/${record.name}`;
          return (
            <>
              <Button
                type="link"
                style={{paddingLeft: 0}}
                onClick={() => this.openDetailDrawer(record)}
              >
                View
              </Button>
              <Button
                type="primary"
                loading={this.state.replayingId === eventId}
                onClick={() => this.replayWebhookEvent(record)}
              >
                Replay
              </Button>
            </>
          );
        },
      },
    ];

    return (
      <Table
        rowKey={(record) => `${record.owner}/${record.name}`}
        columns={columns}
        dataSource={this.state.data}
        loading={this.state.loading}
        pagination={{
          ...this.state.pagination,
          showTotal: (total) => i18next.t("general:{total} in total").replace("{total}", total),
        }}
        scroll={{x: "max-content"}}
        size="middle"
        bordered
        title={() => "Webhook Event Logs"}
        onChange={this.handleTableChange}
      />
    );
  };

  render() {
    if (!this.state.isAuthorized) {
      return (
        <Result
          status="403"
          title="403 Unauthorized"
          subTitle={i18next.t("general:Sorry, you do not have permission to access this page or logged in status invalid.")}
          extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>}
        />
      );
    }

    return (
      <>
        {this.renderTable()}
        <Drawer
          title="Webhook Event Detail"
          width={Setting.isMobile() ? "100%" : 720}
          placement="right"
          destroyOnClose
          onClose={this.closeDetailDrawer}
          open={this.state.detailShow}
        >
          <Descriptions
            bordered
            size="small"
            column={1}
            layout={Setting.isMobile() ? "vertical" : "horizontal"}
            style={{padding: "12px", height: "100%", overflowY: "auto"}}
          >
            <Descriptions.Item label="Webhook Name">
              {this.getDetailField("webhookName") ? this.getWebhookLink(this.getDetailField("webhookName")) : "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Organization")}>
              {this.getDetailField("organization") ? (
                <Link to={`/organizations/${this.getDetailField("organization")}`}>
                  {this.getDetailField("organization")}
                </Link>
              ) : "-"}
            </Descriptions.Item>
            <Descriptions.Item label="Status">
              {this.getStatusTag(this.getDetailField("status"))}
            </Descriptions.Item>
            <Descriptions.Item label="Attempt Count">
              {this.getDetailField("attemptCount") || 0}
            </Descriptions.Item>
            <Descriptions.Item label="Next Retry Time">
              {this.getDetailField("nextRetryTime") ? Setting.getFormattedDate(this.getDetailField("nextRetryTime")) : "-"}
            </Descriptions.Item>
            <Descriptions.Item label="Payload">
              <Editor
                value={this.jsonStrFormatter(this.getDetailField("payload"))}
                lang="json"
                fillHeight
                fillWidth
                maxWidth={this.getEditorMaxWidth()}
                dark
                readOnly
              />
            </Descriptions.Item>
            <Descriptions.Item label="Last Error">
              <Editor
                value={this.getDetailField("lastError") || "-"}
                fillHeight
                fillWidth
                maxWidth={this.getEditorMaxWidth()}
                dark
                readOnly
              />
            </Descriptions.Item>
          </Descriptions>
        </Drawer>
      </>
    );
  }
}

export default WebhookEventListPage;
