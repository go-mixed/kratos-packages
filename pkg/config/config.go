package config

import (
	"github.com/go-kratos/kratos/contrib/config/apollo/v2"
	"github.com/go-kratos/kratos/v2/config/env"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/config/stream"
	"os"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/config/file"
)

type Observer func(string, Value)
type Value = config.Value

const (
	// DriverFile 文件配置驱动
	DriverFile = "file"

	// DriverApollo 阿波罗配置驱动
	DriverApollo = "apollo"

	DriverStream = "stream"
)

var (
// WithSource = config.WithSource
)

// Configure 标准配置扩展
type Configure interface {
	Load() error
	Scan(v any) error
	Value(key string) config.Value
	IsDriver(driver string) bool

	String(key string, defValue ...string) string
	Bool(key string, defValue ...bool) bool
	Int(key string, defValue ...int) int
	Int8(key string, defValue ...int8) int8
	Int16(key string, defValue ...int16) int16
	Int32(key string, defValue ...int32) int32
	Int64(key string, defValue ...int64) int64
	Float32(key string, defValue ...float32) float32
	Float64(key string, defValue ...float64) float64
	Duration(key string, defValue ...time.Duration) time.Duration
	Watch(key string, o Observer) error
}

type configure struct {
	loaded  bool
	driver  string
	origin  config.Config
	sources []config.Source

	paths      []string
	streamOpts StreamOption
	apolloOpts ApolloOption

	nativeOpts []config.Option
}

// NewFromDriver 使用指定驱动实例化配置
func NewFromDriver(driver, dir string, opts ...config.Option) Configure {
	switch strings.ToLower(driver) {
	case DriverApollo:
		return New(WithDriver(DriverApollo), WithNativeOption(opts...))
	case DriverFile:
		return New(WithDriver(DriverFile), WithPath(dir), WithNativeOption(opts...))
	default:
		panic("暂不支持此类型的配置驱动")
	}
}

func Default() Configure {
	return New(WithNativeOption(config.WithSource(env.NewSource("KRATOS_"))))
}

func New(opts ...Option) Configure {
	c := &configure{
		driver: os.Getenv("CONF_DRIVER"),
		apolloOpts: ApolloOption{
			AppID:     os.Getenv("APOLLO_APP_ID"),
			Cluster:   os.Getenv("APOLLO_CLUSTER"),
			Endpoint:  os.Getenv("APOLLO_ENDPOINT"),
			Namespace: os.Getenv("APOLLO_NAMESPACE"),
			Secret:    os.Getenv("APOLLO_SECRET"),
		},
		paths: strings.Split(os.Getenv("CONF_PATH"), ";"),
	}

	if c.driver == "" {
		c.driver = DriverFile
	}

	for _, opt := range opts {
		opt(c)
	}

	switch c.driver {
	case DriverApollo:
		c.sources = append(c.sources, apollo.NewSource(
			apollo.WithEnableBackup(),
			apollo.WithAppID(c.apolloOpts.AppID),
			apollo.WithCluster(c.apolloOpts.Cluster),
			apollo.WithEndpoint(c.apolloOpts.Endpoint),
			apollo.WithNamespace(c.apolloOpts.Namespace),
			apollo.WithSecret(c.apolloOpts.Secret),
			apollo.WithOriginalConfig(),
		))
	case DriverFile:
		var sources []config.Source
		for _, p := range c.paths {
			sources = append(sources, file.NewSource(p))
		}
		c.sources = append(c.sources, sources...)
	case DriverStream:
		c.sources = append(c.sources, stream.NewSource(c.streamOpts.Reader, c.streamOpts.Pipe))
	default:
		panic("暂不支持此类型的配置驱动")
	}

	c.nativeOpts = append(c.nativeOpts, config.WithSource(c.sources...))

	c.origin = config.New(c.nativeOpts...)

	if err := c.autoLoad(); err != nil {
		panic(err)
	}

	return c
}

// IsDriver 判断当前配置驱动
func (c *configure) IsDriver(driver string) bool {
	if c.driver == driver {
		return true
	}
	return false
}

// Load 加载配置文件
func (c *configure) Load() error {
	return c.origin.Load()
}

func (c *configure) autoLoad() error {
	if !c.loaded {
		if err := c.Load(); err != nil {
			return err
		}
		c.loaded = true
	}
	return nil
}

// Scan 配置映射
func (c *configure) Scan(v any) error {
	//if err := c.autoLoad(); err != nil {
	//	return err
	//}
	//var valueKey string
	//if c.IsDriver(DriverApollo) {
	//	valueKey = "application"
	//}
	//if envValueKey, exists := os.LookupEnv("APOLLO_VALUE_KEY"); c.IsDriver(DriverApollo) && exists {
	//	valueKey = envValueKey
	//}
	//if c.IsDriver(DriverApollo) {
	//	return c.origin.Value(valueKey).Scan(v)
	//}
	//return c.origin.Scan(v)
	return c.origin.Scan(v)
}

// Value 通过key操作配置
func (c *configure) Value(key string) config.Value {
	return c.origin.Value(key)
}

// String 获取string类型配置，不存在则使用默认值
func (c *configure) String(key string, defValue ...string) string {
	var defVal string
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	target, err := c.Value(key).String()
	if err != nil {
		return defVal
	}
	return target
}

// Bool 获取boolean类型配置，不存在则使用默认值
func (c *configure) Bool(key string, defValue ...bool) bool {
	var defVal bool
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	target, err := c.Value(key).Bool()
	if err != nil {
		return defVal
	}
	return target
}

// Float32 获取float32类型配置，解析错误则使用默认值
func (c *configure) Float32(key string, defValue ...float32) float32 {
	var defVal float32
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	return float32(c.Float64(key, float64(defVal)))
}

// Float64 获取float64类型配置，解析错误则使用默认值
func (c *configure) Float64(key string, defValue ...float64) float64 {
	var defVal float64
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	target, err := c.origin.Value(key).Float()
	if err != nil {
		return defVal
	}
	return target
}

// Duration 获取time.duration类型配置，解析错误则使用默认值
func (c *configure) Duration(key string, defValue ...time.Duration) time.Duration {
	var defVal time.Duration
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	target, err := c.origin.Value(key).Duration()
	if err != nil {
		return defVal
	}
	return target
}

// Int 获取int类型配置，不存在则使用默认值
func (c *configure) Int(key string, defValue ...int) int {
	var defVal int64
	if len(defValue) > 0 {
		defVal = int64(defValue[0])
	}
	return int(c.Int64(key, defVal))
}

// Int8 获取int8类型配置，不存在则使用默认值
func (c *configure) Int8(key string, defValue ...int8) int8 {
	var defVal int64
	if len(defValue) > 0 {
		defVal = int64(defValue[0])
	}
	return int8(c.Int64(key, defVal))
}

// Int16 获取int16类型配置，不存在则使用默认值
func (c *configure) Int16(key string, defValue ...int16) int16 {
	var defVal int64
	if len(defValue) > 0 {
		defVal = int64(defValue[0])
	}
	return int16(c.Int64(key, defVal))
}

// Int32 获取int32类型配置，不存在则使用默认值
func (c *configure) Int32(key string, defValue ...int32) int32 {
	var defVal int64
	if len(defValue) > 0 {
		defVal = int64(defValue[0])
	}
	return int32(c.Int64(key, defVal))
}

// Int64 获取int64类型配置，不存在则使用默认值
func (c *configure) Int64(key string, defValue ...int64) int64 {
	var defVal int64
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	target, err := c.origin.Value(key).Int()
	if err != nil {
		return defVal
	}
	return target
}

func (c *configure) Strings(key string, defValue ...[]string) []string {
	result, defVal := []string{}, []string{}
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	target, err := c.origin.Value(key).Slice()
	if err != nil {
		return defVal
	}
	for _, item := range target {
		v, err := item.String()
		if err != nil {
			return defVal
		}
		result = append(result, v)
	}
	return result
}

func (c *configure) Int64s(key string, defValue ...[]int64) []int64 {
	result, defVal := []int64{}, []int64{}
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	target, err := c.origin.Value(key).Slice()
	if err != nil {
		return defVal
	}
	for _, item := range target {
		v, err := item.Int()
		if err != nil {
			return defVal
		}
		result = append(result, v)
	}
	return result
}

func (c *configure) Ints(key string, defValue ...[]int) []int {
	defVal, result := []int{}, []int{}
	def64Val := []int64{}
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	for _, d := range defVal {
		def64Val = append(def64Val, int64(d))
	}
	def64Val = c.Int64s(key, def64Val)
	for _, d := range def64Val {
		result = append(result, int(d))
	}
	return result
}

func (c *configure) Int8s(key string, defValue ...[]int8) []int8 {
	defVal, result := []int8{}, []int8{}
	def64Val := []int64{}
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	for _, d := range defVal {
		def64Val = append(def64Val, int64(d))
	}
	def64Val = c.Int64s(key, def64Val)
	for _, d := range def64Val {
		result = append(result, int8(d))
	}
	return result
}

func (c *configure) Int16s(key string, defValue ...[]int16) []int16 {
	defVal, result := []int16{}, []int16{}
	def64Val := []int64{}
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	for _, d := range defVal {
		def64Val = append(def64Val, int64(d))
	}
	def64Val = c.Int64s(key, def64Val)
	for _, d := range def64Val {
		result = append(result, int16(d))
	}
	return result
}

func (c *configure) Int32s(key string, defValue ...[]int32) []int32 {
	defVal, result := []int32{}, []int32{}
	def64Val := []int64{}
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	for _, d := range defVal {
		def64Val = append(def64Val, int64(d))
	}
	def64Val = c.Int64s(key, def64Val)
	for _, d := range def64Val {
		result = append(result, int32(d))
	}
	return result
}

func (c *configure) Float64s(key string, defValue ...[]float64) []float64 {
	result, defVal := []float64{}, []float64{}
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	target, err := c.origin.Value(key).Slice()
	if err != nil {
		return defVal
	}
	for _, item := range target {
		v, err := item.Float()
		if err != nil {
			return defVal
		}
		result = append(result, v)
	}
	return result
}

func (c *configure) Float32s(key string, defValue ...[]float32) []float32 {
	defVal, result := []float32{}, []float32{}
	def64Val := []float64{}
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	for _, d := range defVal {
		def64Val = append(def64Val, float64(d))
	}
	def64Val = c.Float64s(key, def64Val)
	for _, d := range def64Val {
		result = append(result, float32(d))
	}
	return result
}

func (c *configure) Watch(key string, o Observer) error {
	return c.origin.Watch(key, config.Observer(o))
}
