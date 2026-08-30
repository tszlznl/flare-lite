// Package define 存放全局启动参数与路径解析。
package define

import (
	"os"
	"path/filepath"

	"nav/config/model"
)

// AppFlags 全局启动参数，由 server 在启动时注入，之后只读。
var AppFlags model.Flags

const DefaultConfigName = "sites.yml"

// ConfigPath 返回数据文件的绝对路径。相对路径基于当前工作目录，
// 与 flare 一致：把可执行文件和 sites.yml 放在一起即可开箱运行。
func ConfigPath() string {
	name := AppFlags.Config
	if name == "" {
		name = DefaultConfigName
	}
	if filepath.IsAbs(name) {
		return filepath.Clean(name)
	}
	root, err := os.Getwd()
	if err != nil {
		return filepath.Clean(name)
	}
	return filepath.Join(root, name)
}
