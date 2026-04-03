package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	TraceIDKey   contextKey = "trace_id"
	SpanIDKey    contextKey = "span_id"
	ParentSpanID contextKey = "parent_span_id"
	SampledKey   contextKey = "sampled"
	JWTClaimsKey contextKey = "jwt_claims"
)

// IstioHeadersMiddleware extracts B3 trace headers and generic JWT info
func IstioHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract B3 Headers for tracing
			if traceID := r.Header.Get("x-b3-traceid"); traceID != "" {
				ctx = context.WithValue(ctx, TraceIDKey, traceID)
			}
			if spanID := r.Header.Get("x-b3-spanid"); spanID != "" {
				ctx = context.WithValue(ctx, SpanIDKey, spanID)
			}
			if parentSpanID := r.Header.Get("x-b3-parentspanid"); parentSpanID != "" {
				ctx = context.WithValue(ctx, ParentSpanID, parentSpanID)
			}
			if sampled := r.Header.Get("x-b3-sampled"); sampled != "" {
				ctx = context.WithValue(ctx, SampledKey, sampled)
			}

			// Extract JWT payload or Authorization header 
			if auth := r.Header.Get("Authorization"); auth != "" {
				ctx = context.WithValue(ctx, JWTClaimsKey, auth)
			} // Note: Istio RequestAuthentication can also be configured to pass claims in custom headers

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
