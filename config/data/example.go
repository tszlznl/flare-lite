package data

// _exampleYAML 是首次运行时写入 sites.yml 的示例数据，
// 内容取自参考截图中的博客收藏列表。
const _exampleYAML = `# 导航站点数据文件
# 修改后无需重启，下一次请求即生效。
#
# site   站点级展示配置
# groups 可选分组：一旦声明，links 会按 group 拆成多张表
# links  链接本体，desc 即「第一印象」

site:
  title: 連結收藏
  # footer 留空时自动显示「共 N 條連結」
  footer: ""
  open_in_new_tab: true
  show_search: true

# groups:
#   - id: linux
#     title: Linux 與自架
#   - id: life
#     title: 生活與閱讀

links:
  - name: The Lunduke Journal
    link: https://lunduke.com/
    desc: 以反主流、偏向美國另類右派的觀點檢視 Linux 世界開發者的所作所為，有許多獨特的吐槽。

  - name: 大丙的筆記 Dabinn's Note
    link: https://blog.dabinn.net
    desc: 玩車和 VR

  - name: Willie169
    link: https://willie169.github.io

  - name: 烤雞堡的筆記
    link: https://wei.dev
    desc: 討論 self-hosting 技術與雲端管理。

  - name: 閱讀前哨站
    link: https://readingoutpost.com/
    desc: 推薦好書給忙碌的你，透過閱讀成為更好的自己。

  - name: DarkRanger's Secret Area
    link: https://darkranger.no-ip.org
    desc: Fedora 用戶

  - name: 小丰子3C俱樂部
    link: https://tel3c.tw
    desc: 最新 3C 科技與電信資費解析的專業部落格，專業手機電信知識解說。

  - name: 高見龍
    link: https://kaochenlong.com
    desc: 為你自己學 Git 一書的作者。

  - name: ManateeLazyCat
    link: https://manateelazycat.github.io
    desc: Deepin 前 CTO。

  - name: 紅危的部落格
    link: https://bntw.dev/zh

  - name: TQGX's Site
    link: https://tqgx.github.io

  - name: ordinarykuma's blog
    link: https://ordinarykuma.blogspot.com
    desc: 樂於嘗試各種 Linux 發行版。

  - name: 琳的備忘手札
    link: https://琳.tw
    desc: Linux 企業運維知識。

  - name: 極客死亡計劃
    link: https://www.geedea.pro
    desc: 重視輸出非輸入的作者。透過寫週刊的方式傳遞自己的想法。

  - name: 阿波爾的博客
    link: www.zaqizaba.xyz
    desc: 一位鄉村醫生的簡單博客。

  - name: 聆音播放室
    link: https://lingyinaudio.com
    desc: 在 Linux 聽音樂的發燒友。
`
