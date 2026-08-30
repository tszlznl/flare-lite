// Package embedres 把模板与静态资源打进二进制，开箱即用无需外部文件。
// 调试模式下会改为直接读取磁盘上的同名文件，改样式不用重新编译。
package embedres

import "embed"

//go:embed templates assets
var FS embed.FS
