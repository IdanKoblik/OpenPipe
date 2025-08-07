package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/metric/noop"
)

func TestRecordMetrics_ValidInputs(t *testing.T) {
	ctx := context.Background()
	meter := noop.NewMeterProvider().Meter("test-meter")

	tests := []struct {
		name       string
		jsonInput  string
		wantMetric string
		wantCount  int
	}{
		{
			name: "Single object float value",
			jsonInput: `{
				"metricName": "cpu_temp",
				"value": 42.5,
				"core": "0"
			}`,
			wantMetric: "cpu_temp",
			wantCount:  1,
		},
		{
			name: "Array of multiple objects",
			jsonInput: `[
				{"metricName": "requests_total", "value": 123, "method": "GET"},
				{"metricName": "requests_total", "value": 234, "method": "POST"},
				{"metricName": "service_status", "value": ["healthy","running"], "service": "auth"}
			]`,
			wantMetric: "requests_total",
			wantCount:  2,
		},
		{
			name: "String value treated as label",
			jsonInput: `{
				"metricName": "service_status",
				"value": "running",
				"service": "auth"
			}`,
			wantMetric: "service_status",
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RecordMetrics(ctx, meter, []byte(tt.jsonInput))
			if err != nil {
				t.Fatalf("RecordMetrics returned error: %v", err)
			}
			metrics := store.GetMetrics()
			samples, ok := metrics[tt.wantMetric]
			if !ok {
				t.Fatalf("metric %q not found in store", tt.wantMetric)
			}
			if len(samples) != tt.wantCount {
				t.Fatalf("metric %q expected %d samples, got %d", tt.wantMetric, tt.wantCount, len(samples))
			}
		})
	}
}

func TestRecordMetrics_InvalidInputs(t *testing.T) {
	ctx := context.Background()
	meter := noop.NewMeterProvider().Meter("test-meter")

	tests := []struct {
		name      string
		jsonInput string
		wantErr   string
	}{
		{
			name:      "Empty JSON",
			jsonInput: ``,
			wantErr:   "json unmarshal",
		},
		{
			name:      "Invalid JSON syntax",
			jsonInput: `{invalid}`,
			wantErr:   "json unmarshal",
		},
		{
			name:      "Missing metricName",
			jsonInput: `{"value": 42}`,
			wantErr:   "metricName missing",
		},
		{
			name:      "Empty metricName",
			jsonInput: `{"metricName": "", "value": 1}`,
			wantErr:   "metricName must be non-empty",
		},
		{
			name:      "Missing value",
			jsonInput: `{"metricName": "mymetric"}`,
			wantErr:   "value missing",
		},
		{
			name:      "Unsupported value type",
			jsonInput: `{"metricName": "mymetric", "value": {"map": "not supported"}}`,
			wantErr:   "unsupported value type",
		},
		{
			name:      "Array with non-object element",
			jsonInput: `[{"metricName":"m1","value":1}, 42]`,
			wantErr:   "element 1 is not an object",
		},
		{
			name:      "JSON is not object or array",
			jsonInput: `"string"`,
			wantErr:   "json must be object or array of objects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RecordMetrics(ctx, meter, []byte(tt.jsonInput))
			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestStringifyLabelValue(t *testing.T) {
	cases := []struct {
		input interface{}
		want  string
	}{
		{"hello", "hello"},
		{42, "42"},
		{int64(64), "64"},
		{3.14, "3.14"},
		{json.Number("123.45"), "123.45"},
		{[]interface{}{"a", 1, 2.5}, "a,1,2.5"},
		{nil, "<nil>"},
	}

	for _, c := range cases {
		got := stringifyLabelValue(c.input)
		if got != c.want {
			t.Errorf("stringifyLabelValue(%v) = %q; want %q", c.input, got, c.want)
		}
	}
}

func TestLabelsToAttributes(t *testing.T) {
	labels := map[string]string{"z": "last", "a": "first", "m": "middle"}
	attrs := labelsToAttributes(labels)
	if len(attrs) != len(labels) {
		t.Fatalf("expected %d attributes, got %d", len(labels), len(attrs))
	}

	expectedOrder := []string{"a", "m", "z"}
	for i, attr := range attrs {
		if string(attr.Key) != expectedOrder[i] {
			t.Errorf("expected attr key %q at index %d, got %q", expectedOrder[i], i, attr.Key)
		}
	}
}

