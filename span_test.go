// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadog.com/).
// Copyright 2018 Datadog, Inc.

package datadog

import (
	"reflect"
	"testing"
	"time"

	"go.opencensus.io/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

var (
	testStartTime = time.Now()
	testEndTime   = testStartTime.Add(10 * time.Second)
)

// spanPairs holds a set of trace.SpanData and its corresponding conversion to a ddSpan.
var spanPairs = map[string]struct {
	oc *trace.SpanData
	dd *ddSpan
}{
	"root": {
		oc: &trace.SpanData{
			SpanContext: trace.SpanContext{
				TraceID:      trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
				SpanID:       trace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				TraceOptions: 1,
			},
			SpanKind:  trace.SpanKindClient,
			Name:      "/a/b",
			StartTime: testStartTime,
			EndTime:   testEndTime,
			Attributes: map[string]interface{}{
				"str":   "abc",
				"bool":  true,
				"int64": int64(1),
			},
			Status: trace.Status{
				Code:    0,
				Message: "status-msg",
			},
		},
		dd: &ddSpan{
			TraceID:  651345242494996240,
			SpanID:   72623859790382856,
			Type:     "client",
			Name:     "opencensus",
			Resource: "/a/b",
			Start:    testStartTime.UnixNano(),
			Duration: testEndTime.UnixNano() - testStartTime.UnixNano(),
			Metrics:  map[string]float64{"int64": 1},
			Service:  "my-service",
			Meta: map[string]string{
				"bool":               "true",
				"str":                "abc",
				keyStatus:            "OK",
				keyStatusCode:        "0",
				keyStatusDescription: "status-msg",
			},
		},
	},
	"child": {
		oc: &trace.SpanData{
			SpanContext: trace.SpanContext{
				TraceID:      trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
				SpanID:       trace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				TraceOptions: 1,
			},
			ParentSpanID: trace.SpanID([8]byte{8, 7, 6, 5, 4, 3, 2, 1}),
			SpanKind:     trace.SpanKindClient,
			Name:         "/a/b",
			StartTime:    testStartTime,
			EndTime:      testEndTime,
			Attributes:   map[string]interface{}{},
			Status:       trace.Status{},
		},
		dd: &ddSpan{
			TraceID:  651345242494996240,
			SpanID:   72623859790382856,
			ParentID: 578437695752307201,
			Type:     "client",
			Name:     "opencensus",
			Resource: "/a/b",
			Start:    testStartTime.UnixNano(),
			Duration: testEndTime.UnixNano() - testStartTime.UnixNano(),
			Metrics:  map[string]float64{},
			Service:  "my-service",
			Meta: map[string]string{
				keyStatus:     "OK",
				keyStatusCode: "0",
			},
		},
	},
	"server_error_4xx": {
		oc: &trace.SpanData{
			SpanContext: trace.SpanContext{
				TraceID:      trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
				SpanID:       trace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				TraceOptions: 1,
			},
			SpanKind:   trace.SpanKindServer,
			Name:       "/a/b",
			StartTime:  testStartTime,
			EndTime:    testEndTime,
			Attributes: map[string]interface{}{},
			Status: trace.Status{
				Code:    trace.StatusCodeCancelled,
				Message: "status-msg",
			},
		},
		dd: &ddSpan{
			TraceID:  651345242494996240,
			SpanID:   72623859790382856,
			Type:     "server",
			Name:     "opencensus",
			Resource: "/a/b",
			Start:    testStartTime.UnixNano(),
			Duration: testEndTime.UnixNano() - testStartTime.UnixNano(),
			Metrics:  map[string]float64{},
			Error:    0,
			Service:  "my-service",
			Meta: map[string]string{
				keyStatus:            "CANCELLED",
				keyStatusCode:        "1",
				keyStatusDescription: "status-msg",
			},
		},
	},
	"server_error_5xx": {
		oc: &trace.SpanData{
			SpanContext: trace.SpanContext{
				TraceID:      trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
				SpanID:       trace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				TraceOptions: 1,
			},
			SpanKind:   trace.SpanKindServer,
			Name:       "/a/b",
			StartTime:  testStartTime,
			EndTime:    testEndTime,
			Attributes: map[string]interface{}{},
			Status: trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: "status-msg",
			},
		},
		dd: &ddSpan{
			TraceID:  651345242494996240,
			SpanID:   72623859790382856,
			Type:     "server",
			Name:     "opencensus",
			Resource: "/a/b",
			Start:    testStartTime.UnixNano(),
			Duration: testEndTime.UnixNano() - testStartTime.UnixNano(),
			Metrics:  map[string]float64{},
			Error:    1,
			Service:  "my-service",
			Meta: map[string]string{
				ext.ErrorMsg:         "status-msg",
				ext.ErrorType:        "INTERNAL",
				keyStatus:            "INTERNAL",
				keyStatusCode:        "13",
				keyStatusDescription: "status-msg",
			},
		},
	},
	"client_error_4xx": {
		oc: &trace.SpanData{
			SpanContext: trace.SpanContext{
				TraceID:      trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
				SpanID:       trace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				TraceOptions: 1,
			},
			SpanKind:   trace.SpanKindClient,
			Name:       "/a/b",
			StartTime:  testStartTime,
			EndTime:    testEndTime,
			Attributes: map[string]interface{}{},
			Status: trace.Status{
				Code:    trace.StatusCodeCancelled,
				Message: "status-msg",
			},
		},
		dd: &ddSpan{
			TraceID:  651345242494996240,
			SpanID:   72623859790382856,
			Type:     "client",
			Name:     "opencensus",
			Resource: "/a/b",
			Start:    testStartTime.UnixNano(),
			Duration: testEndTime.UnixNano() - testStartTime.UnixNano(),
			Metrics:  map[string]float64{},
			Error:    1,
			Service:  "my-service",
			Meta: map[string]string{
				ext.ErrorMsg:         "status-msg",
				ext.ErrorType:        "CANCELLED",
				keyStatus:            "CANCELLED",
				keyStatusCode:        "1",
				keyStatusDescription: "status-msg",
			},
		},
	},
	"client_error_5xx": {
		oc: &trace.SpanData{
			SpanContext: trace.SpanContext{
				TraceID:      trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
				SpanID:       trace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				TraceOptions: 1,
			},
			SpanKind:   trace.SpanKindClient,
			Name:       "/a/b",
			StartTime:  testStartTime,
			EndTime:    testEndTime,
			Attributes: map[string]interface{}{},
			Status: trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: "status-msg",
			},
		},
		dd: &ddSpan{
			TraceID:  651345242494996240,
			SpanID:   72623859790382856,
			Type:     "client",
			Name:     "opencensus",
			Resource: "/a/b",
			Start:    testStartTime.UnixNano(),
			Duration: testEndTime.UnixNano() - testStartTime.UnixNano(),
			Metrics:  map[string]float64{},
			Error:    0,
			Service:  "my-service",
			Meta: map[string]string{
				keyStatus:            "INTERNAL",
				keyStatusCode:        "13",
				keyStatusDescription: "status-msg",
			},
		},
	},
	"tags": {
		oc: &trace.SpanData{
			SpanContext: trace.SpanContext{
				TraceID:      trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
				SpanID:       trace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				TraceOptions: 1,
			},
			SpanKind:  trace.SpanKindServer,
			Name:      "/a/b",
			StartTime: testStartTime,
			EndTime:   testEndTime,
			Attributes: map[string]interface{}{
				ext.Error:            true,
				ext.ServiceName:      "other-service",
				ext.ResourceName:     "other-resource",
				ext.SpanType:         "other-type",
				ext.SamplingPriority: int64(ext.PriorityUserReject),
			},
			Status: trace.Status{},
		},
		dd: &ddSpan{
			TraceID:  651345242494996240,
			SpanID:   72623859790382856,
			Type:     "other-type",
			Name:     "opencensus",
			Resource: "other-resource",
			Start:    testStartTime.UnixNano(),
			Duration: testEndTime.UnixNano() - testStartTime.UnixNano(),
			Metrics: map[string]float64{
				keySamplingPriority: ext.PriorityUserReject,
			},
			Service: "other-service",
			Error:   1,
			Meta: map[string]string{
				keyStatus:     "OK",
				keyStatusCode: "0",
			},
		},
	},
	"slash": {
		oc: &trace.SpanData{
			SpanContext: trace.SpanContext{
				TraceID:      trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
				SpanID:       trace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				TraceOptions: 1,
			},
			SpanKind:  trace.SpanKindClient,
			Name:      "/",
			StartTime: testStartTime,
			EndTime:   testEndTime,
		},
		dd: &ddSpan{
			TraceID:  651345242494996240,
			SpanID:   72623859790382856,
			Type:     "client",
			Name:     "opencensus",
			Resource: "/",
			Start:    testStartTime.UnixNano(),
			Duration: testEndTime.UnixNano() - testStartTime.UnixNano(),
			Service:  "my-service",
			Meta: map[string]string{
				keyStatus:     "OK",
				keyStatusCode: "0",
			},
			Metrics: map[string]float64{},
		},
	},
}

func TestConvertSpan(t *testing.T) {
	service := "my-service"
	e := newTraceExporter(Options{Service: service})
	defer e.stop()

	for name, tt := range spanPairs {
		t.Run(name, func(t *testing.T) {
			if got := e.convertSpan(tt.oc); !reflect.DeepEqual(got, tt.dd) {
				t.Fatalf("\nGot:\n%#v\n\nWant:\n%#v\n", got, tt.dd)
			}
		})
	}
}

func TestGlobalTags(t *testing.T) {
	e := newTraceExporter(Options{
		Service:    "my-service",
		GlobalTags: map[string]interface{}{"key1": "value1"},
	})
	defer e.stop()

	got := e.convertSpan(spanPairs["tags"].oc)
	if got.Meta["key1"] != "value1" {
		t.Fatal("global tag not set")
	}
}

func TestSetError(t *testing.T) {
	for i, tt := range [...]struct {
		val interface{} // error value
		err int32       // expected error field value
		msg string      // expected error message tag value
	}{
		{val: "error", err: 1, msg: "error"},
		{val: true, err: 1},
		{val: false},
		{val: int64(12), err: 1},
		{val: int64(-1)},
		{val: int64(0)},
		{val: nil},
		{val: float32(0), err: 1},
	} {
		span := &ddSpan{Meta: map[string]string{}}
		setError(span, tt.val)
		if span.Error != tt.err {
			t.Fatalf("%d: span.Error mismatch, wanted %d, got %d", i, tt.err, span.Error)
		}
		if tt.msg != "" {
			if got, ok := span.Meta[ext.ErrorMsg]; !ok || got != tt.msg {
				t.Fatalf("%d: span.Meta[ext.ErrorMsg] mismatch, wanted %q, got %q", i, tt.msg, got)
			}
		}
	}
}

func TestSetStringTag(t *testing.T) {
	span := &ddSpan{Meta: map[string]string{}}
	eq := equalFunc(t)

	setStringTag(span, ext.ServiceName, "service")
	eq(span.Service, "service")

	setStringTag(span, ext.ResourceName, "resource")
	eq(span.Resource, "resource")

	setStringTag(span, ext.SpanType, "type")
	eq(span.Type, "type")

	setStringTag(span, "key", "val")
	eq(span.Meta["key"], "val")
}

func TestSetTag(t *testing.T) {
	testSpan := func() *ddSpan {
		return &ddSpan{
			Meta:    map[string]string{},
			Metrics: map[string]float64{},
		}
	}

	t.Run("error", func(t *testing.T) {
		span := testSpan()
		setTag(span, ext.Error, true)
		equalFunc(t)(span.Error, int32(1))
	})

	t.Run("string", func(t *testing.T) {
		eq := equalFunc(t)
		span := testSpan()
		setTag(span, ext.ResourceName, "resource")
		eq(span.Resource, "resource")
		setTag(span, "key", "value")
		eq(span.Meta["key"], "value")
	})

	t.Run("bool", func(t *testing.T) {
		eq := equalFunc(t)
		span := testSpan()
		setTag(span, "key", true)
		eq(span.Meta["key"], "true")
		setTag(span, "key2", false)
		eq(span.Meta["key2"], "false")
		setTag(span, ext.AnalyticsEvent, true)
		eq(span.Metrics[ext.EventSampleRate], 1.)
		setTag(span, ext.AnalyticsEvent, false)
		eq(span.Metrics[ext.EventSampleRate], 0.)
	})

	t.Run("int64", func(t *testing.T) {
		eq := equalFunc(t)
		span := testSpan()
		setTag(span, "key", int64(12))
		eq(span.Metrics["key"], float64(12))
		setTag(span, ext.SamplingPriority, int64(1))
		eq(span.Metrics[keySamplingPriority], float64(1))
	})

	t.Run("float64", func(t *testing.T) {
		eq := equalFunc(t)
		span := testSpan()
		setTag(span, "key", float64(12))
		eq(span.Metrics["key"], float64(12))
		setTag(span, ext.SamplingPriority, float64(1))
		eq(span.Metrics[keySamplingPriority], float64(1))
		setTag(span, ext.EventSampleRate, float64(0.4))
		eq(span.Metrics[ext.EventSampleRate], float64(0.4))
	})

	t.Run("default", func(t *testing.T) {
		span := testSpan()
		setTag(span, "key", 1)
		equalFunc(t)(span.Meta["key"], "1")
	})
}

// equalFunc returns a function that tests the equality of two values. It fails
// if there is a type mismatch.
func equalFunc(t *testing.T) func(got, want interface{}) {
	return func(a, b interface{}) {
		t.Helper()
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("mismatch: got %v, wanted %v", a, b)
		}
	}
}
