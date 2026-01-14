package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	bucket   *rate.Limiter
	lastSeen time.Time
}

// The manager for holding clients
type RateLimiterObj struct {
	clients map[string]*Client // IP -> client
	mu      sync.Mutex
	rate    rate.Limit //refill speed
	burst   int        //bucket capacity
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiterObj {
	m := &RateLimiterObj{
		clients: make(map[string]*Client),
		rate:    r,
		burst:   b,
	}
	go m.cleanup()
	return m
}

// create/get bucket for a client
func (m *RateLimiterObj) getBucket(ip string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[ip]

	if !exists {
		client = &Client{
			bucket: rate.NewLimiter(m.rate, m.burst),
		}
		m.clients[ip] = client
	}
	client.lastSeen = time.Now()
	return client.bucket
}

// Background cleaner process
func (m *RateLimiterObj) cleanup() {
	ticker := time.NewTicker(1 * time.Minute) //check every 1 minute
	for {
		<-ticker.C
		m.mu.Lock()

		for ip, client := range m.clients {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(m.clients, ip) //user inactive for >3mins
			}
		}
		m.mu.Unlock()
	}
}

// The middleware
func (m *RateLimiterObj) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr // fallback if no port
		}

		bucket := m.getBucket(ip)
		if !bucket.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
