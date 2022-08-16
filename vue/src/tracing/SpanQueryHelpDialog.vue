<template>
  <div>
    <v-btn icon @click="dialog = true">
      <v-icon>mdi-help-circle-outline</v-icon>
    </v-btn>

    <v-dialog v-model="dialog" :max-width="1280" @keydown.esc="dialog = false">
      <v-card>
        <v-toolbar flat color="blue lighten-5">
          <v-toolbar-title>Querying</v-toolbar-title>
          <v-spacer />
          <v-btn icon @click="dialog = false"><v-icon small>mdi-close</v-icon></v-btn>
        </v-toolbar>

        <v-card-text class="py-4">
          <p class="grey--text text--darken-4">
            Uptrace supports SQL-like language to filter, group, and aggregate span
            <a href="https://uptrace.dev/opentelemetry/attributes.html" target="_blank"
              >attributes</a
            >.
          </p>

          <v-simple-table>
            <thead class="v-data-table-header">
              <tr>
                <th>Query</th>
                <th>Comment</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>
                  <UqlChip
                    :uql="uql"
                    column="span.name"
                    op="contains"
                    value="get|post"
                    @click="dialog = false"
                  />
                </td>
                <td>Filter span names that contain "get" or "post" (case-insensitive).</td>
              </tr>
              <tr>
                <td>
                  <UqlChip
                    :uql="uql"
                    query="group by span.group_id | avg(span.duration) | where span.duration &gt; 10ms and span.duration &lt; 50ms"
                    @click="dialog = false"
                  />
                </td>
                <td>Filter span duration.</td>
              </tr>
              <tr>
                <td>
                  <UqlChip
                    :uql="uql"
                    query="group by host.name | p50(span.duration)"
                    @click="dialog = false"
                  />
                </td>
                <td>Select median span duration for each host name.</td>
              </tr>
              <tr>
                <td>
                  <UqlChip
                    :uql="uql"
                    query="group by span.group_id | where span.event_count > 0"
                    @click="dialog = false"
                  />
                </td>
                <td>Select spans with events for each group.</td>
              </tr>
            </tbody>
          </v-simple-table>
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn text color="primary" @click="dialog = false">Close</v-btn>
          <v-btn text color="primary" href="https://uptrace.dev/docs/querying.html" target="_blank"
            >Read more</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, PropType } from 'vue'

// Composables
import { UseUql } from '@/use/uql'

// Components
import UqlChip from '@/components/UqlChip.vue'

export default defineComponent({
  name: 'SpanQueryHelpDialog',
  components: { UqlChip },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
  },

  setup() {
    const dialog = ref(false)
    return { dialog }
  },
})
</script>

<style lang="scss" scoped>
tr:hover td {
  background-color: #fff;
}
</style>
