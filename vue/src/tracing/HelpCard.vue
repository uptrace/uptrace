<template>
  <div class="container--fixed-sm">
    <PageToolbar :loading="loading">
      <v-toolbar-title>Send data to Uptrace</v-toolbar-title>

      <v-spacer />

      <ForceReloadBtn v-if="showReload" />
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          <p>
            To start sending data to Uptrace, you need to configure OpenTelemetry SDK. Use the
            following <strong>DSN</strong> to configure OpenTelemetry for your programming language:
          </p>

          <p>
            For Go, Python, Java, .NET, Rust, Erlang, and Elixir, use
            <strong>OTLP/gRPC</strong> port:
          </p>

          <PrismCode :code="`export UPTRACE_DSN=&quot;${project.grpc.dsn}&quot;`" class="mb-4" />

          <p>For Ruby, Node.JS, and PHP, use <strong>OTLP/HTTP</strong> port:</p>

          <PrismCode :code="`export UPTRACE_DSN=&quot;${project.http.dsn}&quot;`" class="mb-4" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <DistroIcons />
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Already using OpenTelemetry Collector?</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          <p>
            In case you are already using OpenTelemetry
            <a href="https://uptrace.dev/opentelemetry/collector.html" target="_blank">Collector</a
            >, you can send data to Uptrace using
            <a
              href="https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter"
              target="_blank"
              >otlpexporter</a
            >.
          </p>

          <v-alert type="info" prominent border="left" outlined class="mb-0">
            Don't forget to add the Uptrace exporter to <code>service.pipelines</code> section,
            because unused exporters are silently ignored.
          </v-alert>
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <CollectorTabs :http="project.http" :grpc="project.grpc" />
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
import CollectorTabs from '@/components/CollectorTabs.vue'
import DistroIcons from '@/components/DistroIcons.vue'
import DevIcon from '@/components/DevIcon.vue'
import HelpLinks from '@/components/HelpLinks.vue'

export default defineComponent({
  name: 'HelpCard',
  components: {
    ForceReloadBtn,
    CollectorTabs,
    DistroIcons,
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
      return [
        {
          name: 'net/http',
          icon: '/devicon/net-http.svg',
          href: instrumentationLink('opentelemetry-go-net-http.html'),
        },
        {
          name: 'Gin',
          icon: '/devicon/gin.svg',
          href: instrumentationLink('opentelemetry-gin.html'),
        },
        {
          name: 'Beego',
          icon: '/devicon/beego.svg',
          href: instrumentationLink('opentelemetry-beego.html'),
        },
        {
          name: 'Django',
          icon: '/devicon/django.svg',
          href: instrumentationLink('opentelemetry-django.html'),
        },
        {
          name: 'Flask',
          icon: '/devicon/flask.svg',
          href: instrumentationLink('opentelemetry-flask.html'),
        },
        {
          name: 'FastAPI',
          icon: '/devicon/fastapi-original.svg',
          href: instrumentationLink('opentelemetry-fastapi.html'),
        },
        {
          name: 'SQLAlchemy',
          icon: '/devicon/sqlalchemy-original.svg',
          href: instrumentationLink('opentelemetry-sqlalchemy.html'),
        },
        {
          name: 'Rails',
          icon: '/devicon/rails.svg',
          href: instrumentationLink('opentelemetry-rails.html'),
        },
        {
          name: 'Express',
          icon: '/devicon/express.svg',
          href: instrumentationLink('opentelemetry-express.html'),
        },
        {
          name: 'Spring Boot',
          icon: '/devicon/spring-original.svg',
          href: instrumentationLink('opentelemetry-spring-boot.html'),
        },
        {
          name: 'Phoenix',
          icon: '/devicon/phoenix-original.svg',
          href: instrumentationLink('opentelemetry-phoenix.html'),
        },
      ]
    })

    function instrumentationLink(file: string): string {
      return `https://uptrace.dev/get/instrument/${file}`
    }

    return { project, frameworks }
  },
})
</script>

<style lang="scss" scoped></style>
