package base

import (
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"iter"
)

type Connections map[ConnectionID]IConnection

var _ IConnections = Connections{}

func (s Connections) Len() int {
	return len(s)
}

func (s Connections) Get(sessionID ConnectionID) IConnection {
	return s[sessionID]
}

func (s Connections) MGet(sessionIDs ...ConnectionID) IConnections {
	var newSessions = make(Connections)
	for _, sessionID := range sessionIDs {
		if session, ok := s[sessionID]; ok {
			newSessions[sessionID] = session
		}
	}

	return newSessions
}

func (s Connections) HasKey(sessionID ConnectionID) bool {
	_, ok := s[sessionID]
	return ok
}

// Iterator 迭代器，可以使用for := range 遍历所有session。并且可以传入多个FilterFunc来筛选session
func (s Connections) Iterator(fns ...FilterConnectionFunc) iter.Seq2[ConnectionID, IConnection] {
	return utils.MapIterator[ConnectionID, IConnection](s, utils.WrapMapFilterFunc(fns...))
}

func (s Connections) IDs() []ConnectionID {
	return lo.Keys(s)
}
