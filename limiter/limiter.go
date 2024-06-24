package limiter

import (
	"context"
	"encoding/json"
	"limiter/config"
	"limiter/lib"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// init funcion to load data and spawn a goroutine for cleanup
func InitData() {
	LoadClients()
	ticker := time.NewTicker(2 * time.Minute)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func(done chan os.Signal) {
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				for ip, client := range clients {
					logger.Info("Request Dropped", zap.Int(ip, client.Dropped()))
					if time.Since(client.LastSeen()) > 3*time.Minute {
						delete(clients, ip)
					}
				}
				mu.Unlock()

			case <-done:
				return
			}
		}
	}(done)
}

// Interface that exposes configurable limiter functions
type DynamicConfigurableLimiter interface {
	Allow() bool
	Wait(ctx context.Context) error
	LastSeen() time.Time
	SetLastSeen(t time.Time)
	AddDropped()
	Dropped() int
}

// dynamic limiter
type dynamicConfigurableLimiter struct {
	rps            float64
	burst          int
	lastSeen       time.Time
	limiterlib     lib.RateLimiterInterface
	requestDropped int
}

type Message struct {
	Status string
	Body   string
}

var (
	mu sync.Mutex
	//map holds client IP against limiter
	clients = make(map[string]DynamicConfigurableLimiter)
	logger  = zap.NewExample()
	done    = make(chan os.Signal, 1)
)

// Inititaing limiter
func NewConfigurableLimiter(rps float64, b int) DynamicConfigurableLimiter {
	return &dynamicConfigurableLimiter{rps: rps, burst: b, limiterlib: lib.NewRateLimiterInterface(rps, b)}

}

func (d *dynamicConfigurableLimiter) Dropped() int {
	return int(d.requestDropped)
}

func (d *dynamicConfigurableLimiter) AddDropped() {
	d.requestDropped++
}

func (d *dynamicConfigurableLimiter) Allow() bool {
	return d.limiterlib.Allow()
}

func (d *dynamicConfigurableLimiter) Wait(ctx context.Context) error {
	return d.limiterlib.Wait(ctx)
}

func (d *dynamicConfigurableLimiter) LastSeen() time.Time {
	return d.lastSeen
}

func (d *dynamicConfigurableLimiter) SetLastSeen(t time.Time) {
	d.lastSeen = t
}

// function that limits the rate and returns a http.Handler
func Limit(next func(writer http.ResponseWriter, request *http.Request)) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the IP address from the request.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("error in extracting Ip", zap.Error(err))
			return
		}
		// Lock the mutex to protect this section from race conditions.
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = NewConfigurableLimiter(2, 4)
		}
		clients[ip].SetLastSeen(time.Now())

		if !clients[ip].Allow() {
			//unlock if request doest not allow
			mu.Unlock()
			clients[ip].AddDropped()
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, try again later.",
			}

			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		}
		mu.Unlock()
		next(w, r)
	})
}

// creating clients records in middleware
func LoadClients() {
	conf, err := config.LoadConfig(".")
	if err != nil {
		logger.Error("error in loading config date grace shutdown", zap.Error(err))
		close(done)
		os.Exit(1)
	}
	for _, c := range conf.ClientAddrs {
		clients[c.InetAddr] = NewConfigurableLimiter(c.Rate, c.Burst)
	}
}
