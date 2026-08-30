// Package cmd 负责解析命令行参数与环境变量，产出应用启动配置。
package cmd

import (
	"flag"
	"os"
	"strconv"

	"nav/config/model"
)

const (
	_defaultPort   = "5120"
	_defaultConfig = "sites.yml"
)

// Parse 解析命令行参数，环境变量（NAV_*）优先级低于显式命令行参数。
func Parse() model.Flags {
	flags := model.Flags{
		Port:   _envOr("NAV_PORT", _defaultPort),
		Config: _envOr("NAV_CONFIG", _defaultConfig),
		Debug:  os.Getenv("NAV_DEBUG") == "1",
	}

	flag.StringVar(&flags.Port, "port", flags.Port, "监听端口，例如 5120 或 0.0.0.0:5120")
	flag.StringVar(&flags.Config, "config", flags.Config, "书签数据文件路径（YAML）")
	flag.BoolVar(&flags.Debug, "debug", flags.Debug, "调试模式：模板与样式改为从磁盘读取，改动即时生效")
	flag.Parse()

	return flags
}

func _envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Address 把 -port 的取值规范化为 echo 需要的监听地址。
// 允许传入纯端口号，也允许传入 host:port。
func Address(port string) string {
	if port == "" {
		port = _defaultPort
	}
	// 纯数字端口
	if _, err := strconv.Atoi(port); err == nil {
		return ":" + port
	}
	return port
}
