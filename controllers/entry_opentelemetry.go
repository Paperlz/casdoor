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
	"fmt"
	"io"
	"strings"

	"github.com/casdoor/casdoor/object"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// AddOtlpEntry
// @Title AddTrace
// @Tag OTLP API
// @Description receive otlp trace protobuf
// @Success 200 {object} string
// @router /api/v1/traces [post]
func (c *ApiController) AddTrace() {
	var req coltracepb.ExportTraceServiceRequest
	if !c.readOtlpRequest(&req) {
		return
	}

	entry, err := newOtlpEntry("otel.trace", "trace", summarizeTraceIngest(&req), &req)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte(fmt.Sprintf("marshal trace failed: %v", err)))
		return
	}
	entry.Summary = summarizeTraceIngest(&req)

	if _, err := object.AddEntry(entry); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte(fmt.Sprintf("save trace failed: %v", err)))
		return
	}

	resp := &coltracepb.ExportTraceServiceResponse{}
	respBytes, _ := proto.Marshal(resp)

	c.Ctx.Output.Header("Content-Type", "application/x-protobuf")
	c.Ctx.Output.SetStatus(200)
	c.Ctx.Output.Body(respBytes)
}

// AddMetrics
// @Title AddMetrics
// @Tag OTLP API
// @Description receive otlp metrics protobuf
// @Success 200 {object} string
// @router /api/v1/metrics [post]
func (c *ApiController) AddMetrics() {
	var req colmetricspb.ExportMetricsServiceRequest
	if !c.readOtlpRequest(&req) {
		return
	}

	entry, err := newOtlpEntry("otel.metric", "metric", summarizeMetricIngest(&req), &req)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte(fmt.Sprintf("marshal metrics failed: %v", err)))
		return
	}

	if _, err := object.AddEntry(entry); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte(fmt.Sprintf("save metrics failed: %v", err)))
		return
	}

	resp := &colmetricspb.ExportMetricsServiceResponse{}
	respBytes, _ := proto.Marshal(resp)

	c.Ctx.Output.Header("Content-Type", "application/x-protobuf")
	c.Ctx.Output.SetStatus(200)
	c.Ctx.Output.Body(respBytes)
}

// AddLogs
// @Title AddLogs
// @Tag OTLP API
// @Description receive otlp logs protobuf
// @Success 200 {object} string
// @router /api/v1/logs [post]
func (c *ApiController) AddLogs() {
	var req collogspb.ExportLogsServiceRequest
	if !c.readOtlpRequest(&req) {
		return
	}

	entry, err := newOtlpEntry("otel.log", "log", summarizeLogIngest(&req), &req)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte(fmt.Sprintf("marshal logs failed: %v", err)))
		return
	}

	if _, err := object.AddEntry(entry); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte(fmt.Sprintf("save logs failed: %v", err)))
		return
	}

	resp := &collogspb.ExportLogsServiceResponse{}
	respBytes, _ := proto.Marshal(resp)

	c.Ctx.Output.Header("Content-Type", "application/x-protobuf")
	c.Ctx.Output.SetStatus(200)
	c.Ctx.Output.Body(respBytes)
}

func (c *ApiController) readOtlpRequest(message proto.Message) bool {
	if !strings.HasPrefix(c.Ctx.Input.Header("Content-Type"), "application/x-protobuf") {
		c.Ctx.Output.SetStatus(415)
		c.Ctx.Output.Body([]byte("unsupported content type"))
		return false
	}

	body, err := io.ReadAll(c.Ctx.Request.Body)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte("read body failed"))
		return false
	}

	if err := proto.Unmarshal(body, message); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte(fmt.Sprintf("bad protobuf: %v", err)))
		return false
	}

	return true
}

func newOtlpEntry(source, category, summary string, message proto.Message) (*object.Entry, error) {
	payload, err := protojson.Marshal(message)
	if err != nil {
		return nil, err
	}

	return object.NewOtelEntry(source, category, summary, payload), nil
}

func summarizeTraceIngest(req *coltracepb.ExportTraceServiceRequest) string {
	resourceSpanCount := len(req.GetResourceSpans())
	scopeSpanCount := 0
	spanCount := 0

	for _, resourceSpans := range req.GetResourceSpans() {
		scopeSpans := resourceSpans.GetScopeSpans()
		scopeSpanCount += len(scopeSpans)
		for _, scope := range scopeSpans {
			spanCount += len(scope.GetSpans())
		}
	}

	return fmt.Sprintf("OTLP trace ingest: %d resource spans, %d scope spans, %d spans", resourceSpanCount, scopeSpanCount, spanCount)
}

func summarizeMetricIngest(req *colmetricspb.ExportMetricsServiceRequest) string {
	resourceMetricCount := len(req.GetResourceMetrics())
	scopeMetricCount := 0
	metricCount := 0

	for _, resourceMetrics := range req.GetResourceMetrics() {
		scopeMetrics := resourceMetrics.GetScopeMetrics()
		scopeMetricCount += len(scopeMetrics)
		for _, scope := range scopeMetrics {
			metricCount += len(scope.GetMetrics())
		}
	}

	return fmt.Sprintf("OTLP metric ingest: %d resource metrics, %d scope metrics, %d metrics", resourceMetricCount, scopeMetricCount, metricCount)
}

func summarizeLogIngest(req *collogspb.ExportLogsServiceRequest) string {
	resourceLogCount := len(req.GetResourceLogs())
	scopeLogCount := 0
	logRecordCount := 0

	for _, resourceLogs := range req.GetResourceLogs() {
		scopeLogs := resourceLogs.GetScopeLogs()
		scopeLogCount += len(scopeLogs)
		for _, scope := range scopeLogs {
			logRecordCount += len(scope.GetLogRecords())
		}
	}

	return fmt.Sprintf("OTLP log ingest: %d resource logs, %d scope logs, %d log records", resourceLogCount, scopeLogCount, logRecordCount)
}
