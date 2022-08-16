<template>
  <div class="container--fixed-sm">
    <PageToolbar>
      <v-toolbar-title>OpenTelemetry Metrics and Grafana</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          <p>
            Uptrace fully supports
            <a href="https://uptrace.dev/opentelemetry/metrics.html">OpenTelemetry Metrics</a> and
            Prometheus API. That allows you to efficiently store metrics in ClickHouse database and
            view them in Grafana using Prometheus data source.
          </p>

          <p>To view Uptrace metrics in Grafana:</p>

          <ol class="mb-4">
            <li>In Grafana, go to "Configururation" -&gt; "Data source".</li>
            <li>Click on "Add data source" and then select "Prometheus".</li>
            <li>
              As an URL, use Uptrace HTTP address, for example,
              <code>{{ project.http.endpoint }}/api/prometheus</code>.
            </li>
            <li>In "Custom HTTP Headers" section, click "Add header".</li>
            <ul>
              <li>Header: <code>uptrace-dsn</code></li>
              <li>
                Value: <code>{{ project.http.dsn }}</code>
              </li>
            </ul>
          </ol>

          <p>
            You can also use
            <a href="https://github.com/uptrace/uptrace/tree/master/example/docker"
              >Docker example</a
            >
            which comes with pre-configured Grafana available on
            <code>http://localhost:3000/</code>.
          </p>

          <p>
            <v-img src="/grafana/prom-data-source.png" />
          </p>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useProject } from '@/use/project'

export default defineComponent({
  name: 'Metrics',

  setup() {
    const project = useProject()

    return { project }
  },
})
</script>

<style lang="scss" scoped></style>
