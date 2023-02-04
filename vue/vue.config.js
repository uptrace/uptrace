const { defineConfig } = require('@vue/cli-service')

module.exports = defineConfig({
  publicPath: process.env.NODE_ENV === 'production' ? '/UPTRACE_PLACEHOLDER/' : '/',
  transpileDependencies: ['vuetify'],
  devServer: {
    port: 19876,
    allowedHosts: 'all',
    proxy: {
      '^/api': {
        target: 'http://localhost:14318',
        changeOrigin: true,
      },
    },
  },
  configureWebpack: (config) => {
    config.resolve.fallback = {
      querystring: require.resolve('querystring-es3'),
    }
  },
  chainWebpack: (config) => {
    config.plugin('html').tap((args) => {
      args[0].title = 'Uptrace: Open Source APM'
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
})
