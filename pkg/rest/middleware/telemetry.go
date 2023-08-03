package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"

	"github.com/lrweck/clean-api/pkg/envutil"
	"github.com/lrweck/clean-api/pkg/slogger"
	"github.com/lrweck/clean-api/pkg/telemetry"
)

func OpenTelemetry() echo.MiddlewareFunc {

	tProvider := otel.GetTracerProvider()
	textProp := otel.GetTextMapPropagator()

	appName := envutil.AppName()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if Skipper(c) {
				return next(c)
			}

			start := time.Now()

			req := c.Request()

			ctx := textProp.Extract(req.Context(), propagation.HeaderCarrier(req.Header))

			commonLabels := []attribute.KeyValue{
				attribute.String(echo.HeaderXRequestID, req.Header.Get(echo.HeaderXRequestID)),
			}

			opts := []oteltrace.SpanStartOption{
				oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(appName, c.Path(), req)...),
				oteltrace.WithAttributes(commonLabels...),
				oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
			}

			spanName := c.Path()
			if spanName == "" {
				spanName = fmt.Sprintf("HTTP %s route not found", req.Method)
			}

			ctx, span := tProvider.Tracer(appName).Start(ctx, spanName, opts...)
			defer span.End()

			traceID := span.SpanContext().TraceID().String()
			spanID := span.SpanContext().SpanID().String()

			ctx = context.WithValue(ctx, telemetry.TraceID, traceID)
			ctx = context.WithValue(ctx, telemetry.SpanID, spanID)

			logger := slogger.FromContext(ctx)

			logger = logger.With(
				slog.String(string(telemetry.TraceID), traceID),
				slog.String(string(telemetry.SpanID), spanID),
			)

			ctx = context.WithValue(ctx, slogger.LoggerKey, logger)

			// pass the span through the request context
			c.SetRequest(req.WithContext(ctx))

			err := next(c)
			if err != nil {
				span.SetAttributes(attribute.String("http.error", err.Error()))
				span.RecordError(err)
			}

			attrs := semconv.HTTPAttributesFromHTTPStatusCode(c.Response().Status)
			spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(c.Response().Status)
			span.SetAttributes(attrs...)
			span.SetStatus(spanStatus, spanMessage)

			span.SetAttributes(
				attribute.Int64("http.latency", time.Since(start).Milliseconds()),
			)

			return err
		}
	}
}

func RequestMetrics() echo.MiddlewareFunc {
	counter, err := otel.Meter("request.count").
		Int64Counter("request.count",
			metric.WithUnit("1"))
	if err != nil {
		panic(err)
	}

	hist, err := otel.Meter("request.latency").
		Int64Histogram("request.latency",
			metric.WithUnit("ms"))
	if err != nil {
		panic(err)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if Skipper(c) {
				return next(c)
			}

			start := time.Now()
			err := next(c)
			latency := time.Since(start).Milliseconds()

			commonAttrs := []attribute.KeyValue{
				{
					Key:   attribute.Key("http.path"),
					Value: attribute.StringValue(c.Path()),
				},
				{
					Key:   attribute.Key("http.method"),
					Value: attribute.StringValue(c.Request().Method),
				},
				{
					Key:   attribute.Key("http.status_code"),
					Value: attribute.Int64Value(int64(c.Response().Status)),
				},
			}

			counter.Add(context.Background(), 1, metric.WithAttributes(commonAttrs...))
			hist.Record(context.Background(), latency, metric.WithAttributes(commonAttrs...))

			return err
		}
	}
}
