package utils

import (
	"strings"
	"testing"

	"golang.org/x/text/language"
)

func TestRFC5646LanguageDebug(t *testing.T) {
	input := "zh"
	tag, err := language.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	t.Logf("Input: %s", input)
	t.Logf("tag.String(): %s", tag.String())

	region, confidence := tag.Region()
	t.Logf("Region: %s, Confidence: %v", region, confidence)

	base, _ := tag.Base()
	t.Logf("Base: %s", base.String())

	// 测试构建新标签
	newTag := language.MustParse(base.String() + "-CN")
	t.Logf("NewTag: %s", newTag.String())

	// 测试实际函数
	result, err := NormalizeRFC5646Language(input)
	if err != nil {
		t.Fatalf("RFC5646Language error: %v", err)
	}
	t.Logf("Result: %s", result)
}

func TestRFC5646Language(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// 测试自动补充地区
		{
			name:     "Chinese without region should add CN",
			input:    "zh",
			expected: "zh-CN",
			wantErr:  false,
		},
		{
			name:     "English without region should add US",
			input:    "en",
			expected: "en-US",
			wantErr:  false,
		},
		{
			name:     "Korean without region should add KR",
			input:    "ko",
			expected: "ko-KR",
			wantErr:  false,
		},
		{
			name:     "Japanese without region should add JP",
			input:    "ja",
			expected: "ja-JP",
			wantErr:  false,
		},
		{
			name:     "Spanish without region should add ES",
			input:    "es",
			expected: "es-ES",
			wantErr:  false,
		},
		{
			name:     "Portuguese without region should add PT",
			input:    "pt",
			expected: "pt-PT",
			wantErr:  false,
		},
		{
			name:     "French without region should add FR",
			input:    "fr",
			expected: "fr-FR",
			wantErr:  false,
		},
		{
			name:     "German without region should add DE",
			input:    "de",
			expected: "de-DE",
			wantErr:  false,
		},
		{
			name:     "Russian without region should add RU",
			input:    "ru",
			expected: "ru-RU",
			wantErr:  false,
		},
		{
			name:     "Italian without region should add IT",
			input:    "it",
			expected: "it-IT",
			wantErr:  false,
		},
		{
			name:     "Thai without region should add TH",
			input:    "th",
			expected: "th-TH",
			wantErr:  false,
		},
		{
			name:     "Vietnamese without region should add VN",
			input:    "vi",
			expected: "vi-VN",
			wantErr:  false,
		},

		// 测试小语种自动补充地区
		{
			name:     "Swahili without region should add TZ",
			input:    "sw",
			expected: "sw-TZ",
			wantErr:  false,
		},
		{
			name:     "Urdu without region should add PK",
			input:    "ur",
			expected: "ur-PK",
			wantErr:  false,
		},
		{
			name:     "Tamil without region should add IN",
			input:    "ta",
			expected: "ta-IN",
			wantErr:  false,
		},
		{
			name:     "Bengali without region should add BD",
			input:    "bn",
			expected: "bn-BD",
			wantErr:  false,
		},
		{
			name:     "Burmese without region should add MM",
			input:    "my",
			expected: "my-MM",
			wantErr:  false,
		},
		{
			name:     "Khmer without region should add KH",
			input:    "km",
			expected: "km-KH",
			wantErr:  false,
		},
		{
			name:     "Georgian without region should add GE",
			input:    "ka",
			expected: "ka-GE",
			wantErr:  false,
		},
		{
			name:     "Kazakh without region should add KZ",
			input:    "kk",
			expected: "kk-KZ",
			wantErr:  false,
		},
		{
			name:     "Serbian without region should add RS",
			input:    "sr",
			expected: "sr-RS",
			wantErr:  false,
		},
		{
			name:     "Tagalog without region should add PH",
			input:    "tl",
			expected: "fil-PH", // tl 会被规范化为 fil
			wantErr:  false,
		},
		{
			name:     "Filipino without region should add PH",
			input:    "fil",
			expected: "fil-PH",
			wantErr:  false,
		},

		// 测试已有地区信息的情况(应保持不变)
		{
			name:     "Chinese with TW region should keep TW",
			input:    "zh-TW",
			expected: "zh-TW",
			wantErr:  false,
		},
		{
			name:     "Chinese with HK region should keep HK",
			input:    "zh-HK",
			expected: "zh-HK",
			wantErr:  false,
		},
		{
			name:     "English with GB region should keep GB",
			input:    "en-GB",
			expected: "en-GB",
			wantErr:  false,
		},
		{
			name:     "Portuguese with BR region should keep BR",
			input:    "pt-BR",
			expected: "pt-BR",
			wantErr:  false,
		},
		{
			name:     "Spanish with MX region should keep MX",
			input:    "es-MX",
			expected: "es-MX",
			wantErr:  false,
		},

		// 测试大小写不敏感
		{
			name:     "Upper case ZH should work",
			input:    "ZH",
			expected: "zh-CN",
			wantErr:  false,
		},
		{
			name:     "Mixed case En-Us should work",
			input:    "En-Us",
			expected: "en-US",
			wantErr:  false,
		},

		// 测试无效输入
		{
			name:     "Invalid language code should error",
			input:    "invalid",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Empty string should error",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Unsupported ISO 639-3 code 'abc' should error",
			input:    "abc",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Unsupported ISO 639-3 code 'mie' should error",
			input:    "mie",
			expected: "",
			wantErr:  true,
		},

		// 测试地区代码输入
		{
			name:     "CN region code should convert to zh-CN",
			input:    "cn",
			expected: "zh-CN",
			wantErr:  false,
		},
		{
			name:     "TW region code should convert to zh-TW",
			input:    "tw",
			expected: "zh-TW",
			wantErr:  false,
		},
		{
			name:     "HK region code should convert to zh-HK",
			input:    "hk",
			expected: "zh-HK",
			wantErr:  false,
		},
		{
			name:     "US region code should convert to en-US",
			input:    "us",
			expected: "en-US",
			wantErr:  false,
		},
		{
			name:     "UK region code should convert to en-GB",
			input:    "uk",
			expected: "en-GB",
			wantErr:  false,
		},
		{
			name:     "JP region code should convert to ja-JP",
			input:    "jp",
			expected: "ja-JP",
			wantErr:  false,
		},
		{
			name:     "KR region code should convert to ko-KR",
			input:    "kr",
			expected: "ko-KR",
			wantErr:  false,
		},
		{
			name:     "MX region code should convert to es-MX",
			input:    "mx",
			expected: "es-MX",
			wantErr:  false,
		},
		{
			name:     "SG region code should convert to zh-SG",
			input:    "sg",
			expected: "zh-SG",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeRFC5646Language(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeRFC5646Language() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				// 调试输出
				tag, _ := language.Parse(tt.input)
				region, confidence := tag.Region()
				base, _ := tag.Base()
				t.Logf("Debug - Input: %s, Result: %s, Expected: %s", tt.input, result, tt.expected)
				t.Logf("Debug - tag.String(): %s, Region: %s, Confidence: %v, Base: %s",
					tag.String(), region, confidence, base.String())
				t.Errorf("NormalizeRFC5646Language() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestUnicodeEqualFold 测试 Unicode 大小写不敏感比较
func TestUnicodeEqualFold(t *testing.T) {
	tests := []struct {
		name     string
		str1     string
		str2     string
		expected bool
	}{
		// ASCII 快速路径
		{
			name:     "ASCII - 相同字符串",
			str1:     "Hello",
			str2:     "Hello",
			expected: true,
		},
		{
			name:     "ASCII - 大小写不同",
			str1:     "Hello World",
			str2:     "HELLO WORLD",
			expected: true,
		},
		{
			name:     "ASCII - 混合大小写",
			str1:     "HeLLo WoRLd",
			str2:     "hello world",
			expected: true,
		},
		{
			name:     "ASCII - 不相等",
			str1:     "Hello",
			str2:     "World",
			expected: false,
		},

		// Unicode 字符
		{
			name:     "Unicode - 法语重音",
			str1:     "café",
			str2:     "CAFÉ",
			expected: true,
		},
		{
			name:     "Unicode - 德语 ß 小写",
			str1:     "Straße",
			str2:     "straße",
			expected: true,
		},
		{
			name:     "Unicode - 德语 ẞ 大写",
			str1:     "STRAẞE",
			str2:     "straße",
			expected: true,
		},
		{
			name:     "Unicode - 希腊字母",
			str1:     "Ελληνικά",
			str2:     "ελληνικά",
			expected: true,
		},
		{
			name:     "Unicode - 俄语西里尔字母",
			str1:     "Привет",
			str2:     "привет",
			expected: true,
		},

		// Unicode 组合字符（NFD vs NFC）
		{
			name:     "Unicode - 组合字符 é",
			str1:     "café", // é = U+00E9 (NFC)
			str2:     "café", // é = e + ́ (U+0065 + U+0301, NFD)
			expected: true,
		},
		{
			name:     "Unicode - 组合字符混合大小写",
			str1:     "CAFÉ", // É = U+00C9 (NFC)
			str2:     "café", // é = e + ́ (NFD)
			expected: true,
		},

		// 边界情况
		{
			name:     "边界 - 空字符串",
			str1:     "",
			str2:     "",
			expected: true,
		},
		{
			name:     "边界 - 空字符串 vs 非空",
			str1:     "",
			str2:     "hello",
			expected: false,
		},
		{
			name:     "边界 - 单个字符",
			str1:     "A",
			str2:     "a",
			expected: true,
		},

		// 混合 ASCII 和 Unicode
		{
			name:     "混合 - ASCII + Unicode",
			str1:     "Hello café",
			str2:     "HELLO CAFÉ",
			expected: true,
		},
		{
			name:     "混合 - 中英文",
			str1:     "Hello 世界",
			str2:     "HELLO 世界",
			expected: true,
		},

		// 土耳其语特殊大小写（注意：语言无关的 Unicode 转换不处理土耳其语特殊规则）
		{
			name:     "土耳其语 - ı 小写",
			str1:     "ışık",
			str2:     "ışık",
			expected: true,
		},
		{
			name:     "土耳其语 - Ş 大写",
			str1:     "BAŞLIK",
			str2:     "başlik",
			expected: true,
		},

		// 北欧语言
		{
			name:     "北欧 - 挪威语 ø",
			str1:     "København",
			str2:     "københavn",
			expected: true,
		},
		{
			name:     "北欧 - 瑞典语 å",
			str1:     "Ångström",
			str2:     "ångström",
			expected: true,
		},

		// Unicode case folding 的标准行为（cases.Fold 能正确处理）
		{
			name:     "Case folding - 德语 ß → ss",
			str1:     "Straße",
			str2:     "STRASSE",
			expected: true, // cases.Fold() 会将 ß 和 SS 都转为 ss
		},
		{
			name:     "Case folding - 连字符 ﬁ → fi",
			str1:     "ﬁnance",
			str2:     "FINANCE",
			expected: true, // cases.Fold() 会将 ﬁ 展开为 fi
		},
		{
			name:     "Case folding - 连字符 ﬂ → fl",
			str1:     "ﬂower",
			str2:     "FLOWER",
			expected: true, // cases.Fold() 会将 ﬂ 展开为 fl
		},
		{
			name:     "Case folding - 希腊语 Σ (词中)",
			str1:     "ΣΊΓΜΑ",
			str2:     "σίγμα",
			expected: true, // cases.Fold() 会将大写 Σ 转为小写 σ
		},
		{
			name:     "Case folding - 希腊语 ς (词尾)",
			str1:     "τέλος",
			str2:     "ΤΈΛΟΣ",
			expected: true, // cases.Fold() 会统一处理词尾 ς
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnicodeCompare(tt.str1, tt.str2, IgnoreCase)
			if result != tt.expected {
				t.Errorf("UnicodeCompare(%q, %q, IgnoreCase) = %v, want %v", tt.str1, tt.str2, result, tt.expected)
			}
		})
	}
}

// TestUnicodeCompare 测试 Unicode 字符串比较
func TestUnicodeCompare(t *testing.T) {
	tests := []struct {
		name     string
		str1     string
		str2     string
		options  UnicodeNormalizeOption
		expected bool
	}{
		// === 基本测试 ===
		{
			name:     "相同字符串",
			str1:     "Hello",
			str2:     "Hello",
			options:  0,
			expected: true,
		},
		{
			name:     "不同字符串",
			str1:     "Hello",
			str2:     "World",
			options:  0,
			expected: false,
		},

		// === IgnoreCase 测试 ===
		{
			name:     "忽略大小写 - café vs Café",
			str1:     "café",
			str2:     "Café",
			options:  IgnoreCase,
			expected: true,
		},
		{
			name:     "忽略大小写 - ASCII",
			str1:     "Hello World",
			str2:     "HELLO WORLD",
			options:  IgnoreCase,
			expected: true,
		},

		// === IgnoreDiacritics 测试 ===
		{
			name:     "忽略变音符号 - café vs cafe",
			str1:     "café",
			str2:     "cafe",
			options:  IgnoreDiacritics,
			expected: true, // 单独使用 IgnoreDiacritics 也会相等
		},
		{
			name:     "忽略变音符号+大小写 - CAFÉ vs cafe",
			str1:     "CAFÉ",
			str2:     "cafe",
			options:  IgnoreCase | IgnoreDiacritics,
			expected: true,
		},
		{
			name:     "忽略变音符号+大小写 - résumé vs RESUME",
			str1:     "résumé",
			str2:     "RESUME",
			options:  IgnoreCase | IgnoreDiacritics,
			expected: true,
		},

		// === IgnoreWidth 测试 ===
		{
			name:     "忽略宽度 - 全角A vs 半角A",
			str1:     "Ａ",
			str2:     "A",
			options:  IgnoreWidth,
			expected: true, // 单独使用 IgnoreWidth 也会相等
		},
		{
			name:     "忽略宽度+大小写 - Ａ vs a",
			str1:     "Ａ",
			str2:     "a",
			options:  IgnoreCase | IgnoreWidth,
			expected: true,
		},
		{
			name:     "忽略宽度+大小写 - １２３ vs 123",
			str1:     "１２３",
			str2:     "123",
			options:  IgnoreCase | IgnoreWidth,
			expected: true,
		},

		// === 组合选项测试 ===
		{
			name:     "忽略大小写+变音符号 - CAFÉ vs cafe",
			str1:     "CAFÉ",
			str2:     "cafe",
			options:  IgnoreCase | IgnoreDiacritics,
			expected: true,
		},
		{
			name:     "忽略大小写+宽度 - Ａ vs a",
			str1:     "Ａ",
			str2:     "a",
			options:  IgnoreCase | IgnoreWidth,
			expected: true,
		},
		{
			name:     "忽略大小写+变音符号+宽度 - ＣＡＦÉ vs cafe",
			str1:     "ＣＡＦÉ", // 全角 + 变音符号 + 大写
			str2:     "cafe",
			options:  IgnoreCase | IgnoreDiacritics | IgnoreWidth,
			expected: true,
		},

		// === Loose 模式测试 ===
		{
			name:     "德语 ß - 大小写敏感",
			str1:     "Straße",
			str2:     "Strasse",
			options:  0,
			expected: false,
		},
		{
			name:     "德语 ß - 宽松比较",
			str1:     "Straße",
			str2:     "Strasse",
			options:  Loose,
			expected: true,
		},
		{
			name:     "德语 ß - 忽略大小写 + 宽松",
			str1:     "Straße",
			str2:     "STRASSE",
			options:  IgnoreCase | Loose,
			expected: true,
		},
		{
			name:     "Loose 模式 - 综合测试",
			str1:     "ＣＡＦＥ́", // 全角 + 变音符号
			str2:     "cafe",
			options:  Loose,
			expected: true,
		},

		// === Unicode 组合字符测试 ===
		{
			name:     "Unicode 组合字符 - NFD vs NFC",
			str1:     "café", // é = U+00E9
			str2:     "café", // é = e + ́ (U+0065 + U+0301)
			options:  0,
			expected: true, // collate 自动规范化
		},

		// === 边界测试 ===
		{
			name:     "空字符串",
			str1:     "",
			str2:     "",
			options:  IgnoreCase,
			expected: true,
		},
		{
			name:     "空字符串 vs 非空",
			str1:     "",
			str2:     "hello",
			options:  IgnoreCase,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnicodeCompare(tt.str1, tt.str2, tt.options)
			if result != tt.expected {
				t.Errorf("UnicodeCompare(%q, %q, options=%v) = %v, want %v", tt.str1, tt.str2, tt.options, result, tt.expected)
			}
		})
	}
}

// TestUnicodeCanonical 测试 Unicode 规范化键生成
func TestUnicodeCanonical(t *testing.T) {
	// 测试基本功能 - 相同字符串（大小写不同）
	key1 := UnicodeCanonical("Hello", IgnoreCase)
	key2 := UnicodeCanonical("HELLO", IgnoreCase)

	if key1 != key2 {
		t.Errorf("相同字符串（忽略大小写）应生成相同规范化键")
	}

	// 测试作为 map key
	termDB := make(map[string]string)
	termDB[UnicodeCanonical("Straße", IgnoreCase)] = "street (German)"
	termDB[UnicodeCanonical("Cœur", IgnoreCase)] = "heart (French)"
	termDB[UnicodeCanonical("Hello", IgnoreCase)] = "greeting (English)"

	// 使用相同字符串的不同大小写查询
	query1 := UnicodeCanonical("straße", IgnoreCase)
	if val, found := termDB[query1]; !found || val != "street (German)" {
		t.Errorf("查询失败: found: %v, value: %q", found, val)
	}

	query2 := UnicodeCanonical("cœur", IgnoreCase)
	if val, found := termDB[query2]; !found || val != "heart (French)" {
		t.Errorf("查询失败: found: %v, value: %q", found, val)
	}

	query3 := UnicodeCanonical("HELLO", IgnoreCase)
	if val, found := termDB[query3]; !found || val != "greeting (English)" {
		t.Errorf("查询失败: found: %v, value: %q", found, val)
	}
}

// BenchmarkUnicodeEqualFold_ASCII 测试 ASCII 快速路径性能
func BenchmarkUnicodeEqualFold_ASCII(b *testing.B) {
	str1 := "Hello World This Is A Test String"
	str2 := "HELLO WORLD THIS IS A TEST STRING"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UnicodeCompare(str1, str2, IgnoreCase)
	}
}

// BenchmarkUnicodeEqualFold_Unicode 测试 Unicode 完整路径性能
func BenchmarkUnicodeEqualFold_Unicode(b *testing.B) {
	str1 := "Straße café Þórshöfn Ångström"
	str2 := "STRAẞE CAFÉ ÞÓRSHÖFN ÅNGSTRÖM"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UnicodeCompare(str1, str2, IgnoreCase)
	}
}

// BenchmarkUnicodeEqualFold_Mixed 测试混合 ASCII + Unicode 性能
func BenchmarkUnicodeEqualFold_Mixed(b *testing.B) {
	str1 := "Hello café world Straße test"
	str2 := "HELLO CAFÉ WORLD STRAẞE TEST"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UnicodeCompare(str1, str2, IgnoreCase)
	}
}

// BenchmarkStringsEqualFold 对比标准库性能
func BenchmarkStringsEqualFold(b *testing.B) {
	str1 := "Hello World This Is A Test String"
	str2 := "HELLO WORLD THIS IS A TEST STRING"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.EqualFold(str1, str2)
	}
}

// BenchmarkUnicodeCompare_IgnoreCase 对比 UnicodeCompare 性能
func BenchmarkUnicodeCompare_IgnoreCase(b *testing.B) {
	str1 := "Hello World This Is A Test String"
	str2 := "HELLO WORLD THIS IS A TEST STRING"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UnicodeCompare(str1, str2, IgnoreCase)
	}
}

// BenchmarkUnicodeCanonical 性能基准测试
func BenchmarkUnicodeCanonical(b *testing.B) {
	inputs := []string{
		"Straße", "café", "Þórshöfn", "Cœur", "Ångström",
		"Hello World", "中文测试", "Привет", "naïve", "résumé",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range inputs {
			_ = UnicodeCanonical(s, IgnoreCase)
		}
	}
}

// TestUnicodeCompareDiacritics 测试各种语言的变音符号（Loose 模式）
// 对应 SQL 中的 french_diaeresis, french_cedilla, french_acute, portuguese_tilde,
// spanish_tilde, german_umlaut, danish_ring, french_capital_accent 等测试
//
// 注意：此测试对标 MySQL utf8mb4_unicode_ci 排序规则的实际行为
//
// 已知差异（Go 无法复现）：
// 1. emoji_different: MySQL 将 '😀' = '🙂' 视为相等 (1)，但这不符合 Unicode 标准
//   - 原因：MySQL 的 utf8mb4_unicode_ci 可能对 emoji 使用简化的排序权重
//   - Go 行为：遵循 Unicode 标准，不同 emoji 码点不相等
//   - 结论：这是 MySQL 特有行为，不建议依赖
func TestUnicodeCompareDiacritics(t *testing.T) {
	tests := []struct {
		name     string
		str1     string
		str2     string
		expected bool
		comment  string
	}{
		// 法语变音符号
		{
			name:     "French diaeresis - naïve",
			str1:     "naïve",
			str2:     "naive",
			expected: true,
			comment:  "法语分音符（¨）",
		},
		{
			name:     "French cedilla - façade",
			str1:     "façade",
			str2:     "facade",
			expected: true,
			comment:  "法语下加符（ç）",
		},
		{
			name:     "French acute - élève",
			str1:     "élève",
			str2:     "eleve",
			expected: true,
			comment:  "法语尖音符（é/è）",
		},
		{
			name:     "French capital accent - Élodie",
			str1:     "Élodie",
			str2:     "Elodie",
			expected: true,
			comment:  "法语大写尖音符",
		},

		// 葡萄牙语波浪号
		{
			name:     "Portuguese tilde - pão",
			str1:     "pão",
			str2:     "pao",
			expected: true,
			comment:  "葡萄牙语波浪号（ã）",
		},

		// 西班牙语波浪号
		{
			name:     "Spanish tilde - niño",
			str1:     "niño",
			str2:     "nino",
			expected: true,
			comment:  "西班牙语波浪号（ñ）",
		},

		// 德语元音变音
		{
			name:     "German umlaut - Müller",
			str1:     "Müller",
			str2:     "Muller",
			expected: true,
			comment:  "德语变音符号（ü）",
		},

		// 丹麦语/北欧语圆圈
		{
			name:     "Danish ring - Århus",
			str1:     "Århus",
			str2:     "Arhus",
			expected: true,
			comment:  "丹麦语上圆圈（å）",
		},

		// === SQL 对应测试用例 ===
		// 以下测试用例对应 MySQL utf8mb4_unicode_ci 的实际执行结果
		// 注意：MySQL 结果中仅 greek_sigma_fold = 0，其他都 = 1

		// 1. 德语 ß 折叠（SQL: german_ß_fold = 1 ✅）
		{
			name:     "SQL: german_ß_fold - Straße vs STRASSE",
			str1:     "Straße",
			str2:     "STRASSE",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "德语 ß → ss 大小写折叠",
		},

		// 2. 土耳其语 I 问题（SQL: turkish_I_issue = 1 ✅, turkish_i_issue = 1 ✅）
		{
			name:     "SQL: turkish_I_issue - İSTANBUL vs istanbul",
			str1:     "İSTANBUL", // İ = U+0130
			str2:     "istanbul",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "土耳其语 İ vs i",
		},
		{
			name:     "SQL: turkish_i_issue - Istanbul vs istanbul",
			str1:     "Istanbul",
			str2:     "istanbul",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "土耳其语普通 I vs i",
		},

		// 3. 希腊语测试（SQL: greek_sigma_fold = 0 ❌ [唯一失败], greek_accents_fold = 1 ✅）
		{
			name:     "SQL: greek_sigma_fold - ΣΊΓΜΑ vs σígμα (ONLY FAIL CASE)",
			str1:     "ΣΊΓΜΑ",
			str2:     "σígμα", // í 是拉丁字母 U+00ED，不是希腊字母 ί U+03AF
			expected: false,   // MySQL: 0 ❌, Go: false ❌ [唯一失败的测试]
			comment:  "希腊语 Sigma 混合拉丁字母 í - MySQL 和 Go 都不相等",
		},
		{
			name:     "SQL: greek_accents_fold - ΠΕΡΙΣΣΌΤΑΤΟ vs περισσότατο",
			str1:     "ΠΕΡΙΣΣΌΤΑΤΟ",
			str2:     "περισσότατο",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "希腊语重音符号折叠",
		},

		// 补充：使用正确希腊字母的测试（不在原 SQL 中）
		{
			name:     "Greek sigma - ΣΊΓΜΑ vs σίγμα (correct Greek)",
			str1:     "ΣΊΓΜΑ",
			str2:     "σίγμα", // 使用正确的希腊字母 ί U+03AF
			expected: true,
			comment:  "希腊语 Sigma（使用正确的希腊字母 ί）",
		},
		{
			name:     "Greek sigma - final form ς",
			str1:     "τέλος",
			str2:     "ΤΈΛΟΣ",
			expected: true,
			comment:  "希腊语 Sigma 词尾形式（ς）",
		},
		{
			name:     "Greek sigma - σ vs ς",
			str1:     "σ",
			str2:     "ς",
			expected: true,
			comment:  "希腊语两种小写 Sigma 形式",
		},

		// 4. 全角与半角字符（SQL: full_width_* 都 = 1 ✅）
		{
			name:     "SQL: full_width_digits - １２３４５６７８９０ vs 1234567890",
			str1:     "１２３４５６７８９０",
			str2:     "1234567890",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "全角数字 vs 半角数字",
		},
		{
			name:     "SQL: full_width_letters - ＡＢＣＤＥＦＧ vs ABCDEFG",
			str1:     "ＡＢＣＤＥＦＧ",
			str2:     "ABCDEFG",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "全角字母 vs 半角字母",
		},
		{
			name:     "SQL: full_width_symbols - ！＠＃＄％＾＆＊ vs !@#$%^&*",
			str1:     "！＠＃＄％＾＆＊",
			str2:     "!@#$%^&*",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "全角符号 vs 半角符号",
		},

		// 5. 各种语言的变音符号（SQL: 所有都 = 1 ✅）
		{
			name:     "SQL: french_accent - café vs cafe",
			str1:     "café",
			str2:     "cafe",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "法语尖音符（é）",
		},
		{
			name:     "SQL: french_diaeresis - naïve vs naive",
			str1:     "naïve",
			str2:     "naive",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "法语分音符（ï）",
		},
		{
			name:     "SQL: french_cedilla - façade vs facade",
			str1:     "façade",
			str2:     "facade",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "法语下加符（ç）",
		},
		{
			name:     "SQL: french_acute - élève vs eleve",
			str1:     "élève",
			str2:     "eleve",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "法语尖音符（é/è）",
		},
		{
			name:     "SQL: portuguese_tilde - pão vs pao",
			str1:     "pão",
			str2:     "pao",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "葡萄牙语波浪号（ã）",
		},
		{
			name:     "SQL: spanish_tilde - niño vs nino",
			str1:     "niño",
			str2:     "nino",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "西班牙语波浪号（ñ）",
		},
		{
			name:     "SQL: german_umlaut - Müller vs Muller",
			str1:     "Müller",
			str2:     "Muller",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "德语变音符号（ü）",
		},
		{
			name:     "SQL: danish_ring - Århus vs Arhus",
			str1:     "Århus",
			str2:     "Arhus",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "丹麦语上圆圈（å）",
		},
		{
			name:     "SQL: french_capital_accent - Élodie vs Elodie",
			str1:     "Élodie",
			str2:     "Elodie",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "法语大写尖音符（É）",
		},

		// 6. 带圈字符（SQL: circled_numbers = 1 ✅，Go: true ✅）
		{
			name:     "SQL: circled_numbers - ①②③ vs 123",
			str1:     "①②③",
			str2:     "123",
			expected: true, // MySQL: 1 ✅, Go: true ✅ [已修复：使用 NFKC]
			comment:  "带圈数字（使用 NFKC 规范化）",
		},

		// 7. 兼容性字符（SQL: compatibility_letters = 1 ✅，Go: true ✅）
		{
			name:     "SQL: compatibility_letters - ⒶⒷⒸ vs ABC",
			str1:     "ⒶⒷⒸ",
			str2:     "ABC",
			expected: true, // MySQL: 1 ✅, Go: true ✅ [已修复：使用 NFKC]
			comment:  "带括号的兼容性字母（使用 NFKC 规范化）",
		},

		// 8. Emoji 表情符号（SQL: emoji_identical = 1 ✅, emoji_different = 1 ⚠️）
		{
			name:     "SQL: emoji_identical - 😀😁😂 vs 😀😁😂",
			str1:     "😀😁😂",
			str2:     "😀😁😂",
			expected: true, // MySQL: 1 ✅, Go: true ✅
			comment:  "相同 emoji",
		},
		{
			name:     "SQL: emoji_different - 😀 vs 🙂",
			str1:     "😀",
			str2:     "🙂",
			expected: false, // MySQL: 1 ✅, Go: false ❌ [差异：MySQL 特殊行为，Go 无法复现]
			comment:  "不同 emoji（MySQL 将它们视为相等，但 Unicode 标准中它们是不同字符）",
		},

		// 综合测试（混合多种变音符号）
		{
			name:     "Mixed diacritics - café résumé naïve",
			str1:     "café résumé naïve",
			str2:     "cafe resume naive",
			expected: true,
			comment:  "混合多种法语变音符号",
		},
		{
			name:     "Mixed languages - Müller café niño",
			str1:     "Müller café niño",
			str2:     "Muller cafe nino",
			expected: true,
			comment:  "混合德语、法语、西班牙语变音符号",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnicodeCompare(tt.str1, tt.str2, Loose)
			if result != tt.expected {
				t.Errorf("UnicodeCompare(%q, %q, Loose) = %v, want %v\n说明: %s",
					tt.str1, tt.str2, result, tt.expected, tt.comment)

				// 调试输出：显示规范化后的字符串
				canonical1 := UnicodeCanonical(tt.str1, Loose)
				canonical2 := UnicodeCanonical(tt.str2, Loose)
				t.Logf("规范化后: %q -> %q, %q -> %q", tt.str1, canonical1, tt.str2, canonical2)
			}
		})
	}
}
