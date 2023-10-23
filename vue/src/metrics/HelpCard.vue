<template>
  <div class="container--fixed-sm">
    <v-progress-linear v-if="loading" top absolute indeterminate></v-progress-linear>

    <PageToolbar>
      <v-toolbar-title>Send metrics to Uptrace</v-toolbar-title>

      <v-spacer />

      <ForceReloadBtn v-if="showReload" />
    </PageToolbar>

    <v-container fluid class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          <p>There are two types of metrics you can collect:</p>

          <ol class="mb-4">
            <li>
              <a href="#in-app">In-app metrics</a> using OpenTelemetry SDK, for example, Go HTTP
              server metrics or user-defined metrics.
            </li>
            <li>
              <a href="#infra">Infrastructure metrics</a> using OpenTelemetry Collector, for
              example, Linux/Windows system metrics or PostgreSQL metrics.
            </li>
          </ol>

          <p>
            You can check our
            <router-link :to="{ name: 'DashboardList', params: { projectId: 1 } }" target="_blank"
              >playground</router-link
            >
            to play with metrics and
            <a href="https://uptrace.dev/opentelemetry/metrics.html" target="_blank">learn</a>
            how to create your own metrics.
          </p>
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar id="in-app">
      <v-toolbar-title>In-app metrics</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          <p>
            To start sending in-app metrics to Uptrace, you need to configure OpenTelemetry metrics
            SDK. Use the following <strong>DSN</strong> to configure OpenTelemetry for your
            programming language:
          </p>

          <PrismCode :code="`export UPTRACE_DSN=&quot;${project.http.dsn}&quot;`" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <DistroIcons />
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar id="infra">
      <v-toolbar-title>Infrastructure metrics</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          <p>
            To start monitoring your infrastructure, you need to install OpenTelemetry Collector on
            each host that you want to monitor. Collector acts as an agent that pulls metrics from
            monitored systems and exports them to Uptrace using the OTLP exporter.
          </p>

          <p>Use the following <strong>DSN</strong> to configure OpenTelemetry Collector:</p>

          <PrismCode :code="`export UPTRACE_DSN=&quot;${project.http.dsn}&quot;`" />
        </v-col>
      </v-row>

      <v-row>
        <v-col class="text-center">
          <v-btn
            color="primary"
            href="https://uptrace.dev/opentelemetry/collector.html#installation"
            target="_blank"
            >Install Collector</v-btn
          >
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Supported software</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row align="end">
        <v-col
          v-for="item in receivers"
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
import DistroIcons from '@/components/DistroIcons.vue'
import DevIcon from '@/components/DevIcon.vue'
import HelpLinks from '@/components/HelpLinks.vue'

export default defineComponent({
  name: 'HelpCard',
  components: { ForceReloadBtn, DistroIcons, DevIcon, HelpLinks },

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

    const receivers = computed(() => {
      const items = [
        {
          name: 'AWS',
          icon: 'devicon/amazonwebservices-original.svg',
          href: 'https://uptrace.dev/get/ingest/aws-cloudwatch.html',
        },
        {
          name: 'PostgreSQL',
          icon: 'devicon/postgresql-original.svg',
          href: monitorLink('postgresql'),
        },
        {
          name: 'MySQL',
          icon: 'devicon/mysql-original.svg',
          href: monitorLink('mysql'),
        },
        {
          name: 'SQLServer',
          icon: 'devicon/microsoftsqlserver-original.svg',
          href: receiverLink('pulsar'),
        },
        {
          name: 'Riak',
          icon: 'devicon/riak.svg',
          href: receiverLink('riak'),
        },
        {
          name: 'Redis',
          icon: 'devicon/redis-original.svg',
          href: monitorLink('redis'),
        },
        {
          name: 'MongoDB',
          icon: 'devicon/mongodb-original.svg',
          href: receiverLink('mongodb'),
        },
        {
          name: 'Apache',
          icon: 'devicon/apache-original.svg',
          href: receiverLink('apache'),
        },
        {
          name: 'Nginx',
          icon: 'devicon/nginx-original.svg',
          href: receiverLink('nginx'),
        },
        {
          name: 'Kafka',
          icon: 'devicon/apachekafka-original.svg',
          href: receiverLink('kafkametrics'),
        },
        {
          name: 'Docker',
          icon: 'devicon/docker-original.svg',
          href: monitorLink('docker'),
        },
        {
          name: 'Kubernetes',
          icon: 'devicon/kubernetes-plain.svg',
          href: monitorLink('kubernetes'),
        },
        {
          name: 'Zookeeper',
          icon: 'devicon/devicon-original.svg',
          href: receiverLink('zookeeper'),
        },
        {
          name: 'Memcached',
          icon: 'devicon/devicon-original.svg',
          href: receiverLink('memcached'),
        },
        {
          name: 'Foundry',
          icon: 'devicon/cloud-foundry.svg',
          href: receiverLink('cloudfoundry'),
        },
        {
          name: 'CouchDB',
          icon: 'devicon/couchdb-original.svg',
          href: receiverLink('couchdb'),
        },
        {
          name: 'Elastic',
          icon: 'devicon/elastic-search.svg',
          href: receiverLink('elasticsearch'),
        },
        {
          name: 'IIS',
          icon: 'devicon/iis.svg',
          href: receiverLink('iis'),
        },
        {
          name: 'InfluxDB',
          icon: 'devicon/influxdb.svg',
          href: receiverLink('influxdb'),
        },
        {
          name: 'RabbitMQ',
          icon: 'devicon/rabbitmq.svg',
          href: receiverLink('rabbitmq'),
        },
        {
          name: 'Pulsar',
          icon: 'devicon/pulsar.svg',
          href: receiverLink('pulsar'),
        },
      ]

      const publicPath = process.env.BASE_URL
      for (let item of items) {
        item.icon = publicPath + item.icon
      }

      return items
    })

    function receiverLink(receiver: string): string {
      return `https://uptrace.dev/opentelemetry/collector-config.html?receiver=${receiver}`
    }

    function monitorLink(name: string): string {
      return `https://uptrace.dev/get/monitor/opentelemetry-${name}.html`
    }

    return { project, receivers }
  },
})
</script>

<style lang="scss" scoped></style>
