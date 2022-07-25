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
            <li>Open "Configururation" -&gt; "Data source" in Grafana.</li>
            <li>Click on "Add data source" and then select "Prometheus".</li>
            <li>
              As an URL, use Uptrace HTTP address, for example,
              <code>http://localhost:14318/</code>.
            </li>
            <li>
              In "Custom HTTP Headers" section, click "Add header" and add
              <code>uptrace-dsn: http://project1_secret_token@localhost:14318/1</code> header.
            </li>
          </ol>

          <p>
            You can also use
            <a href="https://github.com/uptrace/uptrace/tree/master/example/docker"
              >Docker example</a
            >
            which comes with pre-configured Grafana.
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

export default defineComponent({
  name: 'Metrics',
})
</script>

<style lang="scss" scoped></style>
