package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/attribute"
)

type Visualization string

type ServerMetricMessage struct {
	 Name          string        `json:"name"`
  	 Realm         string        `json:"realm"`
    ServerUUID    uuid.UUID     `json:"serverUUID"`
    Type          string        `json:"type"` 
    Point         Point         `json:"point"`
}

type Point struct {
    Fields map[string]float64 `json:"fields"`
    Name   string             `json:"name"`
    Time   time.Time          `json:"time"`
}

type rawPoint struct {
	Fields map[string]float64 `json:"fields"`
	Name   string             `json:"name"`
	Time   float64            `json:"time"`
}

type rawMessage struct {
	Name          string       `json:"name"`
	Realm         string       `json:"realm"`
	ServerUUID    string       `json:"serverUUID"`
	Type          string       `json:"type"`
	Point         rawPoint     `json:"point"`
}

type MetricsManager struct {
	meter metric.Meter
	instruments sync.Map
}

func NewMetricsManager() *MetricsManager {
	meter := otel.Meter("server_metrics")
	return &MetricsManager{meter: meter}
}

func ParseMessage(data []byte) (*ServerMetricMessage, error) {
	var raw rawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(raw.ServerUUID)
	if err != nil {
		return nil, err
	}

	sec := int64(raw.Point.Time)
	nsec := int64((raw.Point.Time - float64(sec)) * 1e9)
	t := time.Unix(sec, nsec)

	msg := &ServerMetricMessage{
		Name: 			raw.Name,
		Realm: 			raw.Realm,
		ServerUUID:    id,
		Type:          raw.Type,
		Point: Point{
			Fields: raw.Point.Fields,
			Name:   raw.Point.Name,
			Time:   t,
		},
	}

	return msg, nil
}

func (m *MetricsManager) RecordMetrics(ctx context.Context, msg *ServerMetricMessage) error {
	metricName := msg.Point.Name

	instrumentIface, ok := m.instruments.Load(metricName)
	var gauge metric.Float64ObservableGauge
	var err error

	if !ok {
		gauge, err = m.meter.Float64ObservableGauge(
			metricName,
			metric.WithDescription(fmt.Sprintf("Metrics for point %s", metricName)),
		)
		if err != nil {
			return err
		}
		m.instruments.Store(metricName, gauge)
	} else {
		gauge = instrumentIface.(metric.Float64ObservableGauge)
	}

	values := make(map[string]float64)
	for k, v := range msg.Point.Fields {
		values[k] = v
	}

	baseAttrs := []attribute.KeyValue{
		attribute.String("server_uuid", msg.ServerUUID.String()),
		attribute.String("realm", msg.Realm),
		attribute.String("name", msg.Name),
	}

	_, err = m.meter.RegisterCallback(
		func(ctx context.Context, observer metric.Observer) error {
			for fieldKey, val := range values {
				attrs := append(baseAttrs, attribute.String("field", fieldKey))
				attrSet := attribute.NewSet(attrs...)
				observer.ObserveFloat64(gauge, val, metric.WithAttributeSet(attrSet))
			}
			return nil
		},
		gauge,
	)
	if err != nil {
		return err
	}

	return nil
}
