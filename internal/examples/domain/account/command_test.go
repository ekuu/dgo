package account

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ekuu/dgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestCreateCmd_Handle(t *testing.T) {
	initTracer()
	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		//slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		//AddSource: true,
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(dgo.TraceSlog(handler)))

	//slog.SetDefault(slog.New(itrace.SlogHandler(slog.Default().Handler())))
	//c := &CreateCmd{Name: "ss", Balance: 100}
	//a, err := dgo.handle[*Account](context.Background(), c, New(dgo.NewAggBase()))
	//spew.Dump(a, err)
	//time.Sleep(20 * time.Second)
}

func initTracer() {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	conn, err := grpc.DialContext(
		ctx,
		"127.0.0.1:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		panic(err)
	}

	otlpTraceExporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithGRPCConn(conn),
	)
	tp := trace.NewTracerProvider(
		//trace.WithBatcher(exp),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(newResource()),
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(otlpTraceExporter)),
	)
	otel.SetTracerProvider(tp)

}

// newResource returns a resource describing this application.
func newResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("dgo-test"),
			//semconv.ServiceVersion("v0.1.0"),
			//attribute.String("environment", "demo"),
		),
	)
	return r
}
