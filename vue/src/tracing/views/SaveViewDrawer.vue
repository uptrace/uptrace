<template>
  <div class="d-inline-block">
    <v-btn small plain :disabled="disabled" class="ml-2" @click="drawer = true">
      <v-icon left>mdi-plus</v-icon>
      <span>Save view</span>
    </v-btn>

    <v-navigation-drawer
      v-model="drawer"
      v-click-outside="{
        handler: onClickOutside,
        closeConditional,
      }"
      app
      temporary
      stateless
      width="500"
    >
      <v-container fluid class="py-6">
        <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="saveView">
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="name"
                label="New view"
                outlined
                dense
                hide-details="auto"
                autofocus
                :rules="rules.name"
                required
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense justify="end">
            <v-col cols="auto">
              <v-btn small text @click="drawer = false">Cancel</v-btn>
              <v-btn
                :loading="viewMan.pending"
                :disabled="!isValid"
                type="submit"
                small
                class="primary"
                >Save</v-btn
              >
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-divider />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <SavedViewsList
                :loading="views.loading"
                :items="views.items"
                :views="views"
                @click:item="drawer = false"
              />
            </v-col>
          </v-row>
        </v-form>
      </v-container>
    </v-navigation-drawer>
  </div>
</template>

<script lang="ts">
import { omit } from 'lodash-es'
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useSavedViewManager, UseSavedViews } from '@/tracing/views/use-saved-views'

// Components
import SavedViewsList from '@/tracing/views/SavedViewsList.vue'

// Utilities
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'SaveViewDrawer',
  components: { SavedViewsList },

  props: {
    views: {
      type: Object as PropType<UseSavedViews>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const route = useRoute()
    const drawer = shallowRef(false)

    const form = shallowRef()
    const isValid = shallowRef(true)
    const name = shallowRef('')
    const rules = {
      name: [requiredRule],
    }
    const viewMan = useSavedViewManager()

    function saveView() {
      if (!form.value.validate()) {
        return
      }

      viewMan
        .save({
          name: name.value,
          route: route.value.name as string,
          params: route.value.params,
          query: omit(route.value.query, 'time_gte'),
        })
        .then(() => {
          name.value = ''
          form.value.reset()
          props.views.reload()
        })
    }

    function onClickOutside() {
      drawer.value = false
    }

    function closeConditional() {
      return drawer.value
    }

    return {
      form,
      isValid,
      name,
      rules,
      viewMan,
      saveView,

      drawer,
      onClickOutside,
      closeConditional,
    }
  },
})
</script>

<style lang="scss" scoped></style>
