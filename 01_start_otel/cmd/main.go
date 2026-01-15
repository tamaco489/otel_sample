package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func main() {
	// 1. Exporter を作成 (標準出力に出力)
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	// 2. Resource を作成 (サービス情報)
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("hello-trace"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 3. TracerProvider を作成
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	defer tp.Shutdown(context.Background())

	// 4. グローバルに登録
	otel.SetTracerProvider(tp)

	// 5. Tracer を取得
	tracer := otel.Tracer("hello-tracer")

	// 6. Span を作成
	ctx, span := tracer.Start(context.Background(), "hello-world")
	defer span.End()

	// 7. 子 Span を作成
	_, childSpan := tracer.Start(ctx, "say-hello")
	log.Println("Hello, OpenTelemetry!")
	childSpan.End()

	log.Println("Trace generated! Check the output above.")
}

/* Output Example:
2026/01/13 23:52:18 Hello, OpenTelemetry!
2026/01/13 23:52:18 Trace generated! Check the output above.
{
        "Name": "say-hello",
        "SpanContext": {
                "TraceID": "1242c8079787512c6c461428c6713b4e",
                "SpanID": "a27f8dddb8110057",
                "TraceFlags": "01",
                "TraceState": "",
                "Remote": false
        },
        "Parent": {
                "TraceID": "1242c8079787512c6c461428c6713b4e",
                "SpanID": "c24efa1bc42cabb5",
                "TraceFlags": "01",
                "TraceState": "",
                "Remote": false
        },
        "SpanKind": 1,
        "StartTime": "2026-01-13T23:52:18.209153627+09:00",
        "EndTime": "2026-01-13T23:52:18.209196624+09:00",
        "Attributes": null,
        "Events": null,
        "Links": null,
        "Status": {
                "Code": "Unset",
                "Description": ""
        },
        "DroppedAttributes": 0,
        "DroppedEvents": 0,
        "DroppedLinks": 0,
        "ChildSpanCount": 0,
        "Resource": [
                {
                        "Key": "service.name",
                        "Value": {
                                "Type": "STRING",
                                "Value": "hello-trace"
                        }
                },
                {
                        "Key": "service.version",
                        "Value": {
                                "Type": "STRING",
                                "Value": "1.0.0"
                        }
                }
        ],
        "InstrumentationScope": {
                "Name": "hello-tracer",
                "Version": "",
                "SchemaURL": "",
                "Attributes": null
        },
        "InstrumentationLibrary": {
                "Name": "hello-tracer",
                "Version": "",
                "SchemaURL": "",
                "Attributes": null
        }
}
{
        "Name": "hello-world",
        "SpanContext": {
                "TraceID": "1242c8079787512c6c461428c6713b4e",
                "SpanID": "c24efa1bc42cabb5",
                "TraceFlags": "01",
                "TraceState": "",
                "Remote": false
        },
        "Parent": {
                "TraceID": "00000000000000000000000000000000",
                "SpanID": "0000000000000000",
                "TraceFlags": "00",
                "TraceState": "",
                "Remote": false
        },
        "SpanKind": 1,
        "StartTime": "2026-01-13T23:52:18.209137571+09:00",
        "EndTime": "2026-01-13T23:52:18.209215479+09:00",
        "Attributes": null,
        "Events": null,
        "Links": null,
        "Status": {
                "Code": "Unset",
                "Description": ""
        },
        "DroppedAttributes": 0,
        "DroppedEvents": 0,
        "DroppedLinks": 0,
        "ChildSpanCount": 1,
        "Resource": [
                {
                        "Key": "service.name",
                        "Value": {
                                "Type": "STRING",
                                "Value": "hello-trace"
                        }
                },
                {
                        "Key": "service.version",
                        "Value": {
                                "Type": "STRING",
                                "Value": "1.0.0"
                        }
                }
        ],
        "InstrumentationScope": {
                "Name": "hello-tracer",
                "Version": "",
                "SchemaURL": "",
                "Attributes": null
        },
        "InstrumentationLibrary": {
                "Name": "hello-tracer",
                "Version": "",
                "SchemaURL": "",
                "Attributes": null
        }
}
*/
