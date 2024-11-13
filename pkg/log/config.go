package log

import (
	"go.uber.org/zap/zapcore"
	"strings"
)

// simpleLogConf 简化版本的日志配置，用于 NewSimple
type simpleLogConf struct {
	color            bool
	production       bool
	multiLevelOutput bool
	level            zapcore.Level
	dir              string
	rotateConfig     fileConfig
}

// logConfig 完整日志配置
type logConfig struct {
	Writers []writerConfig `json:"writers"`
	Filters []filterConfig `json:"filters"`
}

type fileConfig struct {
	Path       string `json:"path,omitempty"`
	Rotate     bool   `json:"rotate"`
	MaxSize    int    `json:"max_size,omitempty"`
	MaxAge     int    `json:"max_age,omitempty"`
	MaxBackups int    `json:"max_backups,omitempty"`
	LocalTime  bool   `json:"local_time,omitempty"`
	Compress   bool   `json:"compress,omitempty"`
}

type filterConfig struct {
	Enable     bool      `json:"enable"`      // 是否启用过滤配置
	MatchRule  matchRule `json:"match_rule"`  // 匹配规则
	MatchValue bool      `json:"match_value"` // 是否匹配value
	Target     string    `json:"target"`      // 匹配关键字目标
	Replace    *string   `json:"replace"`     // 替换值
	Ignore     *bool     `json:"ignore"`      // 是否忽略此记录
}

type writerConfig struct {
	Enable     bool        `json:"enable"` // 是否启用.
	Level      string      `json:"level"`  // 日志等级
	levelValue Level       // 真正使用的值
	Output     outputType  `json:"output"`      // 输出类型
	Encoder    encoderType `json:"encoder"`     // 编码类型
	Color      bool        `json:"color"`       // 是否color化日志
	TimeFormat string      `json:"time_format"` // 时间格式，比如："2006-01-02 15:04:05.000"

	File fileConfig `json:"file"` // 文件+日志切割配置
}

// outputType 输出类型
type outputType string

const (
	OutputTypeStdout  outputType = "stdout"  // 标准输出
	OutputTypeStderr  outputType = "stderr"  // 标准错误输出
	OutputTypeFile    outputType = "file"    // 文件输出
	OutputTypeFluent  outputType = "fluent"  // fluent输出
	OutputTypeAliyun  outputType = "aliyun"  // aliyun输出
	OutputTypeTencent outputType = "tencent" // tencent输出
)

type matchRule string

const (
	MatchRuleContains = "contains" // 包含匹配
	MatchRulePrefix   = "prefix"   // 前缀匹配
	MatchRuleSuffix   = "suffix"   // 后缀匹配
)

func (rule matchRule) Match() func(a, b string) bool {
	switch rule {
	case MatchRuleContains:
		return strings.Contains
	case MatchRulePrefix:
		return strings.HasPrefix
	case MatchRuleSuffix:
		return strings.HasSuffix
	default:
		return strings.Contains
	}
}

type FilterOption struct {
	Enable     bool      `json:"enable"`      // 是否启用过滤配置
	MatchRule  matchRule `json:"match_rule"`  // 匹配规则
	MatchValue bool      `json:"match_value"` // 是否匹配value
	Target     string    `json:"target"`      // 匹配关键字目标
	Replace    *string   `json:"replace"`     // 替换值
	Ignore     *bool     `json:"ignore"`      // 是否忽略此记录
}

type encoderType string

const (
	encoderTypeConsole encoderType = "console"
	encoderTypeJSON    encoderType = "json"
)
