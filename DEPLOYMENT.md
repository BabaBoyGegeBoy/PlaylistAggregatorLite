# PlaylistAggregator 部署文档

## 部署步骤

### 1. 构建后端项目

在项目根目录执行：

```bash
go build
```

> 本项目已去除 Redis / MySQL 依赖，无需启动任何数据库或缓存服务，编译后可直接独立运行。

### 2. 构建前端项目

进入前端目录并安装依赖：

```bash
cd static
npm install
```

**本地开发**：

```bash
npm run serve
```

**生产部署**：

```bash
npm run build
```

构建产物输出到 `static/dist`，由后端 `router.Static("/", "./static/dist")` 直接托管。

### 3. 配置后端请求地址

前端通过同源请求访问后端 API（`/songlist`、`/aggregate`），默认后端地址为 `http://127.0.0.1:8081`。
如需修改，编辑 `static/src/App.vue` 中的请求地址。

### 4. 访问应用

- 运行编译后的二进制文件 `./gomusic`（Windows 为 `gomusic.exe`，监听 `models.Port` 配置的端口，默认 8081）
- 浏览器访问 `http://127.0.0.1:8081/`

### 5. 使用多歌单聚合

在页面「多歌单聚合」标签页中，每行粘贴一个歌单链接，点击「聚合歌单」即可按出现次数去重、排序并导出 `.txt`。
