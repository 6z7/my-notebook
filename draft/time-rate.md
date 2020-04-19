golang.org/x/time/rate 令牌桶实现


```go
// Each of the three methods consumes a single token.
// They differ in their behavior when no token is available.
// If no token is available, Allow returns false.
// If no token is available, Reserve returns a reservation for a future token
// and the amount of time the caller must wait before using it.
// If no token is available, Wait blocks until one can be obtained
// or its associated context.Context is canceled.
//
// The methods AllowN, ReserveN, and WaitN consume n tokens.
```

```go
type Limiter struct {
    // 每秒生成多少个token
	limit Limit
    // token桶的容量，最大允许的突增并发
	burst int
    // 获取token时需要获取锁
	mu     sync.Mutex
    // 当前token数量
	tokens float64
	// 上次计算token的时间
	last time.Time
	// lastEvent is the latest time of a rate-limited event (past or future)
	lastEvent time.Time
}
```
## 根据时间计算能够获取的token数量

```go
// 每秒生成N个token，根据token数量推算出需要的时间
func (limit Limit) durationFromTokens(tokens float64) time.Duration {
	seconds := tokens / float64(limit)
	return time.Nanosecond * time.Duration(1e9*seconds)
}

// 计算指定时间能够生成多少个token
func (limit Limit) tokensFromDuration(d time.Duration) float64 {
	return d.Seconds() * float64(limit)
}

// 到达指定时间生成的token数量
func (lim *Limiter) advance(now time.Time) (newNow time.Time, newLast time.Time, newTokens float64) {
	last := lim.last   
	if now.Before(last) {
		last = now
	}

	// Avoid making delta overflow below when last is very old.
	// 计算需要多久token桶满装满，防止桶溢出
	// 第一次初始化时桶被填满
	maxElapsed := lim.limit.durationFromTokens(float64(lim.burst) - lim.tokens)
	elapsed := now.Sub(last)
	if elapsed > maxElapsed {
		elapsed = maxElapsed
	}

	// Calculate the new number of tokens, due to time that passed.
	// 过去的这段时间生成了多少token
	delta := lim.limit.tokensFromDuration(elapsed)
	// 桶中可以使用的token数量
	tokens := lim.tokens + delta
	// 不能超过桶容量
	if burst := float64(lim.burst); tokens > burst {
		tokens = burst
	}
    
	return now, last, tokens
}
```
## 预约n个token

```go
type Reservation struct {
	// 获取token是否成功
	ok        bool
	// 
	lim       *Limiter
	// 当前桶中的token
	tokens    int
	// 到达指定时间才能获取到指定数量的token
	timeToAct time.Time
	// This is the Limit at reservation time, it can change later.
	limit Limit
}

// 尝试获取n个token
func (lim *Limiter) reserveN(now time.Time, n int, maxFutureReserve time.Duration) Reservation {
	lim.mu.Lock()
    // 没有频率限制
	if lim.limit == Inf {
		lim.mu.Unlock()
		return Reservation{
			ok:        true,
			lim:       lim,
			tokens:    n,
			timeToAct: now,
		}
	}
    // 当前能够获取的token数量，不会超过设置的最大值
	now, last, tokens := lim.advance(now)

	// Calculate the remaining number of tokens resulting from the request.
	tokens -= float64(n)

	// Calculate the wait duration
	var waitDuration time.Duration
	// 桶中token不足则计算需要等待的时间
	if tokens < 0 {
		waitDuration = lim.limit.durationFromTokens(-tokens)
	}

	// Decide result	
	ok := n <= lim.burst && waitDuration <= maxFutureReserve

	// Prepare reservation
	r := Reservation{
		ok:    ok,
		lim:   lim,
		limit: lim.limit,
	}
	if ok {
		r.tokens = n
		r.timeToAct = now.Add(waitDuration)
	}

	// Update state
	if ok {
		// 更新限流器状态
		lim.last = now
		lim.tokens = tokens
		lim.lastEvent = r.timeToAct
	} else {
		lim.last = last
	}

	lim.mu.Unlock()
	return r
}
```