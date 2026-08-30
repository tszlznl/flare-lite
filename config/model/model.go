// Package model 定义应用的数据结构，与 YAML 文件一一对应。
package model

// Flags 应用启动参数。
type Flags struct {
	Port   string
	Config string
	Debug  bool
}

// Site 单条导航条目，对应参考图中的表格一行。
type Site struct {
	Name string `yaml:"name"`
	URL  string `yaml:"link"`
	Desc string `yaml:"desc,omitempty"`
	// Group 对应 Groups 中的 ID；留空时归入第一个分组（无分组时平铺展示）。
	Group string `yaml:"group,omitempty"`
}

// Group 分组，用于把一张长表拆成多张小表。
type Group struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
}

// Options 站点级展示配置。
type Options struct {
	Title          string `yaml:"title,omitempty"`
	Footer         string `yaml:"footer,omitempty"`
	OpenLinkNewTab bool   `yaml:"open_in_new_tab,omitempty"`
	ShowSearch     bool   `yaml:"show_search,omitempty"`
}

// Data sites.yml 的完整结构。
type Data struct {
	Site   Options `yaml:"site,omitempty"`
	Groups []Group `yaml:"groups,omitempty"`
	Items  []Site  `yaml:"links"`
}
