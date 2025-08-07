package main

import (
	"context"
	"encoding/json"
	"testing"
	"reflect"

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

	expected := map[string]interface{}{
		"point.fields.cpu":  42.5,
		"point.fields.mem":  float64(128), // JSON numbers are unmarshaled as float64
		"point.name":        "system",
		"point.time":        "2025-07-21T10:00:00Z",
		"extra":             "value",
	}

	if !reflect.DeepEqual(msg, expected) {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, msg)
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

	// Case 1: Metric with float values only
	data1 := map[string]interface{}{
		"cpu.usage": 42.5,
		"mem.usage": 128,
	}

	err := mm.RecordMetrics(ctx, data1)
	if err != nil {
		t.Errorf("RecordMetrics returned error for valid float metrics: %v", err)
	}

	// Case 2: Mix of float and non-float values
	data2 := map[string]interface{}{
		"cpu.usage": 75,
		"players":   []interface{}{"alice", "bob"},
		"status":    "active",
	}

	err = mm.RecordMetrics(ctx, data2)
	if err != nil {
		t.Errorf("RecordMetrics returned error for mixed-type input: %v", err)
	}

	// Case 3: Only non-float values (should not error, just skip them)
	data3 := map[string]interface{}{
		"players": []interface{}{"charlie", "dana"},
		"status":  "idle",
	}

	err = mm.RecordMetrics(ctx, data3)
	if err != nil {
		t.Errorf("RecordMetrics returned error for non-float input: %v", err)
	}
}

