package utils

import (
	"strings"

	"golang.org/x/text/language"
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
	baseStr := base.String()
	if region, ok := defaultRegions[baseStr]; ok {
		// 使用 language.Make 创建新的标签
		newTag := language.MustParse(baseStr + "-" + region)
		return newTag.String(), nil
	}

	// 没有默认地区映射,返回原始标签
	return tag.String(), nil
}
