package main

import (
	// 内嵌时区数据库。缺数据时 Go 会静默退回 UTC，{date} 占位符就会在错误的时刻翻篇。
	// 代价约 450 KB，换来运行期直接 -e TZ=... 生效，无需依赖外部时区文件。
	_ "time/tzdata"

	"nav/cmd"
	"nav/internal/server"
)

func main() {
	flags := cmd.Parse()
	server.StartDaemon(&flags)
}
