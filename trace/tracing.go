package tracing

import (
	"context"

	"github.com/micro/go-micro/metadata"
	"github.com/opentracing/opentracing-go"
	zipkintracer "github.com/openzipkin/zipkin-go-opentracing"

	"github.com/mmqbaba/zeus/errors"
)

type TracerWrap struct {
	tracer opentracing.Tracer
}

func NewTracerWrap(tracer opentracing.Tracer) *TracerWrap {
	return &TracerWrap{
		tracer: tracer,
	}
}

// StartSpanFromContext returns a new span with the given operation name and options. If a span
// is found in the context, it will be used as the parent of the resulting span.
func (t *TracerWrap) StartSpanFromContext(ctx context.Context, name string,
	opts ...opentracing.StartSpanOption) (context.Context, opentracing.Span, error) {
	tracer := t.tracer
	if tracer == nil {
		return ctx, nil, errors.New(errors.ECodeSystem, "", "")
	}
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(map[string]string)
	}

	// copy the metadata to prevent race
	md = metadata.Copy(md)

	// find trace in go-micro metadata
	if spanCtx, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(md)); err == nil {
		opts = append(opts, opentracing.ChildOf(spanCtx))
	}

	// find span context in opentracing library
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
	}

	sp := tracer.StartSpan(name, opts...)

	if err := sp.Tracer().Inject(sp.Context(), opentracing.TextMap, opentracing.TextMapCarrier(md)); err != nil {
		return nil, nil, err
	}

	ctx = opentracing.ContextWithSpan(ctx, sp)
	ctx = metadata.NewContext(ctx, md)
	return ctx, sp, nil
}

func (t *TracerWrap) GetTraceID(ctx context.Context) string {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		return span.Context().(zipkintracer.SpanContext).TraceID.ToHex()
	}
	return ""
}

func (t *TracerWrap) GetSpan(ctx context.Context) opentracing.Span {
	return opentracing.SpanFromContext(ctx)
}
