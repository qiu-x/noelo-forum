package limit

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	lim, exists := i.ips[ip]
	if !exists {
		lim = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = lim
	}
	return lim
}

func (i *IPRateLimiter) CheckIP(ip string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	lim := i.getLimiter(ip)
	return lim.Allow()
}

func (i *IPRateLimiter) IPLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !i.CheckIP(ip) {
			http.Error(w, http.StatusText(http.StatusTooManyRequests),
				http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
