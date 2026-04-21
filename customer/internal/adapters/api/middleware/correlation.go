package middleware

import (
	"github.com/labstack/echo/v4"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
)

const CorrelationIDHeader = "X-Correlation-ID"

// CorrelationIDMiddleware extracts the correlation ID from the request header 
// and injects it into the context. If not present, it generates a new one.
func CorrelationIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		correlationID := req.Header.Get(CorrelationIDHeader)
		
		// Ensure correlation ID is in context
		ctx := ctxutil.WithCorrelationID(req.Context(), correlationID)
		
		// Update request with new context
		c.SetRequest(req.WithContext(ctx))
		
		// Add to response headers for traceability
		c.Response().Header().Set(CorrelationIDHeader, ctxutil.GetCorrelationID(ctx))
		
		return next(c)
	}
}
