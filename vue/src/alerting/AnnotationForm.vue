<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-container fluid>
      <v-row dense>
        <v-col>
          <v-text-field
            v-model="annotation.name"
            label="Name"
            hint="Annotation name"
            outlined
            dense
            required
            :rules="rules.name"
            style="max-width: 400px"
          />
        </v-col>
      </v-row>

      <v-row dense>
        <v-col>
          <v-textarea
            v-model="annotation.description"
            label="Description"
            hint="Annotation description text in Markdown format"
            auto-grow
            outlined
            dense
            :rules="rules.description"
            style="max-width: 400px"
          />
        </v-col>
      </v-row>

      <v-row dense align="center" class="mb-3">
        <v-col cols="auto" class="text--secondary"> Color </v-col>
        <v-col>
          <v-dialog width="auto">
            <template #activator="{ on, attrs }">
              <v-chip outlined label class="ml-2" title="Change color" v-bind="attrs" v-on="on">
                <v-avatar :color="annotation.color" size="10" left></v-avatar>
                <span>{{ annotation.color }}</span>
              </v-chip>
            </template>

            <v-card>
              <v-container fluid>
                <v-row>
                  <v-col>
                    <v-color-picker
                      v-model="color"
                      show-swatches
                      swatches-max-height="100%"
                    ></v-color-picker>
                  </v-col>
                </v-row>
              </v-container>
            </v-card>
          </v-dialog>
        </v-col>
      </v-row>

      <v-row dense>
        <v-col>
          <div class="text-subtitle-1 text--secondary">Attributes</div>
          <v-card flat tile color="grey lighten-4" class="px-4 py-5">
            <v-row v-for="(attr, i) in attrs" :key="i" no-gutters>
              <v-col>
                <AnnotationAttrRow :attr="attr" required @click:remove="removeAttr(i)" />
              </v-col>
            </v-row>
            <v-row dense>
              <v-col>
                <v-btn small outlined @click="addAttr">
                  <v-icon left>mdi-plus</v-icon>
                  <span>Add attribute</span>
                </v-btn>
              </v-col>
            </v-row>
          </v-card>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="auto">
          <v-btn
            :disabled="!isValid"
            :loading="annotationMan.pending"
            type="submit"
            color="primary"
            >{{ annotation.id ? 'Save' : 'Create' }}</v-btn
          >
        </v-col>
      </v-row>
    </v-container>
  </v-form>
</template>

<script lang="ts">
import Color from 'color'
import { defineComponent, ref, shallowRef, computed, PropType } from 'vue'

// Composables
import { useAnnotationManager, Attr } from '@/org/use-annotations'

// Components
import AnnotationAttrRow from '@/alerting/AnnotationAttrRow.vue'

// Utilities
import { Annotation } from '@/org/use-annotations'
import { requiredRule, minMaxStringLengthRule } from '@/util/validation'

export default defineComponent({
  name: 'AnnotationForm',
  components: { AnnotationAttrRow },

  props: {
    annotation: {
      type: Object as PropType<Annotation>,
      required: true,
    },
  },

  setup(props, ctx) {
    const attrs = ref<Attr[]>([
      ...Object.entries(props.annotation.attrs).map(([key, value]) => ({ key, value })),
      { key: '', value: '' },
    ])
    const annotationMan = useAnnotationManager()

    const color = computed({
      get(): string {
        return Color(props.annotation.color).hex()
      },
      set(color: string) {
        props.annotation.color = color
      },
    })

    const form = shallowRef()
    const isValid = shallowRef(true)
    const rules = {
      name: [requiredRule],
      description: [minMaxStringLengthRule(0, 5000)],
    }

    function submit() {
      save().then(() => {
        ctx.emit('click:save')
        ctx.emit('click:close')
      })
    }

    function save() {
      if (!form.value.validate()) {
        return Promise.reject()
      }

      props.annotation.attrs = attrs.value
        .filter((attr) => attr.key !== '')
        .reduce((acc: Record<string, string>, attr) => {
          acc[attr.key] = attr.value
          return acc
        }, {})

      if (props.annotation.id) {
        return annotationMan.update(props.annotation)
      }
      return annotationMan.create(props.annotation)
    }

    function addAttr() {
      attrs.value.push({ key: '', value: '' })
    }

    function removeAttr(i: number) {
      attrs.value.splice(i, 1)
    }

    return {
      annotationMan,
      color,
      attrs,

      form,
      isValid,
      rules,

      submit,
      addAttr,
      removeAttr,
    }
  },
})
</script>

<style lang="scss" scoped></style>
