package websocket

type FilterFunc func(*Session) bool

func FilterVersion(version string) FilterFunc {
	return func(session *Session) bool {
		v, ok := session.Get("version")
		return ok && v != version
	}
}

func FilterHasVersion() FilterFunc {
	return func(session *Session) bool {
		_, ok := session.Get("version")
		return ok
	}
}

// FilterMatchAll 在matches中匹配所有的key/value，如果都匹配上了，返回true
func FilterMatchAll(matches map[string]any) FilterFunc {
	return func(session *Session) bool {
		for key, val := range matches {
			val2, ok := session.Get(key)
			if !ok || val != val2 {
				return false
			}
		}
		return true
	}
}
