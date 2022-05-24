module.exports = {
  transpileDependencies: ['vuetify'],
  devServer: {
    port: 19876,
    overlay: {
      warnings: false,
      errors: true,
    },
    disableHostCheck: true,
    proxy: {
      '^/api': {
        target: 'http://localhost:14318',
        changeOrigin: true,
      },
      '^/\\d+/loki': {
        target: 'http://localhost:14318',
        changeOrigin: true,
      },
    },
  },
  chainWebpack: (config) => {
    config.plugin('html').tap((args) => {
      args[0].title = 'Distributed Tracing using OpenTelemetry and ClickHouse'
      return args
    })
  },
  css: {
    loaderOptions: {
      sass: {
        additionalData: `@import "@/styles/_variables.scss"`,
      },
      scss: {
        additionalData: `@import "@/styles/_variables.scss";`,
      },
    },
  },
}
