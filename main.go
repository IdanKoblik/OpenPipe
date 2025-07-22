package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

func InitMeterProvider() (*metric.MeterProvider, error) {
	 exporter, err := prometheus.New()
	 if err != nil {
	 	return nil, err
	 }
 
	 provider := metric.NewMeterProvider(metric.WithReader(exporter))
	 return provider, nil
}

func main() {
	 foundDocker := false
    for _, arg := range os.Args[1:] { 
		 if arg == "--docker" {
            foundDocker = true
            break
        }
    }

	 configFile := "config.yml"
	 if foundDocker {
		configFile = "/home/container/config.yml" 
	 }

	 cfg, err := ReadConfig(configFile)
	 if err != nil {
	 	 log.Fatalf("Cannot parse config file: %v", err)
	 }
 
	 http.Handle("/metrics", promhttp.Handler())
	 log.Printf("Serving metrics at http://%s:%d/metrics\n", cfg.Web.Host, cfg.Web.Port)
	 port := fmt.Sprintf(":%d", cfg.Web.Port)
 
	 go func() {
	 	 if err := http.ListenAndServe(port, nil); err != nil {
	 	 	 log.Fatalf("Failed to start HTTP server: %v", err)
	 	 }
	 }()
 
	 meterProvider, err := InitMeterProvider()
	 if err != nil {
	 	 log.Fatalf("Failed to initialize meter provider: %v", err)
	 }

	 defer func() {
	 	 if err := meterProvider.Shutdown(context.Background()); err != nil {
	 	 	 log.Fatalf("Failed to shut down meter provider: %v", err)
	 	 }
	 }()
	 otel.SetMeterProvider(meterProvider)

	 mm := NewMetricsManager()
	 msgs, err := ConsumeMessages(cfg)
	 if err != nil {
		 log.Fatalf("Failed to start consuming messages: %s", err)
	 }
 
	 log.Println("Waiting for messages...")
	 for msg := range msgs {
	 	 log.Printf("Received: %s", msg.Body)
	 	 msg.Ack(false)
		processMessage(context.Background(), mm, msg.Body)
	 }
}

func processMessage(ctx context.Context, mm *MetricsManager, msgBody []byte) {
	msg, err := ParseMessage(msgBody)
	if err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}
	if err := mm.RecordMetrics(ctx, msg); err != nil {
		log.Printf("Failed to record metrics: %v", err)
	}
}
