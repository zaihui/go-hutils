package hutils

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"

	// nolint:staticcheck
	// ignore SA1019 Need to keep deprecated package for compatibility.
	"github.com/golang/protobuf/proto"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.elastic.co/apm"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const (
	HTTP2Protocol = "HTTP/2"
	LocalHost     = "127.0.0.1"
)

func MarshalJSON(msg interface{}) ([]byte, error) {
	if pb, ok := msg.(proto.Message); ok {
		b := &bytes.Buffer{}
		if err := grpc_zap.JsonPbMarshaller.Marshal(b, pb); err != nil {
			return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
		}
		return b.Bytes(), nil
	}
	return nil, fmt.Errorf("msg not valid: %v", msg)
}

func SetTrace(ctx context.Context, name string, apmTracer *apm.Tracer) context.Context {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(tp)
	tracer := otel.GetTracerProvider().Tracer("grpc")
	ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()
	if apmTracer != nil {
		tx := apmTracer.StartTransaction(name, "grpc")
		defer tx.End()
		ctx = apm.ContextWithTransaction(ctx, tx)
	}
	return ctx
}

// NewUnaryServerAccessLogInterceptor returns a new unary server interceptors tha log access log
func NewUnaryServerAccessLogInterceptor(logger *zap.SugaredLogger, apmTracer *apm.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = SetTrace(ctx, info.FullMethod, apmTracer)
		startTime := time.Now()
		ip, _ := peer.FromContext(ctx)
		resp, err := handler(ctx, req)
		clientIP := strings.Split(ip.Addr.String(), ":")[0]
		// ignore probe requests
		if clientIP == LocalHost {
			return resp, err
		}
		code := grpc_logging.DefaultErrorToCode(err)
		l := UnionLog{
			ClientIP:   clientIP,
			Request:    info.FullMethod,
			Protocol:   HTTP2Protocol,
			Duration:   time.Since(startTime).Milliseconds(),
			LogType:    grpcLogType,
			GrpcStatus: code.String(),
		}
		if msg, ok := req.(proto.Message); ok {
			l.Payload, _ = MarshalJSON(msg)
		}
		if msg, ok := resp.(proto.Message); ok {
			l.Response, _ = MarshalJSON(msg)
		}
		l.Log(ctx, logger)
		return resp, err
	}
}
