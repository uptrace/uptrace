import axios from 'axios'
import axiosRetry from 'axios-retry'

// Utilities
import { redirectToLogin } from '@/org/use-users'

axiosRetry(axios, { retries: 2, retryDelay: axiosRetry.exponentialDelay })

axios.interceptors.request.use((config) => {
  config.baseURL = process.env.VUE_APP_BASE_URL
  config.withCredentials = true
  return config
})

axios.interceptors.response.use(
  (resp) => resp,
  (error) => {
    if (error.response?.status === 401) {
      redirectToLogin()
    }
    return Promise.reject(error)
  },
)
