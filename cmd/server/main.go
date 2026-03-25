package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/amaumene/snowfinder-common/config"
	"github.com/amaumene/snowfinder-common/repository"
	"github.com/amaumene/snowfinder-web/internal/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultPort = "8080"

type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string][]time.Time
	limit    int
	window   time.Duration
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	return &ipRateLimiter{
		visitors: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old entries
	visits := rl.visitors[ip]
	clean := visits[:0]
	for _, t := range visits {
		if t.After(cutoff) {
			clean = append(clean, t)
		}
	}

	if len(clean) >= rl.limit {
		rl.visitors[ip] = clean
		return false
	}

	rl.visitors[ip] = append(clean, now)
	return true
}

func rateLimitMiddleware(limiter *ipRateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.allow(r.RemoteAddr) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://cdn.jsdelivr.net; style-src 'self' https://cdn.jsdelivr.net 'unsafe-inline'")
		next.ServeHTTP(w, r)
	})
}

func main() {
	ctx := context.Background()

	cfg := config.Default()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = cfg.DatabaseURL
	}

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Connect to database: %v", err)
	}
	defer db.Close()

	// Use read-only repository for web interface
	repo := repository.NewReader(db)

	handler, err := handlers.NewHandler(repo)
	if err != nil {
		log.Fatalf("Create web handler: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.IndexHandler)
	mux.HandleFunc("/about", handler.AboutHandler)
	mux.HandleFunc("/api/search", handler.SearchHandler)
	mux.HandleFunc("/api/resorts-with-peaks", handler.ResortsWithPeaksHandler)
	mux.HandleFunc("/api/peak-info", handler.PeakInfoHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Apply middleware: rate limiting -> security headers -> mux
	limiter := newIPRateLimiter(30, time.Minute) // 30 requests per minute per IP
	wrappedHandler := rateLimitMiddleware(limiter, securityHeaders(mux))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      wrappedHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("SnowFinder Web Server starting on http://localhost:%s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server stopped")
}
