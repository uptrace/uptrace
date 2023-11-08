<template>
  <div class="container--fixed-sm">
    <PageToolbar :loading="loading">
      <v-toolbar-title>Send data to Uptrace</v-toolbar-title>

      <v-spacer />

      <ForceReloadBtn v-if="showReload" />
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col>
          To start sending data to Uptrace, you need to configure OpenTelemetry SDK. Use the
          following <strong>DSN</strong> to configure OpenTelemetry for your programming language:
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <OtelSdkCard :dsn="project.dsn" />
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Already using OpenTelemetry Collector?</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col>
          In case you are already using
          <a href="https://uptrace.dev/opentelemetry/collector.html" target="_blank"
            >OpenTelemetry Collector</a
          >, you can send data to Uptrace using
          <a
            href="https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter"
            target="_blank"
            >otlpexporter</a
          >.
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <CollectorTabs :dsn="project.dsn" />
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Quickstart</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row align="end">
        <v-col
          v-for="item in frameworks"
          :key="item.name"
          cols="6"
          sm="3"
          lg="2"
          class="flex-grow-1"
        >
          <DevIcon v-bind="item" />
        </v-col>
      </v-row>
    </v-container>

    <HelpLinks />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Composables
import { useProject } from '@/org/use-projects'

// Components
import ForceReloadBtn from '@/components/date/ForceReloadBtn.vue'
import OtelSdkCard from '@/components/OtelSdkCard.vue'
import CollectorTabs from '@/components/CollectorTabs.vue'
import DevIcon from '@/components/DevIcon.vue'
import HelpLinks from '@/components/HelpLinks.vue'

export default defineComponent({
  name: 'HelpCard',
  components: {
    ForceReloadBtn,
    OtelSdkCard,
    CollectorTabs,
    DevIcon,
    HelpLinks,
  },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    showReload: {
      type: Boolean,
      default: false,
    },
  },

  setup() {
    const project = useProject()

    const frameworks = computed(() => {
      const items = [
        {
          name: 'net/http',
          icon: 'devicon/net-http.svg',
          href: instrumentationLink('opentelemetry-go-net-http.html'),
        },
        {
          name: 'Gin',
          icon: 'devicon/gin.svg',
          href: instrumentationLink('opentelemetry-gin.html'),
        },
        {
          name: 'Beego',
          icon: 'devicon/beego.svg',
          href: instrumentationLink('opentelemetry-beego.html'),
        },
        {
          name: 'Django',
          icon: 'devicon/django.svg',
          href: instrumentationLink('opentelemetry-django.html'),
        },
        {
          name: 'Flask',
          icon: 'devicon/flask.svg',
          href: instrumentationLink('opentelemetry-flask.html'),
        },
        {
          name: 'FastAPI',
          icon: 'devicon/fastapi-original.svg',
          href: instrumentationLink('opentelemetry-fastapi.html'),
        },
        {
          name: 'SQLAlchemy',
          icon: 'devicon/sqlalchemy-original.svg',
          href: instrumentationLink('opentelemetry-sqlalchemy.html'),
        },
        {
          name: 'Rails',
          icon: 'devicon/rails.svg',
          href: instrumentationLink('opentelemetry-rails.html'),
        },
        {
          name: 'Express',
          icon: 'devicon/express.svg',
          href: instrumentationLink('opentelemetry-express.html'),
        },
        {
          name: 'Spring Boot',
          icon: 'devicon/spring-original.svg',
          href: instrumentationLink('opentelemetry-spring-boot.html'),
        },
        {
          name: 'Phoenix',
          icon: 'devicon/phoenix-original.svg',
          href: instrumentationLink('opentelemetry-phoenix.html'),
        },
      ]

      const publicPath = process.env.BASE_URL
      for (let item of items) {
        item.icon = publicPath + item.icon
      }

      return items
    })

    function instrumentationLink(file: string): string {
      return `https://uptrace.dev/get/instrument/${file}`
    }

    return { project, frameworks }
  },
})
</script>

<style lang="scss" scoped></style>
