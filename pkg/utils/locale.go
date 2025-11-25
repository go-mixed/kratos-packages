package utils

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
)

// normalizeRegionShorthand 将常见的地区简写转换为语言标签
// 处理用户可能输入的地区简写,如: cn -> zh-CN, tw -> zh-TW, us -> en-US
// 策略: 对于常见的有歧义的2字母代码,优先作为地区代码处理
func normalizeRegionShorthand(input string) string {
	// 转小写处理
	lower := strings.ToLower(strings.TrimSpace(input))

	// 对于空输入,直接返回
	if lower == "" {
		return input
	}

	// 定义地区代码到语言标签的映射
	// 包含常见的、用户很可能输入为地区代码的情况
	regionToLangMap := map[string]string{
		// === 中文地区 (常见输入) ===
		"cn": "zh-CN", // 中国大陆
		"tw": "zh-TW", // 台湾 (tw 也是 Twi语,但用户更可能是指台湾)
		"hk": "zh-HK", // 香港
		"mo": "zh-MO", // 澳门
		"sg": "zh-SG", // 新加坡 (sg 也是 Sango语,但用户更可能是指新加坡)

		// === 英语地区 (常见输入) ===
		"us": "en-US", // 美国
		"gb": "en-GB", // 英国
		"uk": "en-GB", // 英国 (常见误用, uk 也是乌克兰语但作为地区代码更常见)
		"au": "en-AU", // 澳大利亚
		"nz": "en-NZ", // 新西兰
		"za": "en-ZA", // 南非

		// === 其他明确地区代码 ===
		"jp": "ja-JP", // 日本
		"kr": "ko-KR", // 韩国 (kr 也是 Kanuri语,但用户更可能是指韩国)
		"mx": "es-MX", // 墨西哥

		// 注意: 对于明确的语言代码 (如 zh, en, ja, ko等),不做映射
		// 对于有歧义的代码(如 ca=加拿大/加泰罗尼亚语),优先保持为语言代码
	}

	// 如果找到映射,返回语言标签
	if langTag, ok := regionToLangMap[lower]; ok {
		return langTag
	}

	// 没有找到映射,返回原始输入(可能是语言代码或其他)
	return input
}

// NormalizeRFC5646Language 将各种语言/地区输入标准化为RFC5646语言标签
// 输入可以是: 地区简写(cn/tw/us)、语言代码(zh/en)、或完整语言标签(zh-CN)
// 如果输入的语言代码没有地区信息,会自动补充常用地区
//
// 示例:
//   - zh -> zh-CN, en -> en-US, ko -> ko-KR
//   - cn -> zh-CN, tw -> zh-TW, us -> en-US
//   - zh-CN -> zh-CN (验证并规范化)
//
// 注意:
//   - 仅支持主流语言代码，不支持冷门的 ISO 639-3 语言代码
//   - 如需支持更多语言，请添加到 defaultRegions 映射表中
func NormalizeRFC5646Language(locale string) (string, error) {
	// 预处理: 将常见的地区简写转换为语言标签
	// 这样用户输入 "cn", "tw", "hk" 等也能正确处理
	locale = normalizeRegionShorthand(locale)

	tag, err := language.Parse(locale)
	if err != nil {
		return "", err
	}

	// 检查是否已有明确的地区信息
	// 只有当 confidence 为 Exact 时才认为用户明确指定了地区
	_, confidence := tag.Region()
	if confidence == language.Exact {
		// 已有明确的地区信息,返回规范化后的原始 tag
		return tag.String(), nil
	}

	// 没有明确地区信息,根据语言代码补充默认地区
	// 获取基础语言(不包含脚本和地区)
	base, _ := tag.Base()
	baseStr := base.String()

	// 定义默认地区映射 (按语言使用人口和重要性排序)
	defaultRegions := map[string]string{
		// === 主流语言 (10亿+使用者) ===
		"zh": "CN", // 简体中文 -> 中国大陆
		"en": "US", // 英语 -> 美国
		"hi": "IN", // 印地语 -> 印度
		"es": "ES", // 西班牙语 -> 西班牙
		"ar": "SA", // 阿拉伯语 -> 沙特阿拉伯
		"bn": "BD", // 孟加拉语 -> 孟加拉国
		"pt": "PT", // 葡萄牙语 -> 葡萄牙
		"ru": "RU", // 俄语 -> 俄罗斯
		"ja": "JP", // 日语 -> 日本
		"pa": "IN", // 旁遮普语 -> 印度

		// === 主要亚洲语言 ===
		"ko": "KR", // 韩语 -> 韩国
		"id": "ID", // 印尼语 -> 印度尼西亚
		"th": "TH", // 泰语 -> 泰国
		"vi": "VN", // 越南语 -> 越南
		"ms": "MY", // 马来语 -> 马来西亚
		"fa": "IR", // 波斯语 -> 伊朗
		"ur": "PK", // 乌尔都语 -> 巴基斯坦
		"ta": "IN", // 泰米尔语 -> 印度
		"te": "IN", // 泰卢固语 -> 印度
		"mr": "IN", // 马拉地语 -> 印度
		"gu": "IN", // 古吉拉特语 -> 印度
		"kn": "IN", // 卡纳达语 -> 印度
		"ml": "IN", // 马拉雅拉姆语 -> 印度
		"si": "LK", // 僧伽罗语 -> 斯里兰卡
		"ne": "NP", // 尼泊尔语 -> 尼泊尔
		"my": "MM", // 缅甸语 -> 缅甸
		"km": "KH", // 高棉语 -> 柬埔寨
		"lo": "LA", // 老挝语 -> 老挝
		"ka": "GE", // 格鲁吉亚语 -> 格鲁吉亚
		"hy": "AM", // 亚美尼亚语 -> 亚美尼亚
		"az": "AZ", // 阿塞拜疆语 -> 阿塞拜疆
		"uz": "UZ", // 乌兹别克语 -> 乌兹别克斯坦
		"kk": "KZ", // 哈萨克语 -> 哈萨克斯坦
		"mn": "MN", // 蒙古语 -> 蒙古

		// === 主要欧洲语言 ===
		"fr": "FR", // 法语 -> 法国
		"de": "DE", // 德语 -> 德国
		"it": "IT", // 意大利语 -> 意大利
		"pl": "PL", // 波兰语 -> 波兰
		"nl": "NL", // 荷兰语 -> 荷兰
		"sv": "SE", // 瑞典语 -> 瑞典
		"da": "DK", // 丹麦语 -> 丹麦
		"no": "NO", // 挪威语 -> 挪威
		"fi": "FI", // 芬兰语 -> 芬兰
		"el": "GR", // 希腊语 -> 希腊
		"tr": "TR", // 土耳其语 -> 土耳其
		"cs": "CZ", // 捷克语 -> 捷克
		"hu": "HU", // 匈牙利语 -> 匈牙利
		"ro": "RO", // 罗马尼亚语 -> 罗马尼亚
		"uk": "UA", // 乌克兰语 -> 乌克兰
		"bg": "BG", // 保加利亚语 -> 保加利亚
		"sr": "RS", // 塞尔维亚语 -> 塞尔维亚
		"hr": "HR", // 克罗地亚语 -> 克罗地亚
		"sk": "SK", // 斯洛伐克语 -> 斯洛伐克
		"sl": "SI", // 斯洛文尼亚语 -> 斯洛文尼亚
		"lt": "LT", // 立陶宛语 -> 立陶宛
		"lv": "LV", // 拉脱维亚语 -> 拉脱维亚
		"et": "EE", // 爱沙尼亚语 -> 爱沙尼亚
		"sq": "AL", // 阿尔巴尼亚语 -> 阿尔巴尼亚
		"mk": "MK", // 马其顿语 -> 北马其顿
		"is": "IS", // 冰岛语 -> 冰岛
		"ga": "IE", // 爱尔兰语 -> 爱尔兰
		"cy": "GB", // 威尔士语 -> 英国
		"mt": "MT", // 马耳他语 -> 马耳他
		"eu": "ES", // 巴斯克语 -> 西班牙
		"ca": "ES", // 加泰罗尼亚语 -> 西班牙
		"gl": "ES", // 加利西亚语 -> 西班牙

		// === 非洲语言 ===
		"sw": "TZ", // 斯瓦希里语 -> 坦桑尼亚
		"am": "ET", // 阿姆哈拉语 -> 埃塞俄比亚
		"ha": "NG", // 豪萨语 -> 尼日利亚
		"yo": "NG", // 约鲁巴语 -> 尼日利亚
		"ig": "NG", // 伊博语 -> 尼日利亚
		"zu": "ZA", // 祖鲁语 -> 南非
		"xh": "ZA", // 科萨语 -> 南非
		"af": "ZA", // 南非荷兰语 -> 南非
		"so": "SO", // 索马里语 -> 索马里
		"mg": "MG", // 马达加斯加语 -> 马达加斯加
		"rw": "RW", // 卢旺达语 -> 卢旺达

		// === 美洲原住民语言 ===
		"qu": "PE", // 克丘亚语 -> 秘鲁
		"gn": "PY", // 瓜拉尼语 -> 巴拉圭
		"ay": "BO", // 艾马拉语 -> 玻利维亚

		// === 中东语言 ===
		"he": "IL", // 希伯来语 -> 以色列
		"ku": "IQ", // 库尔德语 -> 伊拉克
		"ps": "AF", // 普什图语 -> 阿富汗

		// === 其他重要小语种 ===
		"tl":  "PH",  // 他加禄语(菲律宾语) -> 菲律宾
		"fil": "PH",  // 菲律宾语(官方代码) -> 菲律宾
		"ceb": "PH",  // 宿务语 -> 菲律宾
		"be":  "BY",  // 白俄罗斯语 -> 白俄罗斯
		"bs":  "BA",  // 波斯尼亚语 -> 波黑
		"eo":  "001", // 世界语 -> 国际通用(001表示世界)
	}

	// 查找默认地区
	if region, ok := defaultRegions[baseStr]; ok {
		// 使用 language.Make 创建新的标签
		newTag := language.MustParse(baseStr + "-" + region)
		return newTag.String(), nil
	}

	// 没有默认地区映射，拒绝不支持的语言代码
	// 返回错误，提示用户该语言代码不被支持
	return "", fmt.Errorf("unsupported language code: %s (please use common language codes like zh, en, ja, etc.)", baseStr)
}

// UnicodeNormalizeOption 定义 Unicode 字符串规范化选项
// 这些选项用于字符串的规范化处理，适用于比较、匹配、转换等场景
type UnicodeNormalizeOption int32

const (
	// IgnoreCase 忽略大小写进行比较（使用 Unicode case folding 标准）
	//
	// Unicode 标准定义：case-insensitive 比较不仅包含基本大小写（A↔a），
	// 还包含特殊字符的规范映射，这是 Unicode 标准推荐的正确实现方式。
	//
	// 特殊字符映射示例：
	//   - 基本大小写: "Hello" == "HELLO"
	//   - 德语 ß: "Straße" == "strasse" (ß 折叠为 ss)
	//   - 希腊语 Σ: "ΣΊΓΜΑ" == "σίγμα"
	//   - 土耳其语 İ: "İstanbul" == "i̇stanbul"
	//
	// 注意：Go 的 strings.EqualFold() 和 Java 的 equalsIgnoreCase() 仅实现了基本大小写转换（不完整实现），无法正确处理 ß→ss 等映射。
	// Python 的 str.casefold() 和 .NET 的 String.Compare(ignoreCase) 使用了完整的 Unicode case folding，与本实现一致。
	IgnoreCase UnicodeNormalizeOption = 1 << iota // 1

	// IgnoreDiacritics 忽略变音符号（音调符号）
	// 示例: "café" == "cafe", "naïve" == "naive"
	IgnoreDiacritics // 2

	// IgnoreWidth 忽略全角/半角字符差异
	// 示例: "Ａ" == "A", "１２３" == "123"
	IgnoreWidth // 4

	// IgnoreCompatibility 忽略兼容性字符差异（使用 NFKC 规范化）
	// 示例: "①②③" == "123", "ⒶⒷⒸ" == "ABC", "㎡" == "m2"
	// 注意: 这会将带圈数字、带括号字母、罗马数字、平方米等兼容性字符转换为基本形式
	IgnoreCompatibility // 8

	// Loose 宽松比较，等价于 IgnoreCase | IgnoreDiacritics | IgnoreWidth | IgnoreCompatibility (= 15)
	// 同时应用所有规范化：case folding + 移除变音符号 + 全角转半角 + 兼容性字符转换
	// 示例: "ＣＡＦÉ" == "cafe", "Straße" == "strasse", "①②③" == "123"
	Loose = IgnoreCase | IgnoreDiacritics | IgnoreWidth | IgnoreCompatibility // 15
)

// mergeUnicodeNormalizeOptions 合并多个 UnicodeNormalizeOption 为单个选项
func mergeUnicodeNormalizeOptions(options ...UnicodeNormalizeOption) UnicodeNormalizeOption {
	var combined UnicodeNormalizeOption
	for _, opt := range options {
		combined |= opt
	}
	return combined
}

// isCaseIgnore 检查是否忽略大小写
func (opt UnicodeNormalizeOption) isCaseIgnore() bool {
	return opt&IgnoreCase != 0
}

// isDiacriticsIgnore 检查是否忽略变音符号
func (opt UnicodeNormalizeOption) isDiacriticsIgnore() bool {
	return opt&IgnoreDiacritics != 0
}

// isWidthIgnore 检查是否忽略全角半角
func (opt UnicodeNormalizeOption) isWidthIgnore() bool {
	return opt&IgnoreWidth != 0
}

// isCompatibilityIgnore 检查是否忽略兼容性字符
func (opt UnicodeNormalizeOption) isCompatibilityIgnore() bool {
	return opt&IgnoreCompatibility != 0
}

// UnicodeCompare 判断两个 Unicode 字符串是否相等
// 支持忽略大小写、变音符号、全角半角、兼容性字符等选项
// 除了emoji之外，其它规则和MySQL 8的utf8mb4_unicode_ci保持一致
//
// 参数：
//   - str1, str2: 要比较的两个字符串
//   - options: 规范化选项，可以是多个选项的位运算组合
//   - 无选项: 严格比较（str1 == str2）
//   - IgnoreCase: 忽略大小写
//   - IgnoreDiacritics: 忽略变音符号
//   - IgnoreWidth: 忽略全角半角
//   - IgnoreCompatibility: 忽略兼容性字符
//   - Loose: 宽松比较（以上所有选项的组合）
//
// 示例:
//
//	UnicodeCompare("Hello", "hello", IgnoreCase)                    // true
//	UnicodeCompare("café", "cafe", IgnoreCase, IgnoreDiacritics)    // true
//	UnicodeCompare("Straße", "strasse", IgnoreCase)                 // true (ß -> ss)
//	UnicodeCompare("ＨＥＬＬＯ", "hello", IgnoreCase, IgnoreWidth)   // true
//	UnicodeCompare("①②③", "123", IgnoreCompatibility)              // true (带圈数字)
//
// 实现说明:
//   - 使用 UnicodeCanonical 将两个字符串规范化后比较
//   - 保证与 UnicodeCanonical 的行为完全一致
func UnicodeCompare(str1, str2 string, options ...UnicodeNormalizeOption) bool {
	// 合并所有选项
	opt := mergeUnicodeNormalizeOptions(options...)

	// 无选项时严格比较
	if opt == 0 {
		return str1 == str2
	}

	// 快速路径: 完全相同
	if str1 == str2 {
		return true
	}

	// 规范化后比较（复用 UnicodeCanonical 逻辑）
	return UnicodeCanonical(str1, options...) == UnicodeCanonical(str2, options...)
}

// UnicodeCanonical 返回字符串的 Unicode 规范形式。
// 返回的字符串是可读的，不是二进制数据。
//
// 参数：
//   - input: 要规范化的字符串
//   - options: 规范化选项，可以是多个选项的位运算组合
//   - 无选项: 返回原字符串（严格匹配）
//   - IgnoreCase: 转换为 case folding 形式（如 "Straße" -> "strasse"）
//   - IgnoreDiacritics: 移除变音符号（如 "café" -> "cafe"）
//   - IgnoreWidth: 全角转半角（如 "Ａ" -> "A"）
//   - IgnoreCompatibility: 兼容性字符转换（如 "①②③" -> "123", "ⒶⒷⒸ" -> "ABC"）
//   - Loose: 以上所有选项的组合
//
// 注意:
//   - 返回的是可读的规范化字符串，不是二进制数据
//   - 相同的输入和选项总是返回相同的结果
func UnicodeCanonical(input string, options ...UnicodeNormalizeOption) string {
	// 合并所有选项
	opt := mergeUnicodeNormalizeOptions(options...)

	// 快速路径: 无选项，返回原字符串
	if opt == 0 {
		return input
	}

	result := input

	// IgnoreCompatibility: 使用 NFKC 规范化（兼容性分解 + 组合）
	// 注意: 必须在其他转换之前执行，因为 NFKC 会转换字符形式
	// 例如: "①②③" -> "123", "ⒶⒷⒸ" -> "ABC", "㎡" -> "m2"
	if opt.isCompatibilityIgnore() {
		result = norm.NFKC.String(result)
	}

	// IgnoreCase: 使用 case folding（而非 ToLower）
	// Case folding 是 Unicode 标准中用于不区分大小写比较的正确方法
	// 例如: "Straße" -> "strasse", "İstanbul" -> "i̇stanbul"
	if opt.isCaseIgnore() {
		result = cases.Fold().String(result)
	}

	// IgnoreDiacritics: 移除变音符号
	// 使用 transform.Chain 实现 NFD → 移除 Mn → NFC 的流式处理
	if opt.isDiacriticsIgnore() {
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		result, _, _ = transform.String(t, result)
	}

	// IgnoreWidth: 转换全角到半角
	if opt.isWidthIgnore() {
		result = width.Narrow.String(result)
	}

	return result
}
