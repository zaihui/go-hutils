package hutils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/SkyAPM/go2sky"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

var (
	ComponentIDGOHttpServer int32 = 5004
)

type operation func(name string, r *http.Request) string

type handler struct {
	tracer    *go2sky.Tracer
	name      string
	next      http.Handler
	extraTags map[string]string
	// filter some health check request.
	filterURLs []string
	// get operation name.
	operationFunc operation
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        []byte
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func WithFilterURL(urls []string) func(*handler) {
	return func(options *handler) {
		options.filterURLs = urls
	}
}

func WithOperation(operation operation) func(*handler) {
	return func(options *handler) {
		options.operationFunc = operation
	}
}

func WithExtraTags(tags map[string]string) func(*handler) {
	return func(options *handler) {
		options.extraTags = tags
	}
}

// FilterURL filter url not proceed trace.
func FilterURL(filterURLs []string, url string) bool {
	for _, filter := range filterURLs {
		if filter == url {
			return true
		}
	}
	return false
}

func NewServerSkywalkingMiddleware(tracer *go2sky.Tracer, opts ...func(*handler)) (func(http.Handler) http.Handler, error) {
	if tracer == nil {
		panic("tracer is nil.")
	}
	return func(next http.Handler) http.Handler {
		h := &handler{
			tracer: tracer,
			next:   next,
			operationFunc: func(name string, r *http.Request) string {
				return name
			},
		}
		for _, o := range opts {
			o(h)
		}
		return h
	}, nil
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if FilterURL(h.filterURLs, r.URL.Path) {
		h.next.ServeHTTP(w, r)
		return
	}
	span, ctx, err := h.tracer.CreateEntrySpan(r.Context(), h.operationFunc(h.name, r), func(key string) (string, error) {
		return r.Header.Get(key), nil
	})
	if err != nil {
		if h.next != nil {
			h.next.ServeHTTP(w, r)
		}
		return
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}

	span.SetComponent(ComponentIDGOHttpServer)
	span.Tag(go2sky.TagHTTPMethod, r.Method)
	span.Tag(go2sky.TagURL, fmt.Sprintf("%s%s", r.Host, r.URL.Path))
	span.SetSpanLayer(v3.SpanLayer_Http)
	span.Log(time.Now(), ReqTag, Param(payload))
	for k, v := range h.extraTags {
		span.Tag(go2sky.Tag(k), v)
	}

	r.Body = io.NopCloser(bytes.NewBuffer(payload))
	rw := wrapResponseWriter(w)
	defer func() {
		if e := recover(); e != nil {
			span.Error(time.Now(), RespTag, MarshalParam(e))
			span.Tag(go2sky.TagStatusCode, strconv.Itoa(500))
			span.End()
			panic(e)
		} else {
			if rw.status >= 400 {
				span.Error(time.Now(), RespTag, string(rw.body))
			} else {
				span.Log(time.Now(), RespTag, string(rw.body))
			}
			span.Tag(go2sky.TagStatusCode, strconv.Itoa(rw.status))
			span.End()
		}
	}()
	if h.next != nil {
		h.next.ServeHTTP(rw, r.WithContext(ctx))
	}
}

func Param(param []byte) string {
	str := string(param)
	if len(str) == 0 {
		return "空"
	}
	return str
}
