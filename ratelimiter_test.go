package ratelimiter_test

import (
	"sync"
	"testing"
	"time"

	"github.com/TRAD3R/ratelimiter"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name     string
		rps      int
		interval time.Duration
	}{
		{
			name:     "normal case",
			rps:      10,
			interval: time.Second,
		},
		{
			name:     "high rps",
			rps:      1000,
			interval: time.Millisecond * 100,
		},
		{
			name:     "low rps",
			rps:      1,
			interval: time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := ratelimiter.NewRateLimiter(tt.rps, tt.interval)

			if rl == nil {
				t.Fatal("ratelimiter.NewRateLimiter returned nil")
			}

			if rl.GetRPS() != tt.rps {
				t.Errorf("Expected rps %d, got %d", tt.rps, rl.GetRPS())
			}

			if rl.GetInterval() != tt.interval {
				t.Errorf("Expected interval %v, got %v", tt.interval, rl.GetInterval())
			}

			available, _ := rl.CurrentState()
			if available != tt.rps {
				t.Errorf("Expected initial available tokens %d, got %d", tt.rps, available)
			}

			// Изначально бакет полон, поэтому токен уже доступен сейчас.
			if rl.GetRetryAfter().After(time.Now()) {
				t.Error("retryAfter should not be in the future when the bucket is full")
			}
		})
	}
}

func TestAllowRequest_Basic(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(3, time.Second)

	// Первые 3 запроса должны быть разрешены
	for i := 0; i < 3; i++ {
		if !rl.AllowRequest() {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4-й запрос должен быть отклонен
	if rl.AllowRequest() {
		t.Error("4th request should be denied")
	}
}

func TestAllowRequest_RefillAfterInterval(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(2, 50*time.Millisecond)

	// Исчерпываем лимит
	if !rl.AllowRequest() {
		t.Error("First request should be allowed")
	}
	if !rl.AllowRequest() {
		t.Error("Second request should be allowed")
	}
	if rl.AllowRequest() {
		t.Error("Third request should be denied")
	}

	// Ждем, пока накопится хотя бы один токен
	time.Sleep(60 * time.Millisecond)

	// После пополнения должны снова разрешить запросы
	if !rl.AllowRequest() {
		t.Error("Request after refill should be allowed")
	}
}

func TestAllowRequest_ZeroRPS(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(0, time.Second)

	// При нулевом RPS все запросы должны отклоняться
	if rl.AllowRequest() {
		t.Error("Request with 0 RPS should be denied")
	}
}

func TestCurrentState(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(5, time.Second)

	// Начальное состояние: бакет полон.
	available, wait := rl.CurrentState()
	if available != 5 {
		t.Errorf("Expected initial available tokens 5, got %d", available)
	}
	if wait != 0 {
		t.Errorf("Expected wait 0 when tokens are available, got %v", wait)
	}

	// Исчерпываем все токены.
	for i := 0; i < 5; i++ {
		if !rl.AllowRequest() {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	available, wait = rl.CurrentState()
	if available != 0 {
		t.Errorf("Expected 0 available tokens, got %d", available)
	}
	if wait <= 0 {
		t.Error("wait should be positive when no tokens are available")
	}
}

func TestConcurrentAccess(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(100, time.Second)

	var wg sync.WaitGroup
	allowed := make(chan bool, 200)

	// Запускаем 200 горутин, каждая делает запрос
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- rl.AllowRequest()
		}()
	}

	wg.Wait()
	close(allowed)

	// Подсчитываем разрешенные запросы
	count := 0
	for result := range allowed {
		if result {
			count++
		}
	}

	// Должно быть разрешено ровно 100 запросов
	if count != 100 {
		t.Errorf("Expected 100 allowed requests, got %d", count)
	}
}

func TestConcurrentStateAccess(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(10, time.Second)

	var wg sync.WaitGroup

	// Запускаем горутины, которые одновременно делают запросы и читают состояние
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.AllowRequest()
			rl.CurrentState()
		}()
	}

	wg.Wait()

	// Проверяем, что состояние корректное
	available, _ := rl.CurrentState()
	if available > 10 {
		t.Errorf("Available tokens should not exceed RPS limit, got %d", available)
	}
}

func TestEdgeCases(t *testing.T) {
	// Тест с очень маленьким интервалом
	t.Run("very small interval", func(t *testing.T) {
		rl := ratelimiter.NewRateLimiter(1, time.Microsecond)

		if !rl.AllowRequest() {
			t.Error("First request should be allowed")
		}

		// Ждем микросекунду
		time.Sleep(2 * time.Microsecond)

		if !rl.AllowRequest() {
			t.Error("Request after microsecond should be allowed")
		}
	})

	// Тест с очень большим интервалом
	t.Run("very large interval", func(t *testing.T) {
		rl := ratelimiter.NewRateLimiter(1, time.Hour)

		if !rl.AllowRequest() {
			t.Error("First request should be allowed")
		}

		if rl.AllowRequest() {
			t.Error("Second request should be denied")
		}
	})

	// Тест с отрицательным RPS (должен обрабатываться как 0)
	t.Run("negative RPS", func(t *testing.T) {
		rl := ratelimiter.NewRateLimiter(-1, time.Second)

		if rl.AllowRequest() {
			t.Error("Request with negative RPS should be denied")
		}
	})
}

func TestRateLimiter_RefillBehavior(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(2, 100*time.Millisecond)

	// Исчерпываем лимит
	rl.AllowRequest()
	rl.AllowRequest()

	if rl.AllowRequest() {
		t.Error("Third request should be denied")
	}

	// Ждем, пока бакет полностью пополнится
	time.Sleep(110 * time.Millisecond)

	// Токены пополняются непрерывно, поэтому CurrentState сразу видит полный бакет.
	available, _ := rl.CurrentState()
	if available != 2 {
		t.Errorf("Expected 2 available tokens after refill, got %d", available)
	}

	// Теперь должны снова разрешать запросы
	if !rl.AllowRequest() {
		t.Error("Request after refill should be allowed")
	}

	// Один токен израсходован.
	available, _ = rl.CurrentState()
	if available != 1 {
		t.Errorf("Expected 1 available token after one request, got %d", available)
	}
}

func BenchmarkAllowRequest(b *testing.B) {
	rl := ratelimiter.NewRateLimiter(1000, time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.AllowRequest()
	}
}

func BenchmarkConcurrentAllowRequest(b *testing.B) {
	rl := ratelimiter.NewRateLimiter(1000, time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rl.AllowRequest()
		}
	})
}

func BenchmarkCurrentState(b *testing.B) {
	rl := ratelimiter.NewRateLimiter(100, time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.CurrentState()
	}
}
