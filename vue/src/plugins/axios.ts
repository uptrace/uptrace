import axios from 'axios'
import axiosRetry from 'axios-retry'

// Misc
import { redirectToLogin } from '@/org/use-users'

axiosRetry(axios, { retries: 2, retryDelay: axiosRetry.exponentialDelay })

axios.interceptors.request.use((config) => {
  config.baseURL = process.env.NODE_ENV === 'production' ? '/UPTRACE_PLACEHOLDER/' : '/'
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
