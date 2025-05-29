package gorgo

import (
	"log"
	"time"
)

// MiddlewareFunc defines a middleware function
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// MiddlewareChain represents a middleware chain
type MiddlewareChain struct {
	middlewares []MiddlewareFunc
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain(middlewares ...MiddlewareFunc) *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: middlewares,
	}
}

// Add adds middleware to the chain
func (mc *MiddlewareChain) Add(middleware MiddlewareFunc) *MiddlewareChain {
	mc.middlewares = append(mc.middlewares, middleware)
	return mc
}

// Execute executes the middleware chain
func (mc *MiddlewareChain) Execute(handler HandlerFunc) HandlerFunc {
	// Apply middleware in reverse order
	for i := len(mc.middlewares) - 1; i >= 0; i-- {
		handler = mc.middlewares[i](handler)
	}
	return handler
}

// Built-in middleware

// LoggerMiddleware logs requests
func LoggerMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			start := time.Now()

			// Execute next handler
			err := next(ctx)

			// Log
			duration := time.Since(start)
			method := string(ctx.fastCtx.Method())
			path := string(ctx.fastCtx.Path())
			status := ctx.fastCtx.Response.StatusCode()

			log.Printf("%s %s %d %v", method, path, status, duration)

			return err
		}
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic recovered: %v", r)
					ctx.fastCtx.SetStatusCode(500)
					ctx.fastCtx.SetBodyString("Internal Server Error")
				}
			}()

			return next(ctx)
		}
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(options CORSOptions) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			// Set CORS headers
			if options.AllowOrigin != "" {
				ctx.fastCtx.Response.Header.Set("Access-Control-Allow-Origin", options.AllowOrigin)
			}

			if len(options.AllowMethods) > 0 {
				methods := ""
				for i, method := range options.AllowMethods {
					if i > 0 {
						methods += ", "
					}
					methods += method
				}
				ctx.fastCtx.Response.Header.Set("Access-Control-Allow-Methods", methods)
			}

			if len(options.AllowHeaders) > 0 {
				headers := ""
				for i, header := range options.AllowHeaders {
					if i > 0 {
						headers += ", "
					}
					headers += header
				}
				ctx.fastCtx.Response.Header.Set("Access-Control-Allow-Headers", headers)
			}

			if options.AllowCredentials {
				ctx.fastCtx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if string(ctx.fastCtx.Method()) == "OPTIONS" {
				ctx.fastCtx.SetStatusCode(200)
				return nil
			}

			return next(ctx)
		}
	}
}

// CORSOptions configuration for CORS
type CORSOptions struct {
	AllowOrigin      string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
}

// DefaultCORSOptions returns default CORS settings
func DefaultCORSOptions() CORSOptions {
	return CORSOptions{
		AllowOrigin:      "*",
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}
}

// RateLimitMiddleware limits the number of requests
func RateLimitMiddleware(options RateLimitOptions) MiddlewareFunc {
	limiter := NewRateLimiter(options)

	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			clientIP := string(ctx.fastCtx.RemoteIP())

			if !limiter.Allow(clientIP) {
				ctx.fastCtx.SetStatusCode(429)
				ctx.fastCtx.SetBodyString("Too Many Requests")
				return nil
			}

			return next(ctx)
		}
	}
}

// RateLimitOptions configuration for rate limiting
type RateLimitOptions struct {
	RequestsPerMinute int
	BurstSize         int
}

// Simple rate limiter implementation
type RateLimiter struct {
	options RateLimitOptions
	clients map[string]*ClientLimiter
}

type ClientLimiter struct {
	lastRequest time.Time
	tokens      int
}

func NewRateLimiter(options RateLimitOptions) *RateLimiter {
	return &RateLimiter{
		options: options,
		clients: make(map[string]*ClientLimiter),
	}
}

func (rl *RateLimiter) Allow(clientID string) bool {
	now := time.Now()

	client, exists := rl.clients[clientID]
	if !exists {
		client = &ClientLimiter{
			lastRequest: now,
			tokens:      rl.options.BurstSize,
		}
		rl.clients[clientID] = client
	}

	// Add tokens based on time
	elapsed := now.Sub(client.lastRequest)
	tokensToAdd := int(elapsed.Minutes() * float64(rl.options.RequestsPerMinute))
	client.tokens += tokensToAdd

	if client.tokens > rl.options.BurstSize {
		client.tokens = rl.options.BurstSize
	}

	client.lastRequest = now

	// Check if there are available tokens
	if client.tokens > 0 {
		client.tokens--
		return true
	}

	return false
}

// AuthMiddleware checks authentication
func AuthMiddleware(authFunc func(ctx *Context) (interface{}, error)) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			user, err := authFunc(ctx)
			if err != nil {
				ctx.fastCtx.SetStatusCode(401)
				return ctx.JSON(Map{"error": "Unauthorized"})
			}

			// Save user in context
			ctx.Set("user", user)

			return next(ctx)
		}
	}
}

// CompressionMiddleware compresses responses
func CompressionMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			// Check if client supports compression
			acceptEncoding := string(ctx.fastCtx.Request.Header.Peek("Accept-Encoding"))

			if acceptEncoding != "" {
				// Set compression (FastHTTP supports this automatically)
				ctx.fastCtx.Response.Header.Set("Content-Encoding", "gzip")
			}

			return next(ctx)
		}
	}
}
