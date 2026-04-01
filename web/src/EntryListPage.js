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
import {Link} from "react-router-dom";
import {Button, Table, Tag} from "antd";
import * as Setting from "./Setting";
import * as EntryBackend from "./backend/EntryBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";

class EntryListPage extends BaseListPage {
  deleteEntry(i) {
    EntryBackend.deleteEntry(this.state.data[i])
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.fetch({
            pagination: {
              ...this.state.pagination,
              current: this.state.pagination.current > 1 && this.state.data.length === 1 ? this.state.pagination.current - 1 : this.state.pagination.current,
            },
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  fetch = (params = {}) => {
    const field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    if (!params.pagination) {
      params.pagination = {current: 1, pageSize: 10};
    }

    this.setState({loading: true});
    EntryBackend.getEntries(Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      .then((res) => {
        this.setState({loading: false});
        if (res.status === "ok") {
          this.setState({
            data: res.data,
            pagination: {
              ...params.pagination,
              total: res.data2,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      });
  };

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

  renderTable(entries) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "170px",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record) => (
          <Link to={`/entries/${record.owner}/${text}`}>
            {text}
          </Link>
        ),
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "180px",
        sorter: true,
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: i18next.t("general:Application"),
        dataIndex: "application",
        key: "application",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("application"),
      },
      {
        title: "Agent",
        dataIndex: "agent",
        key: "agent",
        width: "180px",
        sorter: true,
        ...this.getColumnSearchProps("agent"),
        render: (text) => Setting.getShortText(text, 28),
      },
      {
        title: "Source",
        dataIndex: "source",
        key: "source",
        width: "110px",
        sorter: true,
        ...this.getColumnSearchProps("source"),
      },
      {
        title: "Event type",
        dataIndex: "eventType",
        key: "eventType",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("eventType"),
      },
      {
        title: i18next.t("general:Status"),
        dataIndex: "status",
        key: "status",
        width: "110px",
        sorter: true,
        ...this.getColumnSearchProps("status"),
        render: (text) => this.renderStatus(text),
      },
      {
        title: "Job ID",
        dataIndex: "jobId",
        key: "jobId",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("jobId"),
        render: (text) => Setting.getShortText(text, 24),
      },
      {
        title: "Session ID",
        dataIndex: "sessionId",
        key: "sessionId",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("sessionId"),
        render: (text) => Setting.getShortText(text, 24),
      },
      {
        title: "Model",
        dataIndex: "model",
        key: "model",
        width: "180px",
        sorter: true,
        ...this.getColumnSearchProps("model"),
        render: (text) => Setting.getShortText(text, 30),
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        width: "180px",
        sorter: true,
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: "Summary",
        dataIndex: "summary",
        key: "summary",
        sorter: true,
        ...this.getColumnSearchProps("summary"),
        render: (text) => Setting.getShortText(text, 60),
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "op",
        key: "op",
        width: "180px",
        fixed: (Setting.isMobile()) ? false : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/entries/${record.owner}/${record.name}`)}>{i18next.t("general:View")}</Button>
              <PopconfirmModal title={i18next.t("general:Sure to delete") + `: ${record.name} ?`} onConfirm={() => this.deleteEntry(index)}>
              </PopconfirmModal>
            </div>
          );
        },
      },
    ];

    const filteredColumns = Setting.filterTableColumns(columns, this.props.formItems ?? this.state.formItems);
    const paginationProps = {
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <Table
        scroll={{x: "max-content"}}
        dataSource={entries}
        columns={filteredColumns}
        rowKey={record => `${record.owner}/${record.name}`}
        pagination={{...this.state.pagination, ...paginationProps}}
        loading={this.state.loading}
        onChange={this.handleTableChange}
        size="middle"
        bordered
        title={() => (
          <div>
            {i18next.t("general:Entries")}
          </div>
        )}
      />
    );
  }
}

export default EntryListPage;
