package config

import "gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"

// LoadSettings 加载配置文件
func LoadSettings[C utils.IProtobuf](confDriver string, confPath string) C {
	var driver Configure
	var settings C = utils.New[C]()

	driver = NewFromDriver(confDriver, confPath)

	if err := driver.Load(); err != nil {
		panic(err)
	}

	if err := driver.Scan(settings); err != nil {
		panic(err)
	}

	return settings
}
