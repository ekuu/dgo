package app

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/davecgh/go-spew/spew"

	"github.com/ekuu/dgo"
	"github.com/ekuu/dgo/internal/examples/domain/account"
)

func TestTranslate(t *testing.T) {
	initTracer()
	initLog()
	err := Translate(context.Background(), account.NewTransferCmd(dgo.ID("0a8ac46d4b2944898bb6029a1509d3c1"), dgo.ID("649543d698cf4a4fac73356cb1e2a71e"), 1))
	if err != nil {
		t.Fatalf("%+v", err)
	}
	//time.Sleep(time.Millisecond * 300)
	time.Sleep(time.Second * 10)
}

func TestCreateAccount(t *testing.T) {
	initTracer()
	initLog()
	_, err := CreateAccount(context.Background(), &account.CreateCmd{Name: "lisi2", Balance: 0})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	//spew.Dump(a)
	//time.Sleep(time.Millisecond * 300)
	time.Sleep(time.Second * 10)
}

func TestUpdateAccountName(t *testing.T) {
	a, err := UpdateAccountName(context.Background(), &account.UpdateNameCmd{
		ID:   "acc_zhangsan11",
		Name: "test-name1",
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	spew.Dump(a)
	time.Sleep(time.Millisecond * 300)
}

func initLog() {
	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		//slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		//AddSource: true,
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(dgo.TraceSlog(handler)))
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
			semconv.ServiceName("dgo"),
			//semconv.ServiceVersion("v0.1.0"),
			//attribute.String("environment", "demo"),
		),
	)
	return r
}
