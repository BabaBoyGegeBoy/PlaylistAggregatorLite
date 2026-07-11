# 歌单解析与聚合（PlaylistAggregator）

> 输入各平台歌单链接，一键解析并聚合为统一的「歌名 - 歌手」文本，方便你迁移到 Apple Music / YouTube Music / Spotify。

歌单解析与聚合是一款**纯本地运行**的轻量网页小工具：把网易云、QQ、汽水等平台的歌单链接，解析成统一格式的文本，支持多歌单聚合去重，并可直接用于迁移到其他音乐平台。无需部署数据库，开箱即用。

## 功能特性

- **多平台解析**：支持网易云 / QQ / 汽水 / 酷狗 / 酷我 / 咪咕 / 千千 / JOOX / bilibili / 5Sing / Apple / Jamendo 等平台歌单解析。
- **单歌单解析**：粘贴一个链接，得到按「歌名 - 歌手」排列的歌单文本。
- **多歌单聚合**：每行一个链接，自动合并并**去除重复歌曲**，显示总数、去重数与各来源状态。
- **格式与顺序可调**：歌名-歌手 / 歌手-歌名 / 仅歌名；正序 / 倒序。
- **复制 / 下载**：结果可一键复制，或下载为 `.txt`。
- **零外部依赖**：不依赖 Redis / MySQL 等任何数据库或缓存服务。
- **自适应与内置指南**：前端自适应桌面 / 平板 / 手机；右下角「?」按钮随时打开使用指南抽屉。
- **单文件分发**：前端已编译进二进制，运行只需一个可执行文件，无需额外目录。

## 支持的平台

| 平台 | 说明 |
| --- | --- |
| 网易云音乐、QQ音乐、汽水音乐 | 原生解析 |
| 酷狗音乐、酷我音乐、咪咕音乐、千千音乐、JOOX、bilibili、5Sing、Apple Music、Jamendo | 通过 [music-lib](https://github.com/guohuiyuan/music-lib) 补齐解析 |

> 注：酷狗概念版、波点音乐暂不支持解析。

## 快速开始

### 方式一：直接运行（推荐）

下载 Release 中的 `gomusic.exe`（Windows），双击即可运行，浏览器打开 <http://127.0.0.1:8081/>。

### 方式二：自行编译

前置环境：

- Go 1.25 及以上
- Node.js（仅在前端需要重新构建时需要）

```bash
# 1. 构建前端（输出到 static/dist，随后会被嵌入二进制）
cd static
npm install
# Node 17+ 需加该环境变量以兼容旧版 webpack hash 算法
export NODE_OPTIONS=--openssl-legacy-provider
npm run build
cd ..

# 2. 编译后端（前端已通过 //go:embed 嵌入二进制，无需额外目录）
go build -o gomusic.exe .

# 3. 运行
./gomusic.exe
```

默认监听端口 `8081`（在 `misc/models` 中配置），访问 <http://127.0.0.1:8081/> 即可。

## 使用指南

应用内右下角有「?」悬浮按钮，点击可随时查看图文指南。要点如下：

1. **单歌单解析**：切到「单歌单解析」页签 → 粘贴一个歌单链接 → 点击「获取歌单」→ 复制结果。
2. **多歌单聚合**：切到「多歌单聚合」页签 → 每行一个链接 → 点击「聚合歌单」→ 可复制或下载 `.txt`。
3. **格式/顺序**：按需选择歌曲格式与正序/倒序。
4. **原始歌名开关**：默认不勾选「使用未经处理的原始歌曲名」。处理后的歌名在迁移到其他平台时匹配率更高；如需原样歌名可勾选。

### 迁移到其他平台（第三方服务）

解析出的文本可直接用于迁移。本项目**推荐并引用**以下第三方免费迁移服务（均为外部网站，与本工具无隶属关系）：

- [TunemyMusic（中文版）](https://www.tunemymusic.com/zh-CN/transfer)
- [Spotlistr](https://spotlistr.com)

以 TunemyMusic 为例：打开网站 → 选择来源为「任意文本 / Any Text」→ 粘贴本工具复制的歌单文本 → 选择目的地（Apple Music / YouTube Music / Spotify 等）→ 确认并开始迁移。

## 项目结构

```
PlaylistAggregatorLite/
├── main.go              # 入口，//go:embed 嵌入 static/dist
├── handler/             # HTTP 路由与接口（/songlist、/aggregate）
├── logic/               # 各平台解析与聚合逻辑
├── misc/                # 工具、模型、日志
├── static/              # 前端 Vue 项目（构建后产物被嵌入二进制）
└── go.mod
```

## 致谢与第三方引用

- 原项目 **GoMusic**（[github.com/Bistutu/GoMusic](https://github.com/Bistutu/GoMusic)）：本仓库在其基础上改写，去除了 Redis / MySQL 依赖、精简前端并增加了多歌单聚合能力。
- [music-lib](https://github.com/guohuiyuan/music-lib)（Go 语言库，**AGPL-3.0**）：实际引入的解析核心，提供酷狗 / 酷我 / 咪咕 / 千千 / JOOX / bilibili / 5Sing / Apple / Jamendo 等平台解析能力。
- [go-music-dl](https://github.com/guohuiyuan/go-music-dl)（**AGPL-3.0**，同作者）：本项目的平台识别逻辑（DetectSource）移植自该项目；它自身依赖上述 music-lib。
- [TunemyMusic](https://www.tunemymusic.com/zh-CN/transfer) / [Spotlistr](https://spotlistr.com)：推荐的第三方歌单迁移服务。

> 许可提示：本项目以 MIT 发布，但所依赖的 music-lib / go-music-dl 均为 AGPL-3.0；将 AGPL 组件静态链接进分发的二进制可能涉及许可冲突，公开分发前请自行评估合规（详见 LICENSE 与各上游仓库许可）。

## License

基于 MIT License 发布，详见 [LICENSE](./LICENSE)。请在分发与改写时保留原作者版权声明。
