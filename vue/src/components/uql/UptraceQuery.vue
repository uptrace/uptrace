<template>
  <div>
    <v-row no-gutters align="center" class="mb-n1">
      <v-col>
        <div class="d-flex filters">
          <slot />
        </div>
      </v-col>

      <v-col cols="auto">
        <SpanQueryHelpDialog :uql="uql" class="mr-2" />
      </v-col>
      <v-col cols="auto">
        <v-btn v-if="uql.rawMode" icon title="Apply filters" @click="exitRawMode(true)">
          <v-icon size="26" color="success">mdi-check</v-icon>
        </v-btn>
        <v-btn icon title="Cancel editing" @click="uql.rawMode = !uql.rawMode">
          <v-icon v-if="uql.rawMode" size="22">mdi-pencil-off-outline</v-icon>
          <v-icon v-else size="22">mdi-pencil-outline</v-icon>
        </v-btn>
      </v-col>
    </v-row>

    <v-row align="center" dense>
      <v-col v-if="uql.rawMode">
        <v-textarea
          v-model="query"
          hint="Press ENTER to apply filters and ESC to cancel editing"
          permanent-hint
          rows="1"
          outlined
          clearable
          auto-grow
          autofocus
          @keyup.enter.stop.prevent
          @keydown.enter.stop.prevent="exitRawMode(true)"
          @keydown.esc.stop.prevent="exitRawMode(false)"
        >
        </v-textarea>
      </v-col>
      <v-col v-else-if="!uql.parts.length" class="mb-1 px-3 text-body-2 grey--text text--darken-2">
        Empty query...
      </v-col>
      <v-col v-else class="d-flex flex-wrap align-start">
        <template v-for="(part, i) in uql.parts">
          <div v-if="!part.editMode" :key="part.query" class="mr-2 mb-1 d-inline-block">
            <UptraceQueryChip
              :key="i"
              :query="part.query"
              :error="part.error"
              :disabled="part.disabled || disabled"
              class="mr-2 mb-1"
              @click:edit="uql.enterEditMode(part)"
              @click:delete="uql.removeAt(i)"
            />
          </div>
          <div v-else :key="part.query" class="mr-2 d-inline-block">
            <v-text-field
              v-model="part.editQuery"
              :error-messages="part.error"
              outlined
              dense
              hide-details="auto"
              autofocus
              @keyup.enter.stop.prevent
              @keydown.enter.stop.prevent="uql.exitEditMode(part, true)"
              @keydown.esc.stop.prevent="uql.exitEditMode(part, false)"
              @blur="uql.exitEditMode(part, true)"
            />
          </div>
        </template>
        <v-btn
          v-if="!uql.editing"
          color="text--secondary"
          fab
          elevation="0"
          class="btn--add"
          @click="uql.addPart"
          ><v-icon>mdi-plus</v-icon></v-btn
        >
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from '@vue/composition-api'

// Composables
import { UseUql } from '@/use/uql'

// Components
import SpanQueryHelpDialog from '@/components/SpanQueryHelpDialog.vue'
import UptraceQueryChip from '@/components/UptraceQueryChip.vue'

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
    const editor = props.uql.createEditor()

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
      editor,

      exitRawMode,
    }
  },
})
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
  padding: 0 10px !important;
  color: map-get($grey, 'darken-2') !important;
  text-transform: none;
}
</style>
