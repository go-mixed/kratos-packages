package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// Pipelined 批量执行redis命令（不保证原子性）。
// 注意：这和MySQL的事务有很大的区别。比如MySQL可以在事务中先Select，然后根据结果来Update后续的数据。
// redis的Pipelined只能执行一系列的命令（比如批量的Set），不能在一个事务中根据结果来修改后续的数据。
//
//		 Example:
//			cmds, err1 := c.Pipelined(ctx, func(ctx context.Context) error {
//				c.Get(ctx, "key1") // 此时无法获取返回值
//				return nil
//			}, func(ctx context.Context) error {
//				c.Set(ctx, "key2", "value2")
//				return nil
//			})
//	     cmds[0].Val() // 在事务结束后可以获取到值，cmd[0]即为c.Get(ctx, "key1")的返回值
func (c *Redis) Pipelined(ctx context.Context, funcs ...func(ctx context.Context) error) ([]redis.Cmder, error) {
	return c.GetRedisCmd(ctx).Pipelined(ctx, func(pipeliner redis.Pipeliner) error {
		// 将pipeliner放入上下文中，以便getClient获取
		ctx = newContext(ctx, pipeliner)
		for _, fn := range funcs {
			if err := fn(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// TxPipelined 以事务的方式批量执行redis命令（保证原子性）。
// 注意：这和MySQL的事务有很大的区别。比如MySQL可以在事务中先Select，然后根据结果来Update后续的数据。
// redis的TxPipelined只能执行一系列的命令（比如批量的Set），不能在一个事务中根据结果来修改后续的数据。
//
//		 Example:
//			cmds, err1 := c.TxPipelined(ctx, func(ctx context.Context) error {
//				c.Get(ctx, "key1") // 此时无法获取返回值
//				return nil
//			}, func(ctx context.Context) error {
//				c.Set(ctx, "key2", "value2")
//				return nil
//			})
//	     cmds[0].Val() // 在事务结束后可以获取到值，cmd[0]即为c.Get(ctx, "key1")的返回值
func (c *Redis) TxPipelined(ctx context.Context, funcs ...func(ctx context.Context) error) ([]redis.Cmder, error) {
	return c.GetRedisCmd(ctx).TxPipelined(ctx, func(pipeliner redis.Pipeliner) error {
		// 将pipeliner放入上下文中，以便getClient获取
		ctx = newContext(ctx, pipeliner)
		for _, fn := range funcs {
			if err := fn(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

/*func (c *Redis) Pipeline(ctx context.Context) redis.Pipeliner {
	oldPipeliner := c.GetRedisCmd(ctx).Pipeline()
	return NewPipeliner(oldPipeliner, c.options)
}

func (c *Redis) TxPipeline(ctx context.Context) redis.Pipeliner {
	oldPipeliner := c.GetRedisCmd(ctx).TxPipeline()
	return NewPipeliner(oldPipeliner, c.options)
}
*/
