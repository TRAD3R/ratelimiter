# Rate Limiter

Простой и эффективный rate limiter для Go, который позволяет ограничивать количество запросов в единицу времени.

## Особенности

- 🚀 **Высокая производительность** - использует минимальные блокировки
- 🔒 **Thread-safe** - безопасен для использования в многопоточной среде
- ⚡ **Простой API** - всего несколько методов для работы
- 🎯 **Точное ограничение** - строго соблюдает установленные лимиты
- 📊 **Мониторинг состояния** - возможность получить текущее состояние лимитера

## Установка

```bash
go get github.com/TRAD3R/ratelimiter
```

## Быстрый старт

```go
package main

import (
    "fmt"
    "time"
    "github.com/TRAD3R/ratelimiter"
)

func main() {
    // Создаем rate limiter: 10 запросов в секунду
    rl := ratelimiter.NewRateLimiter(10, time.Second)
    
    // Проверяем, разрешен ли запрос
    if rl.AllowRequest() {
        fmt.Println("Запрос разрешен")
    } else {
        fmt.Println("Запрос отклонен - превышен лимит")
    }
}
```

## API

### NewRateLimiter(rps int, interval time.Duration) *RateLimiter

Создает новый rate limiter с указанным лимитом запросов в секунду и интервалом сброса.

**Параметры:**
- `rps` - максимальное количество запросов за интервал
- `interval` - интервал времени для сброса счетчика

**Пример:**
```go
// 100 запросов в секунду
rl := ratelimiter.NewRateLimiter(100, time.Second)

// 5 запросов в минуту
rl := ratelimiter.NewRateLimiter(5, time.Minute)

// 1000 запросов в 100 миллисекунд
rl := ratelimiter.NewRateLimiter(1000, 100*time.Millisecond)
```

### AllowRequest() bool

Проверяет, разрешен ли запрос согласно установленным лимитам. Если запрос разрешен, увеличивает внутренний счетчик.

**Возвращает:**
- `true` - запрос разрешен
- `false` - запрос отклонен (превышен лимит)

**Пример:**
```go
rl := ratelimiter.NewRateLimiter(2, time.Second)

// Первые два запроса будут разрешены
for i := 0; i < 3; i++ {
    if rl.AllowRequest() {
        fmt.Printf("Запрос %d: разрешен\n", i+1)
    } else {
        fmt.Printf("Запрос %d: отклонен\n", i+1)
    }
}
// Вывод:
// Запрос 1: разрешен
// Запрос 2: разрешен
// Запрос 3: отклонен
```

### CurrentState() (int, time.Duration)

Возвращает текущее состояние rate limiter'а.

**Возвращает:**
- `int` - количество уже использованных запросов в текущем интервале
- `time.Duration` - время до сброса счетчика

**Пример:**
```go
rl := ratelimiter.NewRateLimiter(10, time.Second)

// Делаем несколько запросов
rl.AllowRequest()
rl.AllowRequest()

// Проверяем состояние
requests, timeUntilReset := rl.CurrentState()
fmt.Printf("Использовано запросов: %d, до сброса: %v\n", requests, timeUntilReset)
```

## Примеры использования

### Базовое использование

```go
package main

import (
    "fmt"
    "time"
    "github.com/TRAD3R/ratelimiter"
)

func main() {
    rl := ratelimiter.NewRateLimiter(5, time.Second)
    
    for i := 0; i < 10; i++ {
        if rl.AllowRequest() {
            fmt.Printf("Запрос %d: OK\n", i+1)
        } else {
            fmt.Printf("Запрос %d: RATE LIMITED\n", i+1)
        }
        time.Sleep(100 * time.Millisecond)
    }
}
```

### HTTP middleware

```go
package main

import (
    "net/http"
    "time"
    "github.com/TRAD3R/ratelimiter"
)

var rl = ratelimiter.NewRateLimiter(100, time.Minute) // 100 запросов в минуту

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !rl.AllowRequest() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })
    
    http.ListenAndServe(":8080", rateLimitMiddleware(mux))
}
```

### Мониторинг состояния

```go
package main

import (
    "fmt"
    "time"
    "github.com/TRAD3R/ratelimiter"
)

func main() {
    rl := ratelimiter.NewRateLimiter(3, 2*time.Second)
    
    // Мониторим состояние каждые 500мс
    go func() {
        ticker := time.NewTicker(500 * time.Millisecond)
        defer ticker.Stop()
        
        for range ticker.C {
            requests, timeLeft := rl.CurrentState()
            fmt.Printf("Состояние: %d/%d запросов, до сброса: %v\n", 
                requests, 3, timeLeft.Round(time.Millisecond))
        }
    }()
    
    // Делаем запросы
    for i := 0; i < 10; i++ {
        rl.AllowRequest()
        time.Sleep(300 * time.Millisecond)
    }
    
    time.Sleep(3 * time.Second)
}
```

## Производительность

Rate limiter оптимизирован для высокой производительности:

- **Блокировки**: Использует минимальные блокировки только при необходимости
- **Память**: Низкое потребление памяти (всего несколько полей)
- **CPU**: Минимальные вычисления при каждом запросе

### Бенчмарки

```
BenchmarkAllowRequest-22              26956904    43.04 ns/op
BenchmarkConcurrentAllowRequest-22     9799383   111.5 ns/op
BenchmarkCurrentState-22              44034135    27.94 ns/op
```

## Тестирование

Проект имеет 100% покрытие тестами:

```bash
# Запуск всех тестов
go test -v

# Запуск с покрытием
go test -cover

# Запуск бенчмарков
go test -bench=.

# Запуск примеров
go test -run Example
```

### Типы тестов

- ✅ **Unit тесты** - тестирование отдельных функций
- ✅ **Интеграционные тесты** - тестирование взаимодействия компонентов
- ✅ **Concurrency тесты** - проверка thread-safety
- ✅ **Edge case тесты** - граничные случаи
- ✅ **Performance тесты** - бенчмарки

## Лицензия

Apache License 2.0

## Вклад в проект

Приветствуются pull request'ы и issue'ы! Пожалуйста, убедитесь, что:

1. Код соответствует стилю проекта
2. Все тесты проходят
3. Добавлены тесты для новой функциональности
4. Обновлена документация при необходимости

## Changelog

### v1.0.0
- Первоначальный релиз
- Базовая функциональность rate limiting
- Thread-safe реализация
- Полное покрытие тестами
