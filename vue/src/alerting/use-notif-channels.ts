import { computed, proxyRefs } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useAxios } from '@/use/axios'
import { useWatchAxios, AxiosRequestSource, AxiosParamsSource } from '@/use/watch-axios'
import { AttrMatcher } from '@/use/attr-matcher'

export enum NotifChannelType {
  Slack = 'slack',
  Webhook = 'webhook',
  Alertmanager = 'alertmanager',
}

export enum NotifChannelState {
  Delivering = 'delivering',
  Paused = 'paused',
  Disabled = 'disabled',
}

interface BaseNotifChannel {
  id: number
  orgId: number

  name: string
  state: NotifChannelState

  type: NotifChannelType
  params: Record<string, any>
}

export type NotifChannel = SlackNotifChannel | WebhookNotifChannel

export interface SlackNotifChannel extends BaseNotifChannel {
  type: NotifChannelType.Slack
  params: SlackNotifChannelParams
}

export interface SlackNotifChannelParams {
  webhookUrl: string
}

export interface WebhookNotifChannel extends BaseNotifChannel {
  type: NotifChannelType.Webhook | NotifChannelType.Alertmanager
  params: WebhookNotifChannelParams
}

export interface WebhookNotifChannelParams {
  url: string
  payload: string
}

export type UseNotifChannels = ReturnType<typeof useNotifChannels>

export function useNotifChannels(axiosParamsSource: AxiosParamsSource) {
  const route = useRoute()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    const req = {
      url: `/api/v1/projects/${projectId}/notification-channels`,
      params: axiosParamsSource(),
    }
    return req
  })

  const channels = computed((): NotifChannel[] => {
    return data.value?.channels ?? []
  })

  return proxyRefs({
    status,
    loading,
    items: channels,

    reload,
  })
}

export function emptySlackNotifChannel(): SlackNotifChannel {
  return {
    id: 0,
    orgId: 0,
    state: NotifChannelState.Delivering,
    name: '',

    type: NotifChannelType.Slack,
    params: {
      webhookUrl: '',
    },
  }
}

export function emptyWebhookNotifChannel(): WebhookNotifChannel {
  return {
    id: 0,
    orgId: 0,
    state: NotifChannelState.Delivering,
    name: '',

    type: NotifChannelType.Webhook,
    params: {
      url: '',
      payload: '',
    },
  }
}

export function emptyAlertmanagerNotifChannel(): WebhookNotifChannel {
  return {
    id: 0,
    orgId: 0,
    state: NotifChannelState.Delivering,
    name: '',

    type: NotifChannelType.Alertmanager,
    params: {
      url: '',
      payload: '',
    },
  }
}

//------------------------------------------------------------------------------

export function useSlackNotifChannel(axiosReqSource: AxiosRequestSource) {
  const { status, loading, data, reload } = useWatchAxios(axiosReqSource)

  const channel = computed((): SlackNotifChannel | undefined => {
    return data.value?.channel
  })

  return proxyRefs({
    status,
    loading,
    reload,

    data: channel,
  })
}

export function useWebhookNotifChannel(axiosReqSource: AxiosRequestSource) {
  const { status, loading, data, reload } = useWatchAxios(axiosReqSource)

  const channel = computed((): WebhookNotifChannel | undefined => {
    return data.value?.channel
  })

  return proxyRefs({
    status,
    loading,
    reload,

    data: channel,
  })
}

export function useAlertmanagerNotifChannel(axiosReqSource: AxiosRequestSource) {
  const { status, loading, data, reload } = useWatchAxios(axiosReqSource)

  const channel = computed((): WebhookNotifChannel | undefined => {
    return data.value?.channel
  })

  return proxyRefs({
    status,
    loading,
    reload,

    data: channel,
  })
}

//------------------------------------------------------------------------------

export function useNotifChannelManager() {
  const route = useRoute()
  const { loading: pending, request } = useAxios()

  function pause(channelId: number) {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/notification-channels/${channelId}/paused`
    return request({ method: 'PUT', url })
  }

  function unpause(channelId: number) {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/notification-channels/${channelId}/unpaused`
    return request({ method: 'PUT', url })
  }

  function del(channelId: number) {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/notification-channels/${channelId}`
    return request({ method: 'DELETE', url })
  }

  //------------------------------------------------------------------------------

  function slackCreate(channel: SlackNotifChannel) {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/notification-channels/slack`
    return request({ method: 'POST', url, data: channel }).then((resp) => {
      return resp.data.channel as SlackNotifChannel
    })
  }

  function slackUpdate(channel: SlackNotifChannel) {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/notification-channels/slack/${channel.id}`
    return request({ method: 'PUT', url, data: channel }).then((resp) => {
      return resp.data.channel as SlackNotifChannel
    })
  }

  //------------------------------------------------------------------------------

  function webhookCreate(channel: WebhookNotifChannel) {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/notification-channels/webhook`
    return request({ method: 'POST', url, data: channel }).then((resp) => {
      return resp.data.channel as WebhookNotifChannel
    })
  }

  function webhookUpdate(channel: WebhookNotifChannel) {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/notification-channels/webhook/${channel.id}`
    return request({ method: 'PUT', url, data: channel }).then((resp) => {
      return resp.data.channel as WebhookNotifChannel
    })
  }

  //------------------------------------------------------------------------------

  function emailUpdate(channel: EmailNotifChannel) {
    const errorMatchers = channel.errorMatchers.filter((m) => m.attr && m.value)
    const data = {
      ...channel,
      errorMatchers,
    }

    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/notification-channels/email`
    return request({ method: 'PUT', url, data })
  }

  return proxyRefs({
    pending,

    pause,
    unpause,
    delete: del,

    slackCreate,
    slackUpdate,

    webhookCreate,
    webhookUpdate,

    emailUpdate,
  })
}

//------------------------------------------------------------------------------

export interface EmailNotifChannel {
  notifyOnMetrics: boolean
  notifyOnNewErrors: boolean
  notifyOnRecurringErrors: boolean
  errorMatchers: AttrMatcher[]
}

export function useEmailChannel() {
  const route = useRoute()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/projects/${projectId}/notification-channels/email`,
    }
  })

  const channel = computed((): EmailNotifChannel | undefined => {
    const channel = data.value?.channel
    if (!channel) {
      return channel
    }
    channel.errorMatchers ??= []
    return channel
  })

  return proxyRefs({
    status,
    loading,

    channel,

    reload,
  })
}
