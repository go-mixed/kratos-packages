package base

import (
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"time"
)

type FilterConnectionFunc = utils.MapFilterFunc[ConnectionID, IConnection]

func FilterVersion(version string) FilterConnectionFunc {
	return func(id ConnectionID, conn IConnection) bool {
		v, ok := conn.GetMetadata("version")
		return ok && v != version
	}
}

func FilterHasVersion() FilterConnectionFunc {
	return func(id ConnectionID, conn IConnection) bool {
		_, ok := conn.GetMetadata("version")
		return ok
	}
}

// FilterMatchAll 在matches中匹配所有的key/value，如果都匹配上了，返回true
func FilterMatchAll(matches map[string]any) FilterConnectionFunc {
	return func(id ConnectionID, conn IConnection) bool {
		for key, val := range matches {
			val2, ok := conn.GetMetadata(key)
			if !ok || val != val2 {
				return false
			}
		}
		return true
	}
}

// FilterBefore 返回所有最近一次更新时间在beforeAt之前（不含）的连接
func FilterBefore(beforeAt time.Time) FilterConnectionFunc {
	return func(id ConnectionID, conn IConnection) bool {
		return conn.GetLastRecvAt().Before(beforeAt)
	}
}

// FilterAfter 返回所有最近一次更新时间在afterAt之后（不含）的连接
func FilterAfter(afterAt time.Time) FilterConnectionFunc {
	return func(id ConnectionID, conn IConnection) bool {
		return conn.GetLastRecvAt().After(afterAt)
	}
}
