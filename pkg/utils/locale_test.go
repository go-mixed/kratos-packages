package utils

import (
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
