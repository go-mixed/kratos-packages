package limit

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	xrate "golang.org/x/time/rate"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	tokenFormat     = "throttle:{%s}.tokens"
	timestampFormat = "throttle:{%s}.ts"
	pingInterval    = time.Millisecond * 100
)

// to be compatible with aliyun redis, we cannot use `local key = KEYS[1]` to reuse the key
// KEYS[1] as tokens_key
// KEYS[2] as timestamp_key
var script = redis.NewScript(`local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local fill_time = capacity/rate
local ttl = math.floor(fill_time*2)
local last_tokens = tonumber(redis.call("get", KEYS[1]))
if last_tokens == nil then
    last_tokens = capacity
end

local last_refreshed = tonumber(redis.call("get", KEYS[2]))
if last_refreshed == nil then
    last_refreshed = 0
end

local delta = math.max(0, now-last_refreshed)
local filled_tokens = math.min(capacity, last_tokens+(delta*rate))
local allowed = filled_tokens >= requested
local new_tokens = filled_tokens
if allowed then
    new_tokens = filled_tokens - requested
end

redis.call("setex", KEYS[1], ttl, new_tokens)
redis.call("setex", KEYS[2], ttl, now)

return allowed`)

// A TokenLimiter controls how frequently events are allowed to happen with in one second.
type TokenLimiter struct {
	rate           int
	burst          int
	store          *redis.Client
	tokenKey       string
	timestampKey   string
	rescueLock     sync.Mutex
	redisAlive     uint32
	monitorStarted bool
	rescueLimiter  *xrate.Limiter

	logger *log.Helper
}

// NewTokenLimiter returns a new TokenLimiter that allows events up to rate and permits
// bursts of at most burst tokens.
//   - rate: the rate to be generated in tokens per second.
//     令牌桶的速率，表示每秒生成多少个令牌。
//   - burst: the maximum number of tokens that can be consumed in a single burst.
//     令牌桶的总量，表示最多可以积累多少个令牌。
//
// eg: rate = 5, burst = 20,
// It means that 5 tokens can be generated per second, and up to 20 tokens can be accumulated.
// If you don't use it for 1st second, you can request 10 tokens in the second.
// If you don't use it for the first 3 seconds, you can request 20 tokens in the fourth second.
// 表示每秒生成5个令牌，如果第1秒没有使用，那么第二秒就可以使用10个令牌。前3秒都没使用，则第4秒可以使用20个令牌。
//
// eg: rate = 6, burst = 6,
// It can request 6 tokens per second if the values of two parameters are the same 6
// 两个值相同，表示每秒只能使用6个令牌。
func NewTokenLimiter(
	rate, burst int,
	store *redis.Client,
	key string,
	logger log.Logger,
) *TokenLimiter {
	tokenKey := fmt.Sprintf(tokenFormat, key)
	timestampKey := fmt.Sprintf(timestampFormat, key)

	return &TokenLimiter{
		rate:          rate,
		burst:         burst,
		store:         store,
		tokenKey:      tokenKey,
		timestampKey:  timestampKey,
		redisAlive:    1,
		rescueLimiter: xrate.NewLimiter(xrate.Every(time.Second/time.Duration(rate)), burst),

		logger: log.NewModuleHelper(logger, "pkg/limit"),
	}
}

// Allow is shorthand for AllowN(time.Now(), 1).
func (lim *TokenLimiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

// AllowCtx is shorthand for AllowNCtx(ctx,time.Now(), 1) with incoming context.
func (lim *TokenLimiter) AllowCtx(ctx context.Context) bool {
	return lim.AllowNCtx(ctx, time.Now(), 1)
}

// AllowN reports whether n events may happen at time now.
// Use this method if you intend to drop / skip events that exceed the rate.
// Otherwise, use Reserve or Wait.
func (lim *TokenLimiter) AllowN(now time.Time, n int) bool {
	return lim.reserveN(context.Background(), now, n)
}

// AllowNCtx reports whether n events may happen at time now with incoming context.
// Use this method if you intend to drop / skip events that exceed the rate.
// Otherwise, use Reserve or Wait.
func (lim *TokenLimiter) AllowNCtx(ctx context.Context, now time.Time, n int) bool {
	return lim.reserveN(ctx, now, n)
}

func (lim *TokenLimiter) reserveN(ctx context.Context, now time.Time, n int) bool {
	if atomic.LoadUint32(&lim.redisAlive) == 0 {
		return lim.rescueLimiter.AllowN(now, n)
	}

	resp, err := script.Run(ctx, lim.store, []string{
		lim.tokenKey,
		lim.timestampKey,
	}, []string{
		strconv.Itoa(lim.rate),
		strconv.Itoa(lim.burst),
		strconv.FormatInt(now.Unix(), 10),
		strconv.Itoa(n),
	}).Result()

	// redis allowed == false
	// Lua boolean false -> r Nil bulk reply
	if err == redis.Nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		lim.logger.Errorf("fail to use rate limiter: %s", err)
		return false
	}
	if err != nil {
		lim.logger.Errorf("fail to use rate limiter: %s, use in-process limiter for rescue", err)
		lim.startMonitor()
		return lim.rescueLimiter.AllowN(now, n)
	}

	code, ok := resp.(int64)
	if !ok {
		lim.logger.Errorf("fail to eval redis script: %v, use in-process limiter for rescue", resp)
		lim.startMonitor()
		return lim.rescueLimiter.AllowN(now, n)
	}

	// redis allowed == true
	// Lua boolean true -> r integer reply with value of 1
	return code == 1
}

func (lim *TokenLimiter) startMonitor() {
	lim.rescueLock.Lock()
	defer lim.rescueLock.Unlock()

	if lim.monitorStarted {
		return
	}

	lim.monitorStarted = true
	atomic.StoreUint32(&lim.redisAlive, 0)

	go lim.waitForRedis()
}

func (lim *TokenLimiter) waitForRedis() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		lim.rescueLock.Lock()
		lim.monitorStarted = false
		lim.rescueLock.Unlock()
	}()

	for range ticker.C {
		if lim.store.Ping(context.Background()).Val() == "PONG" {
			atomic.StoreUint32(&lim.redisAlive, 1)
			return
		}
	}
}
