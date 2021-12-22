<template>
  <div class="container--fixed-sm">
    <PageToolbar :loading="loading">
      <v-toolbar-title>Send data to Uptrace</v-toolbar-title>

      <v-spacer />

      <v-btn v-if="dateRange" small outlined @click="dateRange.reload">
        <v-icon small>mdi-refresh</v-icon>
        <span class="ml-1">Reload</span>
      </v-btn>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-body-1">
          <p>
            To start sending data to Uptrace, you need to install Uptrace client. Use the following
            DSN to configure Uptrace client by following instructions for your programming language:
          </p>

          <XCode :code="dsn" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <DistroList />
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Already using OpenTelemetry Collector?</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-body-1">
          Uptrace natively supports OpenTelemetry Protocol (OTLP) in case you are already using
          OpenTelemetry Collector. Use the following OTLP exporter config:
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <CollectorTabs :otlp="otlp" :dsn="dsn" />
        </v-col>
      </v-row>
    </v-container>

    <HelpLinks />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useWatchAxios } from '@/use/watch-axios'

// Components
import DistroList from '@/components/DistroList.vue'
import CollectorTabs from '@/components/CollectorTabs.vue'
import HelpLinks from '@/components/HelpLinks.vue'

export default defineComponent({
  name: 'HelpCard',
  components: { DistroList, CollectorTabs, HelpLinks },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    dateRange: {
      type: Object as PropType<UseDateRange>,
      default: undefined,
    },
  },

  setup() {
    const { data } = useWatchAxios(() => {
      return {
        url: '/api/tracing/conn-info',
      }
    })

    const otlp = computed(() => {
      return data.value?.otlp ?? 'localhost:4317'
    })

    const dsn = computed(() => {
      return data.value?.dsn ?? 'http://localhost:4317'
    })

    return { otlp, dsn }
  },
})
</script>

<style lang="scss" scoped></style>
