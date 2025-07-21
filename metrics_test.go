package main

import (
	"context"
	"encoding/json"
	"testing"

	"go.opentelemetry.io/otel"
)

func setupMeterProvider(t *testing.T) func() {
	mp, err := InitMeterProvider()
	if err != nil {
		t.Fatalf("failed to init meter provider: %v", err)
	}
	otel.SetMeterProvider(mp)
	return func() {
		_ = mp.Shutdown(context.Background())
	}
}

func TestParseMessage(t *testing.T) {
	validJSON := []byte(`{
		"point": {
			"fields": {"cpu": 42.5, "mem": 128},
			"name": "system",
			"time": "2025-07-21T10:00:00Z"
		},
		"extra": "value"
	}`)

	msg, err := ParseMessage(validJSON)
	if err != nil {
		t.Fatalf("ParseMessage returned error: %v", err)
	}

	if msg["cpu"] != 42.5 {
		t.Errorf("Expected cpu 42.5, got %v", msg["cpu"])
	}

	if msg["point_name"] != "system" {
		t.Errorf("Expected point_name 'system', got %v", msg["point_name"])
	}

	if _, ok := msg["point_time"]; !ok {
		t.Errorf("Expected point_time present")
	}

	if msg["extra"] != "value" {
		t.Errorf("Expected extra value, got %v", msg["extra"])
	}
}

func TestParseMessageInvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{invalid json}`)

	_, err := ParseMessage(invalidJSON)
	if err == nil {
		t.Errorf("Expected error for invalid JSON")
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input interface{}
		want  float64
		ok    bool
	}{
		{float64(1.23), 1.23, true},
		{int(5), 5.0, true},
		{int64(7), 7.0, true},
		{json.Number("8.9"), 8.9, true},
		{"10.11", 10.11, true},
		{"notafloat", 0, false},
		{true, 0, false},
	}

	for _, test := range tests {
		got, ok := toFloat64(test.input)
		if ok != test.ok || (ok && got != test.want) {
			t.Errorf("toFloat64(%v) = %v, %v; want %v, %v", test.input, got, ok, test.want, test.ok)
		}
	}
}

func TestRecordMetrics(t *testing.T) {
	teardown := setupMeterProvider(t)
	defer teardown()

	mm := NewMetricsManager()
	ctx := context.Background()

	err := mm.RecordMetrics(ctx, map[string]interface{}{"foo": 1})
	if err == nil {
		t.Error("Expected error when point_name is missing")
	}

	err = mm.RecordMetrics(ctx, map[string]interface{}{"point_name": 123})
	if err == nil {
		t.Error("Expected error when point_name is not string")
	}

	data := map[string]interface{}{
		"point_name": "test_metric",
		"field1":     42,
		"tag1":       "tagvalue",
	}

	err = mm.RecordMetrics(ctx, data)
	if err != nil {
		t.Errorf("RecordMetrics returned error: %v", err)
	}

	err = mm.RecordMetrics(ctx, data)
	if err != nil {
		t.Errorf("RecordMetrics returned error on reuse: %v", err)
	}
}

