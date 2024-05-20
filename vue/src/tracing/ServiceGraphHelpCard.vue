<template>
  <v-card flat>
    <p>
      Service Graphs provide a visual representation of service interactions, dependencies, and
      performance metrics. Service graphs are built by analyzing span relationships and require
      specific span attributes.
    </p>

    <p>
      To have a service graph or to improve the quality of the existing service graph, consider
      adding the following attributes:
    </p>

    <ul class="mb-4">
      <li>
        <code>service.name</code> is an attribute that represents the logical name of the service.
        Required.
      </li>
      <li>
        <code>_kind</code> (<code>span_kind</code>) categorizes the type of work being represented
        by the span. Make sure your traces have <code>client/server</code> or
        <code>producer/consumer</code> span pairs. Required.
      </li>
      <li>
        <code>_status_code</code> (<code>span_status_code</code>) provides information about the
        result of a specific operation or unit of work represented by this span. This attribute
        helps to monitor and understand the success, failure, or other conditions of a given
        operation. Recommended.
      </li>
      <li>
        <code>rpc.service</code> represents the full (logical) name of the service being called.
        Optional.
      </li>
      <li>
        <code>db.name</code> contains the name of the database being accessed. Recommended if you
        have multiple databases.
      </li>
      <li>
        <code>server.socket.domain</code> represents the domain name of an immediate peer. Required
        for HTTP calls.
      </li>
      <li>
        <code>server.socket.address</code> represents the physical server IP address or Unix socket
        address. Required for HTTP calls if <code>server.socket.domain</code> is not provided.
      </li>
    </ul>
  </v-card>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'ServiceGraphHelpCard',
})
</script>

<style lang="scss" scoped>
li {
  margin-bottom: 8px !important;
}
</style>
