package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var (
	collectorURL = "localhost:4317"
)

func main() {
	fmt.Println("starting otlp metric exporter...")
	initMeter()
}

func initMeter() {
	ctx := context.Background()

	ectx, ecancel := context.WithTimeout(ctx, 2*time.Second)
	defer ecancel()
	otlpExp, err := newOtlpMetricExporter(ectx)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			octx, ocancel := context.WithTimeout(ctx, 2*time.Second)

			rm := getResourceMetrics()
			err = otlpExp.Export(octx, rm)
			if err != nil {
				panic(err)
			}
			ocancel()
			fmt.Println("metric sent.")
			time.Sleep(5 * time.Second)
		}
	}()

	ictx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	<-ictx.Done()
}

func newOtlpMetricExporter(ctx context.Context) (metric.Exporter, error) {
	return otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(collectorURL),
		otlpmetricgrpc.WithAggregationSelector(metric.DefaultAggregationSelector),
		otlpmetricgrpc.WithTemporalitySelector(DeltaTemporalitySelector),
	)
}

func DeltaTemporalitySelector(ik metric.InstrumentKind) metricdata.Temporality {
	return metricdata.DeltaTemporality
}

func getResourceMetrics() *metricdata.ResourceMetrics {
	end := time.Now().UTC()
	start := end.Add(-60 * time.Second)

	alice := attribute.NewSet(attribute.String("user", "alice"))
	bob := attribute.NewSet(attribute.String("user", "bob"))

	otelDPtsFloat64 := []metricdata.DataPoint[float64]{
		{Attributes: alice, StartTime: start, Time: end, Value: 1.0},
		{Attributes: bob, StartTime: start, Time: end, Value: 2.0},
	}
	otelDPtsInt64 := []metricdata.DataPoint[int64]{
		{Attributes: alice, StartTime: start, Time: end, Value: 1},
		{Attributes: bob, StartTime: start, Time: end, Value: 2},
	}
	otelSumFloat64 := metricdata.Sum[float64]{
		Temporality: metricdata.DeltaTemporality,
		IsMonotonic: false,
		DataPoints:  otelDPtsFloat64,
	}
	otelSumInt64 := metricdata.Sum[int64]{
		Temporality: metricdata.CumulativeTemporality,
		IsMonotonic: true,
		DataPoints:  otelDPtsInt64,
	}
	otelGaugeFloat64 := metricdata.Gauge[float64]{DataPoints: otelDPtsFloat64}
	otelMetrics := []metricdata.Metrics{
		{
			Name:        "float64-sum",
			Description: "Sum with float64 values",
			Unit:        "1",
			Data:        otelSumFloat64, //not working
		},
		{
			Name:        "float64-gauge",
			Description: "Gauge with float64 values",
			Unit:        "1",
			Data:        otelGaugeFloat64,
		},
		{
			Name:        "int64-sum",
			Description: "Sum with int64 values",
			Unit:        "1",
			Data:        otelSumInt64,
		},
	}
	otelScopeMetrics := []metricdata.ScopeMetrics{{
		Scope: instrumentation.Scope{
			Name:      "test-poc-otlpexp",
			Version:   "v0.1.0",
			SchemaURL: semconv.SchemaURL,
		},
		Metrics: otelMetrics,
	}}
	otelRes := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("test-poc-otlpexp"),
		semconv.ServiceVersion("v0.1.0"),
	)
	otelResourceMetrics := &metricdata.ResourceMetrics{
		Resource:     otelRes,
		ScopeMetrics: otelScopeMetrics,
	}
	return otelResourceMetrics
}
