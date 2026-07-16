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

	// Ждем, пока накопится токен
	time.Sleep(110 * time.Millisecond)

	// После пополнения запрос снова разрешен
	if rl.AllowRequest() {
		fmt.Println("Request after refill: allowed")
	}
	// Output:
	// Request 1: allowed
	// Request 2: allowed
	// Request 3: denied (rate limit exceeded)
	// Request after refill: allowed
}

func ExampleRateLimiter_CurrentState() {
	rl := ratelimiter.NewRateLimiter(5, time.Second)

	// Делаем несколько запросов
	rl.AllowRequest()
	rl.AllowRequest()

	// Проверяем текущее состояние
	available, wait := rl.CurrentState()
	fmt.Printf("Available tokens: %d\n", available)
	if wait == 0 {
		fmt.Println("Rate limiter is not blocking")
	}
	// Output:
	// Available tokens: 3
	// Rate limiter is not blocking
}
