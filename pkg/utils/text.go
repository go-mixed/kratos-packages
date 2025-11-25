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
// options: Unicode 比较选项，如果为空则使用严格匹配
func maxMatchForward(text string, terms []string, options ...UnicodeNormalizeOption) map[string]bool {
	text = UnicodeCanonical(text, options...)
	extracted := make(map[string]bool)
	textLen := len(text)

	// 批量规范化所有术语（性能优化）
	canonicalTerms := lo.Map(terms, func(term string, _ int) string {
		return UnicodeCanonical(term, options...)
	})
	// 提取所有唯一的术语长度（按字节计算），并降序排序
	uniqueLengths := buildUniqueLengths(canonicalTerms)

	// 构建规范化键到原始术语的映射
	termSet := make(map[string]string)
	for i, newTerm := range canonicalTerms {
		termSet[newTerm] = terms[i]
	}

	i := 0

	// 统一的匹配逻辑
	for i < textLen {
		matched := false
		// 按原始术语长度从大到小尝试匹配
		for _, length := range uniqueLengths {
			if i+length > textLen {
				continue
			}

			candidate := text[i : i+length]

			if originalTerm, ok := termSet[candidate]; ok {
				extracted[originalTerm] = true
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
// options: Unicode 比较选项，如果为空则使用严格匹配
func maxMatchBackward(text string, terms []string, options ...UnicodeNormalizeOption) map[string]bool {
	text = UnicodeCanonical(text, options...)
	extracted := make(map[string]bool)
	textLen := len(text)

	// 批量规范化所有术语
	canonicalTerms := lo.Map(terms, func(term string, _ int) string {
		return UnicodeCanonical(term, options...)
	})
	// 提取所有唯一的术语长度（按字节计算），并降序排序
	uniqueLengths := buildUniqueLengths(canonicalTerms)

	// 构建规范化键到原始术语的映射
	termSet := make(map[string]string)
	for i, newTerm := range canonicalTerms {
		termSet[newTerm] = terms[i]
	}

	i := textLen

	// 统一的匹配逻辑
	for i > 0 {
		matched := false
		// 按原始术语长度从大到小尝试匹配
		for _, length := range uniqueLengths {
			if i-length < 0 {
				continue
			}

			candidate := text[i-length : i]

			if originalTerm, ok := termSet[candidate]; ok {
				extracted[originalTerm] = true
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

// MaxMatchExtract 使用最大匹配原则从文本中提取匹配的术语（正向扫描）
// text: 待匹配的文本
// terms: 术语列表
// options: Unicode 比较选项（可选）
//   - 无选项: 严格字节匹配（高性能）
//   - IgnoreCase: 忽略大小写（使用 Unicode case folding 标准）
//     例: "Hello" 匹配 "hello", "Straße" 匹配 "strasse"（ß→ss）
//   - IgnoreDiacritics: 忽略变音符号（建议与 IgnoreCase 组合）
//     例: "café" 匹配 "cafe", "naïve" 匹配 "naive"
//   - IgnoreWidth: 忽略全角半角（建议与 IgnoreCase 组合）
//     例: "Ｈｅｌｌｏ" 匹配 "Hello"
//   - Loose: 宽松比较 (= IgnoreCase | IgnoreDiacritics | IgnoreWidth)
//
// 返回：匹配到的术语集合（无序）
//
// 示例:
//
//	MaxMatchExtract("Hello World", []string{"hello", "world"})                          // [] (严格匹配)
//	MaxMatchExtract("Hello World", []string{"hello", "world"}, IgnoreCase)              // ["hello", "world"]
//	MaxMatchExtract("café naïve", []string{"cafe", "naive"}, IgnoreCase, IgnoreDiacritics) // ["cafe", "naive"]
//	MaxMatchExtract("Ｈｅｌｌｏ", []string{"Hello"}, Loose)                              // ["Hello"]
func MaxMatchExtract(text string, terms []string, options ...UnicodeNormalizeOption) []string {
	if len(text) == 0 || len(terms) == 0 {
		return nil
	}

	// 正向最大匹配（复用核心算法）
	extracted := maxMatchForward(text, terms, options...)

	// 转换为切片返回
	if len(extracted) == 0 {
		return nil
	}
	return lo.Keys(extracted)
}

// MaxMatchReplace 使用最大匹配原则替换文本中的术语（仅支持严格字节匹配）
// text: 待替换的文本
// replaceMap: 术语替换映射 (原术语 -> 替换后的文本)
//
// 返回：替换后的文本
//
// 示例:
//
//	MaxMatchReplace("中国人民银行", map[string]string{
//	    "中国":   "A",
//	    "中国人": "B",
//	    "银行":   "D",
//	})  // "B民D" (最大匹配: "中国人"→B, "民"保留, "银行"→D)
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
// options: Unicode 比较选项（可选）
//   - 无选项: 严格字节匹配（高性能）
//   - IgnoreCase: 忽略大小写（使用 Unicode case folding 标准）
//     例: "Hello" 匹配 "hello", "Straße" 匹配 "strasse"（ß→ss）
//   - IgnoreDiacritics: 忽略变音符号（建议与 IgnoreCase 组合）
//     例: "café" 匹配 "cafe", "naïve" 匹配 "naive"
//   - IgnoreWidth: 忽略全角半角（建议与 IgnoreCase 组合）
//     例: "Ｈｅｌｌｏ" 匹配 "Hello"
//   - Loose: 宽松比较 (= IgnoreCase | IgnoreDiacritics | IgnoreWidth)
//
// 返回：匹配到的术语集合（无序，取正向和反向匹配的并集）
//
// 双向匹配说明：
// 正向扫描可能产生分词歧义，反向扫描可以发现更多匹配，取并集提高召回率。
// 例: "研究生命起源" + terms["研究生", "生命", "起源"]
//   - 正向: ["研究生", "起源"] (生命被跳过)
//   - 反向: ["生命", "起源"] (研究生被拆分)
//   - 并集: ["研究生", "生命", "起源"]
//
// 示例:
//
//	BiMaxMatchExtract("Hello World", []string{"hello", "world"}, IgnoreCase)              // ["hello", "world"]
//	BiMaxMatchExtract("café naïve", []string{"cafe", "naive"}, Loose)                     // ["cafe", "naive"]
//	BiMaxMatchExtract("研究生命起源", []string{"研究生", "生命", "起源"})                   // ["研究生", "生命", "起源"]
func BiMaxMatchExtract(text string, terms []string, options ...UnicodeNormalizeOption) []string {
	if len(text) == 0 || len(terms) == 0 {
		return nil
	}

	// 正向最大匹配（复用核心算法）
	forwardMatched := maxMatchForward(text, terms, options...)

	// 反向最大匹配（复用核心算法）
	backwardMatched := maxMatchBackward(text, terms, options...)

	// 合并结果（取并集）
	result := lo.Uniq(append(lo.Keys(forwardMatched), lo.Keys(backwardMatched)...))

	return result
}
