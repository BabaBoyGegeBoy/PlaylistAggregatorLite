const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,
  publicPath: './',
  // 关闭 sourcemap：避免 .map 文件被打进二进制，显著减小体积
  productionSourceMap: false,
  pages: {
    index: {
      entry: 'src/main.js',
      title: '歌单解析与聚合-PlaylistAggregator Lite'
    }
  }
})
