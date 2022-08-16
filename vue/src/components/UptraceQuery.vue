<template>
  <div>
    <v-row no-gutters align="center" style="margin-bottom: 1px">
      <v-col>
        <div class="d-flex filters">
          <slot />
        </div>
      </v-col>

      <v-col cols="auto">
        <SpanQueryHelpDialog :uql="uql" />
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
            @keydown.enter.stop.prevent="exitRawMode(true)"
            @keydown.esc.stop.prevent="exitRawMode(false)"
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
          <v-btn text small @click="exitRawMode(false)">Cancel</v-btn>
          <v-btn small color="primary" class="ml-2" @click="exitRawMode(true)">OK</v-btn>
        </v-col>
      </v-row>
    </template>

    <v-row v-else-if="!uql.parts.length" no-gutters align="center">
      <v-col class="mb-1 px-3 text-body-2">
        <div v-if="disabled" class="text--disabled">The query is empty...</div>
        <div v-else class="text--secondary cursor-pointer" @click="uql.rawMode = true">
          Click to edit the query...
        </div>
      </v-col>
    </v-row>

    <v-row v-else no-gutters align="center">
      <v-col class="d-flex flex-wrap align-start">
        <template v-for="(part, i) in uql.parts">
          <div v-if="i === partEditor.index" :key="i" class="mr-2 d-inline-block">
            <v-text-field
              v-model="partEditor.query"
              v-autowidth="{ minWidth: '40px' }"
              :error-messages="part.error"
              outlined
              dense
              hide-details="auto"
              autofocus
              @keyup.enter.stop.prevent
              @keydown.enter.stop.prevent="partEditor.applyEdits(part)"
              @keydown.esc.stop.prevent="partEditor.cancelEdits(part)"
              @blur="partEditor.cancelEdits(part)"
            />
          </div>
          <UptraceQueryChip
            v-else
            :key="i"
            :query="part.query"
            :error="part.error"
            :disabled="part.disabled || disabled"
            class="mr-2 mb-1"
            @click:edit="partEditor.startEditing(i, part)"
            @click:delete="uql.removePart(i)"
          />
        </template>
        <v-btn
          v-if="!partEditor.editing"
          color="text--secondary"
          fab
          elevation="0"
          class="btn--add"
          @click="partEditor.add"
          ><v-icon>mdi-plus</v-icon></v-btn
        >
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, proxyRefs, PropType } from 'vue'

// Composables
import { createPart, Part, UseUql } from '@/use/uql'

// Components
import UptraceQueryChip from '@/components/UptraceQueryChip.vue'
import SpanQueryHelpDialog from '@/tracing/SpanQueryHelpDialog.vue'

export default defineComponent({
  name: 'UptraceQuery',
  components: { SpanQueryHelpDialog, UptraceQueryChip },

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

    function exitRawMode(save: boolean) {
      props.uql.rawMode = false
      if (save) {
        props.uql.query = query.value
      }
    }

    watch(
      () => props.uql.query,
      (s) => {
        query.value = s
      },
      { immediate: true },
    )

    return {
      query,

      exitRawMode,
      partEditor: usePartEditor(props.uql),
    }
  },
})

function usePartEditor(uql: UseUql) {
  const partIndex = shallowRef<number>()
  const partQuery = shallowRef('')

  const editing = computed(() => {
    return partIndex.value !== undefined
  })

  function addPart() {
    const part = createPart()
    uql.addPart(part)

    startEditing(uql.parts.length - 1, part)
  }

  function startEditing(i: number, part: Part) {
    partIndex.value = i
    partQuery.value = part.query
  }

  function applyEdits(part: Part) {
    if (partQuery.value !== part.query) {
      part.error = ''
    }
    part.query = partQuery.value

    cancelEdits()
  }

  function cancelEdits() {
    partIndex.value = undefined
    partQuery.value = ''
    uql.cleanup()
  }

  return proxyRefs({
    index: partIndex,
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

.btn--add {
  height: 32px !important;
  width: 32px !important;
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
