package tools

import (
	"sync"
	"golang.org/x/time/rate"
	"time"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mtx  sync.Mutex

	visitors map[string]*visitor
}

func NewRateLimiter() *RateLimiter {
	controller := &RateLimiter{}
	controller.visitors = make(map[string]*visitor)

	go controller.cleanupVisitors()

	return controller
}

func (c *RateLimiter) addVisitor(appid string) *rate.Limiter {
	limiter := rate.NewLimiter(2, 5)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	// Include the current time when creating a new visitor.
	c.visitors[appid] = &visitor{limiter, time.Now()}

	return limiter
}

func (c *RateLimiter) GetVisitor(appid string) *rate.Limiter {
	c.mtx.Lock()

	v, exists := c.visitors[appid]
	if !exists {
		c.mtx.Unlock()
		return c.addVisitor(appid)
	}

	// Updatae the last seen time for the visitor.
	v.lastSeen = time.Now()
	c.mtx.Unlock()
	return v.limiter
}

func (c *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		c.mtx.Lock()
		for appid, v := range c.visitors {
			if (time.Now().Sub(v.lastSeen) > 3 * time.Minute) {
				delete(c.visitors, appid)
			}
		}
		c.mtx.Unlock()
	}
}


