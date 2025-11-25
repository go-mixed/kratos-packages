package utils

import (
	"reflect"
	"slices"
	"sort"
	"testing"
)

// TestMaxMatchReplace 测试最大匹配替换函数
func TestMaxMatchReplace(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		replaceMap map[string]string
		want       string
	}{
		{
			name: "基本替换-单个术语",
			text: "hello world",
			replaceMap: map[string]string{
				"hello": "你好",
			},
			want: "你好 world",
		},
		{
			name: "基本替换-多个术语",
			text: "hello world",
			replaceMap: map[string]string{
				"hello": "你好",
				"world": "世界",
			},
			want: "你好 世界",
		},
		{
			name: "最大匹配原则-优先匹配长术语",
			text: "abcd",
			replaceMap: map[string]string{
				"ab":   "X",
				"abc":  "Y",
				"abcd": "Z",
			},
			want: "Z",
		},
		{
			name: "最大匹配原则-部分重叠",
			text: "abcdef",
			replaceMap: map[string]string{
				"abc": "X",
				"def": "Y",
			},
			want: "XY",
		},
		{
			name: "最大匹配原则-贪婪匹配",
			text: "aaaa",
			replaceMap: map[string]string{
				"a":  "1",
				"aa": "2",
			},
			want: "22",
		},
		{
			name: "中文字符处理",
			text: "我爱北京天安门",
			replaceMap: map[string]string{
				"北京":  "上海",
				"天安门": "外滩",
			},
			want: "我爱上海外滩",
		},
		{
			name: "中文最大匹配",
			text: "中国人民银行",
			replaceMap: map[string]string{
				"中国":   "A",
				"中国人":  "B",
				"人民":   "C",
				"银行":   "D",
				"人民银行": "E",
			},
			want: "B民D", // 正向最大匹配: "中国人"→B + "民"(保留) + "银行"→D
		},
		{
			name: "混合中英文",
			text: "hello中国world",
			replaceMap: map[string]string{
				"hello": "你好",
				"中国":    "China",
				"world": "世界",
			},
			want: "你好China世界",
		},
		{
			name: "空文本",
			text: "",
			replaceMap: map[string]string{
				"test": "TEST",
			},
			want: "",
		},
		{
			name:       "空替换表",
			text:       "hello world",
			replaceMap: map[string]string{},
			want:       "hello world",
		},
		{
			name:       "nil替换表",
			text:       "hello world",
			replaceMap: nil,
			want:       "hello world",
		},
		{
			name: "无匹配项",
			text: "hello world",
			replaceMap: map[string]string{
				"foo": "bar",
			},
			want: "hello world",
		},
		{
			name: "连续匹配",
			text: "ababab",
			replaceMap: map[string]string{
				"ab": "X",
			},
			want: "XXX",
		},
		{
			name: "单字符替换",
			text: "abc",
			replaceMap: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
			},
			want: "123",
		},
		{
			name: "替换为空字符串",
			text: "hello world",
			replaceMap: map[string]string{
				"hello": "",
				"world": "",
			},
			want: " ",
		},
		{
			name: "替换为更长的字符串",
			text: "a b c",
			replaceMap: map[string]string{
				"a": "AAA",
				"b": "BBB",
				"c": "CCC",
			},
			want: "AAA BBB CCC",
		},
		{
			name: "特殊字符",
			text: "hello@world#test",
			replaceMap: map[string]string{
				"@": " at ",
				"#": " hash ",
			},
			want: "hello at world hash test",
		},
		{
			name: "emoji字符",
			text: "hello😀world",
			replaceMap: map[string]string{
				"😀": "🎉",
			},
			want: "hello🎉world",
		},
		{
			name: "重复文本",
			text: "testtest",
			replaceMap: map[string]string{
				"testtest": "ONCE",
				"test":     "TWICE",
			},
			want: "ONCE",
		},
		{
			name: "边界情况-只有一个字符",
			text: "a",
			replaceMap: map[string]string{
				"a": "b",
			},
			want: "b",
		},
		{
			name: "部分匹配不替换",
			text: "abcdefgh",
			replaceMap: map[string]string{
				"xyz": "123",
			},
			want: "abcdefgh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaxMatchReplace(tt.text, tt.replaceMap)
			if got != tt.want {
				t.Errorf("MaxMatchReplace() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMaxMatchExtract 测试正向最大匹配提取函数
func TestMaxMatchExtract(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		terms []string
		want  []string
	}{
		{
			name:  "基本提取-单个术语",
			text:  "hello world",
			terms: []string{"hello"},
			want:  []string{"hello"},
		},
		{
			name:  "基本提取-多个术语",
			text:  "hello world",
			terms: []string{"hello", "world"},
			want:  []string{"hello", "world"},
		},
		{
			name:  "最大匹配原则-优先匹配长术语",
			text:  "abcd",
			terms: []string{"ab", "abc", "abcd"},
			want:  []string{"abcd"},
		},
		{
			name:  "最大匹配原则-贪婪匹配",
			text:  "aaaa",
			terms: []string{"a", "aa", "aaa"},
			want:  []string{"aaa", "a"}, // 匹配"aaa"+"a"
		},
		{
			name:  "中文分词-基本",
			text:  "我爱北京天安门",
			terms: []string{"我", "爱", "北京", "天安门"},
			want:  []string{"我", "爱", "北京", "天安门"},
		},
		{
			name:  "中文分词-最大匹配",
			text:  "中国人民银行",
			terms: []string{"中国", "中国人", "人民", "银行", "人民银行"},
			want:  []string{"中国人", "银行"}, // 正向最大匹配: "中国人"(9字节) + 跳过"民" + "银行"(6字节)
		},
		{
			name:  "中文分词-复杂情况",
			text:  "南京市长江大桥",
			terms: []string{"南京", "南京市", "市长", "长江", "大桥", "长江大桥"},
			want:  []string{"南京市", "长江大桥"},
		},
		{
			name:  "混合中英文",
			text:  "hello中国world",
			terms: []string{"hello", "中国", "world"},
			want:  []string{"hello", "中国", "world"},
		},
		{
			name:  "空文本",
			text:  "",
			terms: []string{"test"},
			want:  nil,
		},
		{
			name:  "空术语列表",
			text:  "hello world",
			terms: []string{},
			want:  nil,
		},
		{
			name:  "nil术语列表",
			text:  "hello world",
			terms: nil,
			want:  nil,
		},
		{
			name:  "无匹配项",
			text:  "hello world",
			terms: []string{"foo", "bar"},
			want:  nil,
		},
		{
			name:  "部分匹配",
			text:  "abcdefgh",
			terms: []string{"abc", "def", "xyz"},
			want:  []string{"abc", "def"},
		},
		{
			name:  "重复术语",
			text:  "test test test",
			terms: []string{"test"},
			want:  []string{"test"}, // 去重后只返回一次
		},
		{
			name:  "重叠术语-只取第一个",
			text:  "abcde",
			terms: []string{"abc", "bcd", "cde"},
			want:  []string{"abc"}, // 正向最大匹配，abc匹配后，bcd被跳过
		},
		{
			name:  "单字符术语",
			text:  "abc",
			terms: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "长术语优先",
			text:  "programming",
			terms: []string{"pro", "program", "programming"},
			want:  []string{"programming"},
		},
		{
			name:  "特殊字符",
			text:  "hello@world#test",
			terms: []string{"@", "#", "hello", "world", "test"},
			want:  []string{"hello", "@", "world", "#", "test"},
		},
		{
			name:  "emoji字符",
			text:  "hello😀world🎉",
			terms: []string{"😀", "🎉", "hello", "world"},
			want:  []string{"hello", "😀", "world", "🎉"},
		},
		{
			name:  "连续相同字符",
			text:  "aaaa",
			terms: []string{"aa"},
			want:  []string{"aa"},
		},
		{
			name:  "数字和字母混合",
			text:  "abc123def456",
			terms: []string{"abc", "123", "def", "456"},
			want:  []string{"abc", "123", "def", "456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaxMatchExtract(tt.text, tt.terms)

			// 由于返回的是map的keys，顺序不确定，需要排序后比较
			sort.Strings(got)
			want := tt.want
			if want != nil {
				sort.Strings(want)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("MaxMatchExtract() = %v, want %v", got, want)
			}
		})
	}
}

// TestBiMaxMatchExtract 测试双向最大匹配提取函数
func TestBiMaxMatchExtract(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		terms []string
		want  []string
	}{
		{
			name:  "基本提取-单个术语",
			text:  "hello world",
			terms: []string{"hello"},
			want:  []string{"hello"},
		},
		{
			name:  "基本提取-多个术语",
			text:  "hello world",
			terms: []string{"hello", "world"},
			want:  []string{"hello", "world"},
		},
		{
			name:  "双向匹配-正向反向结果一致",
			text:  "abcd",
			terms: []string{"ab", "cd"},
			want:  []string{"ab", "cd"},
		},
		{
			name:  "双向匹配-正向反向结果不一致",
			text:  "abc",
			terms: []string{"ab", "bc"},
			want:  []string{"ab", "bc"}, // 正向匹配ab，反向匹配bc，取并集
		},
		{
			name:  "中文分词-歧义消解",
			text:  "南京市长江大桥",
			terms: []string{"南京", "南京市", "市长", "长江", "大桥", "长江大桥"},
			want:  []string{"南京市", "长江大桥", "南京", "市长"}, // 正向和反向结果的并集
		},
		{
			name:  "中文分词-复杂歧义",
			text:  "中国人民银行",
			terms: []string{"中国", "中国人", "人民", "银行", "人民银行"},
			want:  []string{"中国人", "民银行", "中国", "人民银行"}, // 取决于正向和反向的结果
		},
		{
			name:  "中文分词-基本",
			text:  "我爱北京天安门",
			terms: []string{"我", "爱", "北京", "天安门"},
			want:  []string{"我", "爱", "北京", "天安门"},
		},
		{
			name:  "混合中英文",
			text:  "hello中国world",
			terms: []string{"hello", "中国", "world"},
			want:  []string{"hello", "中国", "world"},
		},
		{
			name:  "空文本",
			text:  "",
			terms: []string{"test"},
			want:  nil,
		},
		{
			name:  "空术语列表",
			text:  "hello world",
			terms: []string{},
			want:  nil,
		},
		{
			name:  "nil术语列表",
			text:  "hello world",
			terms: nil,
			want:  nil,
		},
		{
			name:  "无匹配项",
			text:  "hello world",
			terms: []string{"foo", "bar"},
			want:  nil,
		},
		{
			name:  "重复术语去重",
			text:  "test test",
			terms: []string{"test"},
			want:  []string{"test"},
		},
		{
			name:  "单字符术语",
			text:  "abc",
			terms: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "长术语优先",
			text:  "programming",
			terms: []string{"pro", "program", "programming"},
			want:  []string{"programming"},
		},
		{
			name:  "正向反向不同结果-简单",
			text:  "abcd",
			terms: []string{"abc", "bcd"},
			want:  []string{"abc", "bcd"}, // 正向abc+d，反向a+bcd，并集是abc和bcd
		},
		{
			name:  "正向反向不同结果-复杂",
			text:  "abcdef",
			terms: []string{"abc", "bcd", "def"},
			want:  []string{"abc", "def", "bcd"}, // 正向abc+def，反向abc+def或bcd+其他
		},
		{
			name:  "特殊字符",
			text:  "hello@world#test",
			terms: []string{"@", "#", "hello", "world", "test"},
			want:  []string{"hello", "@", "world", "#", "test"},
		},
		{
			name:  "emoji字符",
			text:  "hello😀world🎉",
			terms: []string{"😀", "🎉", "hello", "world"},
			want:  []string{"hello", "😀", "world", "🎉"},
		},
		{
			name:  "连续重复字符",
			text:  "aaaa",
			terms: []string{"a", "aa", "aaa"},
			want:  []string{"aaa", "a"}, // 正向aaa+a，反向a+aaa，并集
		},
		{
			name:  "对称文本",
			text:  "abccba",
			terms: []string{"abc", "cba"},
			want:  []string{"abc", "cba"}, // 正向abc，反向cba
		},
		{
			name:  "数字和字母混合",
			text:  "abc123def",
			terms: []string{"abc", "123", "def", "abc123", "123def"},
			want:  []string{"abc123", "def", "abc", "123def"}, // 取决于正向和反向
		},
		{
			name:  "边界情况-单字符文本",
			text:  "a",
			terms: []string{"a"},
			want:  []string{"a"},
		},
		{
			name:  "边界情况-术语长度大于文本",
			text:  "ab",
			terms: []string{"abc", "abcd"},
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BiMaxMatchExtract(tt.text, tt.terms)

			// 由于返回的是去重后的切片，顺序不确定，需要排序后比较
			sort.Strings(got)
			want := tt.want
			if want != nil {
				sort.Strings(want)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("BiMaxMatchExtract() = %v, want %v", got, want)
			}
		})
	}
}

// BenchmarkMaxMatchReplace 性能测试
func BenchmarkMaxMatchReplace(b *testing.B) {
	text := "中国人民银行和中国工商银行是中国最大的银行"
	replaceMap := map[string]string{
		"中国":   "CN",
		"人民":   "People",
		"银行":   "Bank",
		"工商":   "ICBC",
		"人民银行": "PBC",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MaxMatchReplace(text, replaceMap)
	}
}

// BenchmarkMaxMatchExtract 性能测试
func BenchmarkMaxMatchExtract(b *testing.B) {
	text := "中国人民银行和中国工商银行是中国最大的银行"
	terms := []string{"中国", "人民", "银行", "工商", "人民银行", "工商银行", "中国人"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MaxMatchExtract(text, terms)
	}
}

// BenchmarkBiMaxMatchExtract 性能测试
func BenchmarkBiMaxMatchExtract(b *testing.B) {
	text := "中国人民银行和中国工商银行是中国最大的银行"
	terms := []string{"中国", "人民", "银行", "工商", "人民银行", "工商银行", "中国人"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BiMaxMatchExtract(text, terms)
	}
}

// TestMaxMatchExtractWithOptions 测试带 Unicode 选项的 MaxMatchExtract
func TestMaxMatchExtractWithOptions(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		terms   []string
		options []UnicodeNormalizeOption
		want    []string
	}{
		{
			name:    "IgnoreCase - 大小写不敏感",
			text:    "Hello World",
			terms:   []string{"hello", "world"},
			options: []UnicodeNormalizeOption{IgnoreCase},
			want:    []string{"hello", "world"},
		},
		{
			name:    "无选项 - 严格匹配",
			text:    "Hello World",
			terms:   []string{"hello", "world"},
			options: nil,
			want:    nil,
		},
		{
			name:    "IgnoreCase + IgnoreDiacritics - 忽略变音符号",
			text:    "café résumé",
			terms:   []string{"cafe", "resume"},
			options: []UnicodeNormalizeOption{IgnoreCase, IgnoreDiacritics},
			want:    []string{"cafe", "resume"},
		},
		{
			name:    "IgnoreCase + IgnoreWidth - 忽略全角半角",
			text:    "ＨＥＬＬＯ ＷＯＲＬＤ",
			terms:   []string{"hello", "world"},
			options: []UnicodeNormalizeOption{IgnoreCase, IgnoreWidth},
			want:    []string{"hello", "world"},
		},
		{
			name:    "Loose - 宽松比较",
			text:    "Straße ＣＡＦÉ",
			terms:   []string{"strasse", "cafe"},
			options: []UnicodeNormalizeOption{Loose},
			want:    []string{"strasse", "cafe"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaxMatchExtract(tt.text, tt.terms, tt.options...)
			// 排序以便比较
			slices.Sort(result)
			slices.Sort(tt.want)
			if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("MaxMatchExtract() = %v, want %v", result, tt.want)
			}
		})
	}
}
