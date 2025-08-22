package miniws

import (
	"log"
	"maps"

	"golang.org/x/time/rate"
)

const CRL_MAX_CLIENT_REMEMBER_SEC float64 = 60.

type clientRateLimiter struct {
	limits         map[string]*rate.Limiter
	maxConnsPerSec float64
}

func newClientRateLimiter(maxConnectionsPerMin float64) *clientRateLimiter {
	return &clientRateLimiter{
		limits:         make(map[string]*rate.Limiter),
		maxConnsPerSec: maxConnectionsPerMin / 60.,
	}
}

func (crl *clientRateLimiter) canConnect(clientIp string) bool {
	limiter, ok := crl.limits[clientIp]
	if !ok {
		crl.limits[clientIp] = rate.NewLimiter(rate.Limit(crl.maxConnsPerSec), 1)
		limiter = crl.limits[clientIp]
		log.Println("new client: " + clientIp)
	}
	allowed := limiter.Allow()
	if allowed {
		log.Println("client " + clientIp + " was allowed")
	} else {
		log.Println("client " + clientIp + " has been rate limited")
	}
	crl._cleanup()
	return allowed
}

func (crl *clientRateLimiter) _cleanup() {
	maps.DeleteFunc(crl.limits, func(clIp string, limit *rate.Limiter) bool {
		// forget the client if they haven't connected in CRL_MAX_CLIENT_REMEMBER_SEC seconds
		forgetClient := limit.Tokens() >= CRL_MAX_CLIENT_REMEMBER_SEC*crl.maxConnsPerSec
		if forgetClient {
			log.Println("Forgetting client "+clIp+": accumulated tokens ", limit.Tokens(),
				"exceed limit of ", CRL_MAX_CLIENT_REMEMBER_SEC*crl.maxConnsPerSec)
		}
		return forgetClient
	})
}
