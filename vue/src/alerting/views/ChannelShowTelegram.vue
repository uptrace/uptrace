<template>
  <v-container fluid class="fill-height grey lighten-5">
    <v-row align="center" justify="center">
      <v-col cols="auto">
        <v-skeleton-loader v-if="!channel.data" width="600" type="card"></v-skeleton-loader>

        <v-card v-else width="600">
          <v-toolbar flat color="light-blue lighten-5">
            <v-breadcrumbs :items="breadcrumbs" divider=">" large class="pl-0"></v-breadcrumbs>
          </v-toolbar>

          <div class="pa-4">
            <NotifChannelTelegramForm
              :channel="channel.data"
              @click:close="$router.push({ name: 'NotifChannelList' })"
            />
          </div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useRoute } from '@/use/router'
import { useProject } from '@/org/use-projects'
import { useTelegramNotifChannel } from '@/alerting/use-notif-channels'

// Components
import NotifChannelTelegramForm from '@/alerting/NotifChannelTelegramForm.vue'

export default defineComponent({
  name: 'ChannelShowTelegram',
  components: { NotifChannelTelegramForm },

  setup() {
    useTitle('Telegram Channel')
    const route = useRoute()
    const project = useProject()

    const channel = useTelegramNotifChannel(() => {
      const { projectId, channelId } = route.value.params
      return {
        url: `/internal/v1/projects/${projectId}/notification-channels/telegram/${channelId}`,
      }
    })

    const breadcrumbs = computed(() => {
      const bs: any[] = []

      bs.push({
        text: project.data?.name ?? 'Project',
        to: {
          name: 'ProjectShow',
        },
        exact: true,
      })

      bs.push({
        text: 'Channels',
        to: {
          name: 'NotifChannelList',
        },
        exact: true,
      })

      bs.push({ text: 'Telegram' })

      return bs
    })

    return { channel, breadcrumbs }
  },
})
</script>

<style lang="scss" scoped></style>
