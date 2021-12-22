import axios from 'axios'
import axiosRetry from 'axios-retry'

axiosRetry(axios, { retries: 2, retryDelay: axiosRetry.exponentialDelay })

axios.interceptors.request.use((config) => {
  config.baseURL = process.env.VUE_APP_BASE_URL
  config.withCredentials = true
  return config
})
