package ratelimiter_test

import (
	"fmt"
	"time"

	"github.com/TRAD3R/ratelimiter"
)

func ExampleNewRateLimiter() {
	// Создаем rate limiter с лимитом 10 запросов в секунду
	rl := ratelimiter.NewRateLimiter(10, time.Second)

	// Проверяем, разрешен ли запрос
	if rl.AllowRequest() {
		fmt.Println("Request allowed")
	} else {
		fmt.Println("Request denied")
	}
	// Output: Request allowed
}

func ExampleRateLimiter_AllowRequest() {
	// Создаем rate limiter с лимитом 2 запроса в 100 миллисекунд
	rl := ratelimiter.NewRateLimiter(2, 100*time.Millisecond)

	// Первые два запроса должны быть разрешены
	for i := 0; i < 2; i++ {
		if rl.AllowRequest() {
			fmt.Printf("Request %d: allowed\n", i+1)
		}
	}

	// Третий запрос должен быть отклонен
	if !rl.AllowRequest() {
		fmt.Println("Request 3: denied (rate limit exceeded)")
	}

	// Ждем сброса лимита
	time.Sleep(110 * time.Millisecond)

	// После сброса запрос снова разрешен
	if rl.AllowRequest() {
		fmt.Println("Request after reset: allowed")
	}
	// Output:
	// Request 1: allowed
	// Request 2: allowed
	// Request 3: denied (rate limit exceeded)
	// Request after reset: allowed
}

func ExampleRateLimiter_CurrentState() {
	rl := ratelimiter.NewRateLimiter(5, time.Second)

	// Делаем несколько запросов
	rl.AllowRequest()
	rl.AllowRequest()

	// Проверяем текущее состояние
	requests, timeUntilReset := rl.CurrentState()
	fmt.Printf("Current requests: %d\n", requests)
	if timeUntilReset > 0 {
		fmt.Println("Rate limiter is active")
	}
	// Output:
	// Current requests: 2
	// Rate limiter is active
}
