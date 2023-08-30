<template>
  <v-card max-width="800" class="mx-auto my-4">
    <v-toolbar color="light-blue lighten-5" flat>
      <v-toolbar-title>Dashboard types</v-toolbar-title>
    </v-toolbar>

    <div class="pa-5 grey--text text--darken-3">
      <h3 class="mb-4 text-h5">How it works?</h3>

      <p>To visualize metrics, Uptrace uses 2 types of dashboards together:</p>

      <ul class="mb-4">
        <li>A <strong>grid</strong>-based dashboard is a classic grid of charts.</li>
        <li>
          A <strong>table</strong>-based dashboard is a table where each row leads to the grid-based
          dashboard described above, for example, a table of host names with rows leading to a grid
          dashboard filtered by the <code>host.name = ${host_value}</code>.
        </li>
      </ul>

      <p>
        In other words, table dashboards allow you to parameterize grid dashboards with attributes
        from the table. You can use tables as a replacement for Grafana variables.
      </p>

      <p>
        If you don't need a table-based dashboard, just leave it blank and use only the grid-based
        dashboard.
      </p>

      <h3 class="mb-4 text-h5">Example</h3>

      <p>
        For example, after installing
        <a href="https://uptrace.dev/opentelemetry/collector-host-metrics.html" target="_blank"
          >OpenTelemetry hostmetrics</a
        >, you can monitor infrastructure metrics by creating a table dashboard that contains a list
        of hostnames along with their CPU usage:
      </p>

      <v-img src="/metrics/dash-table-form.png" class="framed" />

      <p>
        In the accompanying grid dashboard, you can have charts with more detailed metrics such us
        CPU time for every CPU state or core:
      </p>

      <v-img src="/metrics/dash-grid-chart.png" class="framed" />

      <p>
        When viewing the table dashboard, you can click on a row to view the grid dashboard filtered
        by the <code>host.name</code> attribute:
      </p>

      <v-img src="/metrics/dash-table-row.png" class="framed" />
    </div>
  </v-card>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'DashHelp',
})
</script>

<style lang="scss" scoped>
li {
  margin-bottom: 4px;
}

.framed {
  margin-bottom: 26px;
  border: 1px solid map-get($blue-grey, 'lighten-4');
  border-radius: 8px;
}
</style>
