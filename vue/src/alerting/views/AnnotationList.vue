<template>
  <div class="container--fixed-md">
    <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="py-1">
      <slot name="breadcrumbs" />
    </v-container>

    <PageToolbar :fluid="$vuetify.breakpoint.lgAndDown">
      <v-toolbar-title>Annotations</v-toolbar-title>

      <v-spacer />

      <ForceReloadBtn small />
    </PageToolbar>

    <v-container :fluid="$vuetify.breakpoint.lgAndDown">
      <v-row>
        <v-col>
          <v-btn color="primary" dark @click="dialog = true">
            <v-icon left>mdi-plus</v-icon>
            <span>Create annotation</span>
          </v-btn>
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-sheet rounded="lg" outlined class="mb-4">
            <div class="pa-4">
              <v-skeleton-loader
                v-if="!annotations.status.hasData()"
                type="table"
                height="600px"
              ></v-skeleton-loader>

              <template v-else>
                <AnnotationsTable
                  :loading="annotations.loading"
                  :order="order"
                  :annotations="annotations.items"
                  @change="annotations.reload()"
                >
                </AnnotationsTable>
              </template>
            </div>
          </v-sheet>

          <XPagination :pager="annotations.pager" />
        </v-col>
      </v-row>
    </v-container>

    <v-dialog v-if="dialog" v-model="dialog" max-width="700">
      <v-card>
        <v-toolbar flat color="light-blue lighten-5">
          <v-toolbar-title>New Annotation</v-toolbar-title>
          <v-spacer />
          <v-toolbar-items>
            <v-btn icon @click="dialog = false"><v-icon>mdi-close</v-icon></v-btn>
          </v-toolbar-items>
        </v-toolbar>

        <AnnotationForm
          :annotation="emptyAnnotation()"
          @click:save="annotations.reload"
          @click:close="dialog = false"
        />
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useSyncQueryParams } from '@/use/router'
import { injectForceReload } from '@/use/force-reload'
import { useOrder } from '@/use/order'
import { useAnnotations, emptyAnnotation } from '@/org/use-annotations'

// Components
import ForceReloadBtn from '@/components/ForceReloadBtn.vue'
import AnnotationsTable from '@/alerting/AnnotationsTable.vue'
import AnnotationForm from '@/alerting/AnnotationForm.vue'

export default defineComponent({
  name: 'AnnotationList',
  components: {
    ForceReloadBtn,
    AnnotationsTable,
    AnnotationForm,
  },

  setup() {
    useTitle('Annotations')
    const dialog = shallowRef(false)
    const forceReload = injectForceReload()
    const order = useOrder()

    const annotations = useAnnotations(() => {
      return {
        ...forceReload.params,
        ...order.axiosParams,
      }
    })

    useSyncQueryParams({
      fromQuery(queryParams) {
        queryParams.setDefault('sort_by', 'createdAt')
        queryParams.setDefault('sort_desc', true)

        order.parseQueryParams(queryParams)
      },
      toQuery() {
        return {
          ...order.queryParams(),
        }
      },
    })

    return {
      dialog,

      order,
      annotations,

      emptyAnnotation,
    }
  },
})
</script>

<style lang="scss" scoped></style>
