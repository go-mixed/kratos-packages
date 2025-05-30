package connection

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/cache"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/redis"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
	"hash/fnv"
	"iter"
	"time"
)

type ConnectionManager struct {
	cache         *cache.Cache
	logger        *log.Helper
	onlineTimeout time.Duration

	connections utils.ConcurrentMap[base.ConnectionID, base.IConnection]
}

var _ base.IConnections = (*ConnectionManager)(nil)

const defaultConnectionZSetCount = 4

func NewConnectionManager(
	cache *cache.Cache,
	logger *log.Helper,

	onlineTimeout time.Duration,
) *ConnectionManager {
	return &ConnectionManager{
		cache:  cache,
		logger: logger,

		onlineTimeout: onlineTimeout,
		connections:   utils.ConcurrentMap[base.ConnectionID, base.IConnection]{},
	}
}

func (h *ConnectionManager) cacheKey(id base.ConnectionID) string {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(id))
	index := hash.Sum32()
	return h._cacheKey(index)
}

func (h *ConnectionManager) _cacheKey(index uint32) string {
	return fmt.Sprintf("ws:connections:%d", index%defaultConnectionZSetCount)

}

// Upsert 加入/修改连接
func (h *ConnectionManager) Upsert(conn base.IConnection) {
	h.connections.Store(conn.GetID(), conn)
	h.TouchRedis(conn)
}

// CompareAndDelete 对比conn之后，再删除。（如果有同名ID的conn连接，之前的连接会被挤掉，但是此时该id不应该被删除）
func (h *ConnectionManager) CompareAndDelete(id base.ConnectionID, conn base.IConnection) bool {
	deleted := h.connections.CompareAndDelete(id, conn)
	if deleted {
		h.RemoveFromRedis(conn)
	}
	return deleted
}

// TouchRedis 更新redis中的连接记录
func (h *ConnectionManager) TouchRedis(conn base.IConnection) {
	cacheKey := h.cacheKey(conn.GetID())

	_, _ = h.cache.ZAdd(conn.Context(), cacheKey, redis.Z{Member: conn.GetID(), Score: float64(time.Now().Unix())})
	_, _ = h.cache.Expire(conn.Context(), cacheKey, h.onlineTimeout)
}

// RemoveFromRedis 移除redis中的连接记录
func (h *ConnectionManager) RemoveFromRedis(conn base.IConnection) {
	cacheKey := h.cacheKey(conn.GetID())

	_, _ = h.cache.ZRem(conn.Context(), cacheKey, conn.GetID())
	_, _ = h.cache.Expire(conn.Context(), cacheKey, h.onlineTimeout)
}

func (h *ConnectionManager) GetAllConnectionIDs(ctx context.Context) []base.ConnectionID {
	after := cast.ToString(time.Now().Add(-h.onlineTimeout).Unix())

	var connectionIds []string
	for i := uint32(0); i < defaultConnectionZSetCount; i++ {
		ids, _ := h.cache.ZRangeByScore(ctx, h._cacheKey(i), &redis.ZRangeBy{Min: after, Max: "+inf"})
		connectionIds = append(connectionIds, ids...)
	}

	return lo.Map(connectionIds, func(item string, index int) base.ConnectionID {
		return base.ConnectionID(item)
	})
}

// ----- 基础调用 ------

func (h *ConnectionManager) Len() int {
	return h.connections.Len()
}

func (h *ConnectionManager) Get(connectionID base.ConnectionID) base.IConnection {
	conn, _ := h.connections.Load(connectionID)
	return conn
}

func (h *ConnectionManager) MGet(connectionIDS ...base.ConnectionID) base.IConnections {
	var conns base.Connections = base.Connections{}
	for _, id := range connectionIDS {
		conn, ok := h.connections.Load(id)
		if ok {
			conns[conn.GetID()] = conn
		}
	}

	return conns
}

func (h *ConnectionManager) HasKey(connectionID base.ConnectionID) bool {
	_, ok := h.connections.Load(connectionID)
	return ok
}

func (h *ConnectionManager) Iterator(fns ...base.FilterConnectionFunc) iter.Seq2[base.ConnectionID, base.IConnection] {
	return h.connections.Iterator(fns...)
}

func (h *ConnectionManager) IDs() []base.ConnectionID {
	var keys []base.ConnectionID
	for k := range h.connections.Iterator() {
		keys = append(keys, k)
	}

	return keys
}
