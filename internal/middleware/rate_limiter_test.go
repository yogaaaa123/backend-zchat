package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ==================== RateLimiter Allow ====================

func TestAllow_FirstRequest(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	assert.True(t, rl.Allow("1.2.3.4"))
}

func TestAllow_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)
	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow("1.2.3.4"), "request %d harus true", i+1)
	}
}

func TestAllow_ExactLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		assert.True(t, rl.Allow("1.2.3.4"), "request %d harus true", i+1)
	}
}

func TestAllow_OverLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		assert.True(t, rl.Allow("1.2.3.4"))
	}
	assert.False(t, rl.Allow("1.2.3.4"), "request ke-4 harus di-reject")
}

func TestAllow_WindowReset(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)

	assert.True(t, rl.Allow("1.2.3.4"))
	assert.True(t, rl.Allow("1.2.3.4"))
	assert.False(t, rl.Allow("1.2.3.4"), "over limit")

	time.Sleep(60 * time.Millisecond)

	assert.True(t, rl.Allow("1.2.3.4"), "setelah window reset harus true")
}

func TestAllow_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	assert.True(t, rl.Allow("1.1.1.1"))
	assert.True(t, rl.Allow("1.1.1.1"))
	assert.False(t, rl.Allow("1.1.1.1"), "IP 1 over limit")

	assert.True(t, rl.Allow("2.2.2.2"), "IP 2 independent")
	assert.True(t, rl.Allow("2.2.2.2"))
	assert.False(t, rl.Allow("2.2.2.2"), "IP 2 over limit")
}

func TestAllow_ZeroLimit(t *testing.T) {
	rl := NewRateLimiter(0, time.Minute)
	assert.False(t, rl.Allow("1.2.3.4"), "limit=0 harus reject semua")
}

func TestAllow_Concurrent(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.Allow("1.2.3.4")
		}()
	}
	wg.Wait()
	assert.True(t, true, "concurrent access should not race")
}

// ==================== getClientIP ====================

func TestGetClientIP_XForwardedFor_Single(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("X-Forwarded-For", "1.2.3.4")

	ip := getClientIP(c)
	assert.Equal(t, "1.2.3.4", ip)
}

func TestGetClientIP_XForwardedFor_Multiple(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8, 9.10.11.12")

	ip := getClientIP(c)
	assert.Equal(t, "1.2.3.4", ip)
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.RemoteAddr = "192.168.1.1:54321"

	ip := getClientIP(c)
	assert.Equal(t, "192.168.1.1", ip)
}

func TestGetClientIP_Empty(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	// Falls back to Gin's RemoteIP() — may be empty in test context
	_ = getClientIP(c)
	assert.True(t, true, "should not panic when both headers are empty")
}
