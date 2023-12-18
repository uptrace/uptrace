<template>
  <div class="uptrace-query text-truncate">
    <v-row no-gutters align="center" style="margin-bottom: 1px">
      <v-col>
        <div class="d-flex filters">
          <slot />
        </div>
      </v-col>

      <v-col cols="auto">
        <slot name="actions" />
      </v-col>
    </v-row>

    <template v-if="uql.rawMode">
      <v-row align="center" no-gutters>
        <v-col>
          <v-textarea
            v-model="query"
            rows="1"
            outlined
            clearable
            auto-grow
            autofocus
            hide-details="auto"
            spellcheck="false"
            @keyup.enter.stop.prevent
            @keydown.enter.stop.prevent="apply"
            @keydown.esc.stop.prevent="cancel"
          >
          </v-textarea>
        </v-col>
      </v-row>
      <v-row no-gutters class="mt-1">
        <v-col class="text-caption grey--text text--darken-2">
          Press ENTER to apply and ESC to cancel
        </v-col>
        <v-spacer />
        <v-col cols="auto">
          <v-btn text small @click="cancel">Cancel</v-btn>
          <v-btn small color="primary" class="ml-2" @click="apply">OK</v-btn>
        </v-col>
      </v-row>
    </template>

    <v-row v-else-if="!uql.parts.length" no-gutters align="center">
      <v-col class="mb-1 px-2 text-body-2">
        <div v-if="disabled" class="text--disabled">The query is empty...</div>
        <div v-else class="text--secondary cursor-pointer" @click="uql.rawMode = true">
          Click to edit the query...
        </div>
      </v-col>
    </v-row>

    <v-row v-else no-gutters align="center">
      <v-col class="d-flex flex-wrap align-start">
        <div v-for="part in uql.parts" :key="part.id" class="mr-2 mb-1 d-flex text-truncate">
          <v-text-field
            v-if="part.id === partEditor.partId"
            v-model="partEditor.query"
            v-autowidth="{ minWidth: '200px', maxWidth: 580 }"
            :error-messages="part.error"
            outlined
            dense
            hide-details="auto"
            autofocus
            style="max-width: 600px"
            @keyup.enter.stop.prevent
            @keydown.enter.stop.prevent="partEditor.applyEdits(part)"
            @keydown.esc.stop.prevent="partEditor.cancelEdits(part)"
            @blur="partEditor.applyEdits(part)"
          />
          <UptraceQueryChip
            v-else
            :key="part.id"
            :query="part.query"
            :error="part.error"
            :disabled="part.disabled || disabled"
            deletable
            class="mr-2 mb-1"
            @click:edit="partEditor.startEditing(part)"
            @click:delete="uql.removePart(part)"
          />
        </div>

        <v-btn v-if="!partEditor.editing" depressed small class="py-4" @click="partEditor.add">
          <v-icon left>mdi-plus</v-icon>
          <span>Add query</span>
        </v-btn>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, proxyRefs, PropType } from 'vue'

// Composables
import { createQueryPart, QueryPart, UseUql } from '@/use/uql'

// Components
import UptraceQueryChip from '@/components/UptraceQueryChip.vue'

export default defineComponent({
  name: 'UptraceQuery',
  components: { UptraceQueryChip },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const query = shallowRef('')

    watch(
      () => props.uql.query,
      (s) => {
        query.value = s
      },
      { immediate: true },
    )

    function apply() {
      exitRawMode(true)
    }

    function cancel() {
      exitRawMode(false)
    }

    function exitRawMode(save: boolean) {
      props.uql.rawMode = false
      if (save) {
        props.uql.query = query.value
      }
    }

    return {
      query,

      partEditor: usePartEditor(props.uql),
      apply,
      cancel,
    }
  },
})

function usePartEditor(uql: UseUql) {
  const partId = shallowRef<number>()
  const partQuery = shallowRef('')

  const editing = computed(() => {
    return partId.value !== undefined
  })

  function addPart() {
    const part = createQueryPart()
    uql.addPart(part)

    startEditing(part)
  }

  function startEditing(part: QueryPart) {
    partId.value = part.id
    partQuery.value = part.query
  }

  function applyEdits(part: QueryPart) {
    if (partQuery.value !== part.query) {
      part.error = ''
    }
    part.query = partQuery.value

    cancelEdits()
  }

  function cancelEdits() {
    partId.value = undefined
    partQuery.value = ''
    uql.cleanup()
  }

  return proxyRefs({
    partId: partId,
    query: partQuery,
    editing,

    add: addPart,
    startEditing,
    applyEdits,
    cancelEdits,
  })
}
</script>

<style lang="scss" scoped>
.v-chip ::v-deep .v-icon {
  font-size: 20px;
  width: 20px;
  height: 20px;
}

.uptrace-query ::v-deep .v-divider--vertical {
  min-height: 80% !important;
  height: 80% !important;
  max-height: 80% !important;
  align-self: center !important;
}
</style>

<style lang="scss">
.v-btn--filter {
  height: 34px !important;
  padding: 0 8px !important;
  color: map-get($grey, 'darken-2') !important;
  text-transform: none;
}
</style>
