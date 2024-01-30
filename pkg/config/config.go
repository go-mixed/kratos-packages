package config

import (
	"github.com/go-kratos/kratos/contrib/config/apollo/v2"
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/config"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/config/file"
)

const (
	// DriverFile 文件配置驱动
	DriverFile = "file"

	// DriverApollo 阿波罗配置驱动
	DriverApollo = "apollo"
)

var (
	WithSource = config.WithSource
)

// Configure 标准配置扩展
type Configure interface {
	Load() error
	Scan(v interface{}) error
	Value(key string) config.Value
	Bool(key string, defValue ...bool) bool
	Int64(key string, defValue ...int64) int64
	IsDriver(driver string) bool
}

type configure struct {
	loaded bool
	driver string
	origin config.Config
}

// NewFromDriver 使用指定驱动实例化配置
func NewFromDriver(driver, path string, opts ...config.Option) Configure {
	switch strings.ToLower(driver) {
	case DriverApollo:
		return NewWithApolloSource(opts...)
	case DriverFile:
		return NewWithFileSource(path, opts...)
	default:
		panic("暂不支持此类型的配置驱动")
	}
}

// NewWithApolloSource 使用阿波罗驱动实例化配置
func NewWithApolloSource(opts ...config.Option) Configure {
	opts = append(opts, WithSource(apollo.NewSource(
		apollo.WithEnableBackup(),
		apollo.WithAppID(os.Getenv("APOLLO_APP_ID")),
		apollo.WithCluster(os.Getenv("APOLLO_CLUSTER")),
		apollo.WithEndpoint(os.Getenv("APOLLO_ENDPOINT")),
		apollo.WithNamespace(os.Getenv("APOLLO_NAMESPACE")),
		apollo.WithSecret(os.Getenv("APOLLO_SECRET")),
	)))
	return &configure{
		driver: DriverApollo,
		origin: config.New(opts...),
	}
}

// NewWithFileSource 使用文件驱动实例化配置
func NewWithFileSource(dir string, opts ...config.Option) Configure {
	opts = append(opts, WithSource(file.NewSource(dir)))
	return &configure{
		origin: config.New(opts...),
	}
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
func (c *configure) Scan(v interface{}) error {
	if err := c.autoLoad(); err != nil {
		return err
	}
	var valueKey string
	if c.IsDriver(DriverApollo) {
		valueKey = "application"
	}
	if envValueKey, exists := os.LookupEnv("APOLLO_VALUE_KEY"); c.IsDriver(DriverApollo) && exists {
		valueKey = envValueKey
	}
	if c.IsDriver(DriverApollo) {
		return c.origin.Value(valueKey).Scan(v)
	}
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

// Int64 获取int64类型配置，不存在则使用默认值
func (c *configure) Int64(key string, defValue ...int64) int64 {
	var defVal int64
	if len(defValue) > 0 {
		defVal = defValue[0]
	}
	target, err := c.Value(key).Int()
	if err != nil {
		return defVal
	}
	return target
}
