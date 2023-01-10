package hutils

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"

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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const (
	ReqTag  = "[请求参数]"
	RespTag = "[响应结果]"

	HTTP2Protocol         = "HTTP/2"
	LocalHost             = "127.0.0.1"
	ComponentIDGrpcClient = 5013
	ComponentIDGrpcGo     = 23
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

type Option func(*options)

type options struct {
	// custom span tags form MD.
	reportTags []string
	// filter some health check request.
	filterMethods []string
}

func WithFilterMethod(methods []string) func(*options) {
	return func(options *options) {
		options.filterMethods = methods
	}
}

func WithReportTags(tags []string) func(*options) {
	return func(options *options) {
		options.reportTags = tags
	}
}

func MarshalParam(v interface{}) string {
	json, err := JSONMarshal(v)
	if err != nil {
		return ""
	}
	return string(json)
}

// FilterMethod filter method not proceed trace.
func FilterMethod(filterMethods []string, method string) bool {
	for _, filter := range filterMethods {
		if filter == method {
			return true
		}
	}
	return false
}

// NewUnaryClientSkywalkingInterceptor skywalking client interceptor.
func NewUnaryClientSkywalkingInterceptor(tracer *go2sky.Tracer) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		span, ctx, err := tracer.CreateEntrySpan(ctx, method, func(key string) (string, error) {
			return "", nil
		})
		if err != nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		s, ok := span.(go2sky.ReportedSpan)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		defer func() {
			span.SetComponent(ComponentIDGrpcClient)
			span.SetSpanLayer(v3.SpanLayer_RPCFramework)
			span.Log(time.Now(), ReqTag, MarshalParam(req))
			if err != nil {
				span.Error(time.Now(), RespTag, MarshalParam(err))
			} else {
				span.Log(time.Now(), RespTag, MarshalParam(reply))
			}
			span.End()
		}()
		spanContext := &propagation.SpanContext{
			Sample:             1,
			TraceID:            s.Context().TraceID,
			ParentSegmentID:    s.Context().SegmentID,
			ParentSpanID:       s.Context().SpanID,
			ParentEndpoint:     s.Context().FirstSpan.GetOperationName(),
			CorrelationContext: s.Context().CorrelationContext,
		}
		ctx = metadata.NewOutgoingContext(ctx, map[string][]string{
			propagation.Header:            {spanContext.EncodeSW8()},
			propagation.HeaderCorrelation: {spanContext.EncodeSW8Correlation()},
		})
		err = invoker(ctx, method, req, reply, cc, opts...)
		return err
	}
}

// NewUnaryServerSkywalkingInterceptor skywalking server interceptor.
// nolint: govet
func NewUnaryServerSkywalkingInterceptor(tracer *go2sky.Tracer, opts ...Option) grpc.UnaryServerInterceptor {
	options := &options{
		reportTags:    []string{},
		filterMethods: []string{},
	}
	for _, o := range opts {
		o(options)
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if FilterMethod(options.filterMethods, info.FullMethod) {
			return handler(ctx, req)
		}
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			span, ctx, err := tracer.CreateEntrySpan(ctx, info.FullMethod, func(key string) (string, error) {
				return strings.Join(md.Get(key), ""), nil
			})
			if err != nil {
				return handler(ctx, req)
			}
			for _, k := range options.reportTags {
				span.Tag(go2sky.Tag(k), strings.Join(md.Get(k), ""))
			}
			var reply interface{}
			defer func() {
				span.SetComponent(ComponentIDGrpcGo)
				span.SetSpanLayer(v3.SpanLayer_RPCFramework)
				span.Log(time.Now(), ReqTag, MarshalParam(req))
				if err != nil {
					span.Error(time.Now(), RespTag, MarshalParam(err))
				} else {
					span.Log(time.Now(), RespTag, MarshalParam(reply))
				}
				span.End()
			}()
			reply, err = handler(ctx, req)
			return reply, err
		}
		return handler(ctx, req)
	}
}
