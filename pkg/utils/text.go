package utils

import (
	"math/rand"
	"slices"

	"github.com/samber/lo"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// WildcardMatchSimple - finds whether the text matches/satisfies the pattern string.
// supports only '*' wildcard in the pattern.
// considers a file system path as a flat name space.
func WildcardMatchSimple(pattern, name string) bool {
	if pattern == "" {
		return name == pattern
	}
	if pattern == "*" {
		return true
	}
	// Does only wildcard '*' match.
	return deepWildcardMatchRune([]rune(name), []rune(pattern), true)
}

// WildcardMatch -  finds whether the text matches/satisfies the pattern string.
// supports  '*' and '?' wildcards in the pattern string.
// unlike path.Match(), considers a path as a flat name space while matching the pattern.
// The difference is illustrated in the example here https://play.golang.org/p/Ega9qgD4Qz .
func WildcardMatch(pattern, name string) (matched bool) {
	if pattern == "" {
		return name == pattern
	}
	if pattern == "*" {
		return true
	}
	// Does extend wildcard '*' and '?' match.
	return deepWildcardMatchRune([]rune(name), []rune(pattern), false)
}

func deepWildcardMatchRune(str, pattern []rune, simple bool) bool {
	for len(pattern) > 0 {
		switch pattern[0] {
		default:
			if len(str) == 0 || str[0] != pattern[0] {
				return false
			}
		case '?':
			if len(str) == 0 && !simple {
				return false
			}
		case '*':
			return deepWildcardMatchRune(str, pattern[1:], simple) || // 当前str[0]的字符是否匹配*之后(pattern[1])的字符
				(len(str) > 0 && deepWildcardMatchRune(str[1:], pattern, simple)) // 上面没匹配到, 则继续用*匹配下一个(str[1])字符
		}
		str = str[1:]
		pattern = pattern[1:]
	}
	return len(str) == 0 && len(pattern) == 0
}

// buildUniqueLengths 构建唯一的术语长度数组（按字节计算），并降序排序
func buildUniqueLengths(terms []string) []int {
	uniqueLengths := lo.Uniq(lo.Map(terms, func(term string, _ int) int {
		return len(term)
	}))
	slices.SortFunc(uniqueLengths, func(a, b int) int {
		return b - a // 降序
	})
	return uniqueLengths
}

// maxMatchForward 正向最大匹配（核心算法）
func maxMatchForward(text string, termSet map[string]bool, uniqueLengths []int) map[string]bool {
	extracted := make(map[string]bool)
	textLen := len(text)
	i := 0

	for i < textLen {
		matched := false
		for _, length := range uniqueLengths {
			if i+length > textLen {
				continue
			}

			candidate := text[i : i+length]
			if termSet[candidate] {
				extracted[candidate] = true
				i += length
				matched = true
				break
			}
		}

		if !matched {
			i++
		}
	}

	return extracted
}

// maxMatchBackward 反向最大匹配
func maxMatchBackward(text string, termSet map[string]bool, uniqueLengths []int) map[string]bool {
	extracted := make(map[string]bool)
	textLen := len(text)
	i := textLen

	for i > 0 {
		matched := false
		for _, length := range uniqueLengths {
			if i-length < 0 {
				continue
			}

			candidate := text[i-length : i]
			if termSet[candidate] {
				extracted[candidate] = true
				i -= length
				matched = true
				break
			}
		}

		if !matched {
			i--
		}
	}

	return extracted
}

// MaxMatchExtract 使用最大匹配原则从文本中提取匹配的术语
// text: 待匹配的文本
// terms: 术语列表
// 返回：匹配到的术语集合
func MaxMatchExtract(text string, terms []string) []string {
	if len(text) == 0 || len(terms) == 0 {
		return nil
	}

	// 构建术语集合用于快速查找
	termSet := lo.SliceToMap(terms, func(term string) (string, bool) {
		return term, true
	})

	// 提取所有唯一的术语长度（按字节计算），并降序排序
	uniqueLengths := buildUniqueLengths(terms)

	// 正向最大匹配（复用核心算法）
	extracted := maxMatchForward(text, termSet, uniqueLengths)

	// 转换为切片返回
	return lo.Keys(extracted)
}

// MaxMatchReplace 使用最大匹配原则替换文本中的术语
// text: 待替换的文本
// replaceMap: 术语替换映射 (原术语 -> 替换后的文本)
// 返回：替换后的文本
func MaxMatchReplace(text string, replaceMap map[string]string) string {
	if len(text) == 0 || len(replaceMap) == 0 {
		return text
	}

	// 提取所有唯一的术语长度（按字节计算），并降序排序
	keys := lo.Keys(replaceMap)
	uniqueLengths := buildUniqueLengths(keys)

	// 获取最小长度用于边界检查
	minLen := uniqueLengths[len(uniqueLengths)-1]

	textLen := len(text)
	pos := 0
	result := make([]byte, 0, textLen)

	// 最大匹配替换（与maxMatchForward逻辑相同，但是执行替换操作）
	for pos < textLen {
		// 在每次循环中调整 maxLen（避免越界）
		currentMaxLen := uniqueLengths[0]
		if pos+currentMaxLen > textLen {
			currentMaxLen = textLen - pos
		}

		found := false

		// 按降序遍历唯一长度数组
		for _, length := range uniqueLengths {
			if length > currentMaxLen || length < minLen {
				continue
			}
			if pos+length > textLen {
				continue
			}

			candidate := text[pos : pos+length]
			if replacement, exists := replaceMap[candidate]; exists {
				// 匹配成功，追加替换后的文本
				result = append(result, replacement...)
				pos += length
				found = true
				break
			}
		}

		// 没有匹配到，保留当前字节
		if !found {
			result = append(result, text[pos])
			pos++
		}
	}

	return string(result)
}

// BiMaxMatchExtract 使用双向最大匹配原则从文本中提取匹配的术语（更精确）
// text: 待匹配的文本
// terms: 术语列表
// 返回：匹配到的术语集合
func BiMaxMatchExtract(text string, terms []string) []string {
	if len(text) == 0 || len(terms) == 0 {
		return nil
	}

	// 构建术语集合用于快速查找
	termSet := lo.SliceToMap(terms, func(term string) (string, bool) {
		return term, true
	})

	// 提取所有唯一的术语长度（按字节计算），并降序排序
	uniqueLengths := buildUniqueLengths(terms)

	// 正向最大匹配（复用核心算法）
	forwardMatched := maxMatchForward(text, termSet, uniqueLengths)

	// 反向最大匹配（复用核心算法）
	backwardMatched := maxMatchBackward(text, termSet, uniqueLengths)

	// 合并结果（取并集）
	result := lo.Uniq(append(lo.Keys(forwardMatched), lo.Keys(backwardMatched)...))

	return result
}
