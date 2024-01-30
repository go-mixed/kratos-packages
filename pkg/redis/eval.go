package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// Eval 执行脚本
// EVAL script numkeys key [key ...] arg [arg ...]
// https://redis.io/commands/eval
// 返回值：
// 1. 脚本执行后的返回值，或者是错误信息。
func (c *Redis) Eval(ctx context.Context, script string, keys []string, args ...any) *redis.Cmd {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).Eval(ctx, script, _keys, args...)
}

// EvalSha 执行哈希为sha1在redis缓存中的脚本。
// EVALSHA sha1 numkeys key [key ...] arg [arg ...]
// https://redis.io/commands/evalsha
// 返回值：
// 1. 脚本执行后的返回值，或者是错误信息。
// 2. 如果脚本不存在于缓存当中，返回错误。
func (c *Redis) EvalSha(ctx context.Context, sha1 string, keys []string, args ...any) *redis.Cmd {
	_keys := c.formatKeys(keys)
	return c.GetRedisCmd(ctx).EvalSha(ctx, sha1, _keys, args...)
}

// ScriptExists 检查给定的sha1是否已经被保存在redis缓存当中。
// SCRIPT EXISTS sha1 [sha1 ...]
// https://redis.io/commands/script-exists
// 返回值：
// 1. 一个包含布尔值的列表，列表中的每个布尔值表示一个脚本是否已经被缓存。
func (c *Redis) ScriptExists(ctx context.Context, hashes ...string) ([]bool, error) {
	return c.GetRedisCmd(ctx).ScriptExists(ctx, hashes...).Result()
}

// ScriptFlush 清除所有 Lua 脚本缓存。
// SCRIPT FLUSH
// https://redis.io/commands/script-flush
// 返回值：
// 1. 总是返回 OK 。
func (c *Redis) ScriptFlush(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).ScriptFlush(ctx).Result()
}

// ScriptKill 杀死当前正在运行的 Lua 脚本。
// SCRIPT KILL
// https://redis.io/commands/script-kill
// 返回值：
// 1. 总是返回 OK 。
func (c *Redis) ScriptKill(ctx context.Context) (string, error) {
	return c.GetRedisCmd(ctx).ScriptKill(ctx).Result()
}

// ScriptLoad 将脚本script添加到redis缓存中，但并不立即执行这个脚本。
// SCRIPT LOAD script
// https://redis.io/commands/script-load
// 返回值：
// 1. 返回脚本的 SHA1（不论是否存在）。
func (c *Redis) ScriptLoad(ctx context.Context, script string) (string, error) {
	return c.GetRedisCmd(ctx).ScriptLoad(ctx, script).Result()
}

// Script 将脚本script添加到缓存中，后续可以用Run方法执行。（链式调用）
func (c *Redis) Script(script string) *chainScript {
	return newScript(c, script)
}
