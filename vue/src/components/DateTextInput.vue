<template>
  <div>
    <v-menu
      v-model="menu"
      :close-on-content-click="false"
      transition="slide-x-transition"
      offset-y
      min-width="auto"
    >
      <template #activator="{ on, attrs }">
        <v-icon v-bind="attrs" v-on="on">mdi-calendar</v-icon>
        <v-text-field
          v-model="formattedDate"
          style="width: 115px"
          dense
          outlined
          hint="dd-mm-yyyy"
          persistent-hint
          :rules="rules.date"
          class="d-inline-block ml-2"
          @blur="onBlur"
        >
        </v-text-field>
      </template>
      <v-date-picker v-model="datePicker" no-title @input="menu = false"></v-date-picker>
    </v-menu>

    <v-text-field
      v-model="formattedTime"
      style="width: 75px"
      outlined
      dense
      hint="24 hours"
      persistent-hint
      :rules="rules.time"
      class="d-inline-block ml-4"
      @blur="onBlur"
    ></v-text-field>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch } from 'vue'

// Utilities
import { parse, format, isValid } from 'date-fns'
import { toLocal } from '@/util/date'

const DATE_FORMAT = 'dd-MM-yyyy'
const TIME_FORMAT = 'H:mm'

export default defineComponent({
  name: 'DateTextInput',

  props: {
    value: {
      type: Date,
      required: true,
    },
  },

  setup(props, { emit }) {
    const menu = shallowRef(false)
    const formattedDate = shallowRef('')
    const formattedTime = shallowRef('')

    const datePicker = computed({
      get(): string {
        return toLocal(props.value).toISOString().substr(0, 10)
      },
      set(s: string) {
        formattedDate.value = format(new Date(s), DATE_FORMAT)

        const date = parseDateTime(`${formattedDate.value} ${formattedTime.value}`)
        emit('input', date)
      },
    })

    watch(
      () => props.value,
      (date: Date) => {
        formattedDate.value = format(date, DATE_FORMAT)
        formattedTime.value = format(date, TIME_FORMAT)
      },
      { immediate: true },
    )

    const rules = {
      time: [(v: string) => isValidDateTime(v, TIME_FORMAT) || '24 hours'],
      date: [(v: string) => isValidDateTime(v, DATE_FORMAT) || 'dd-mm-yyyy'],
    }

    function isValidDateTime(s: string, format: string): boolean {
      return isValid(parse(s, format, new Date()))
    }

    function onBlur() {
      const date = parseDateTime(`${formattedDate.value} ${formattedTime.value}`)
      if (!isValid(date)) {
        return
      }

      emit('input', date)
    }

    return {
      menu,
      datePicker,
      formattedDate,
      formattedTime,
      rules,

      onBlur,
    }
  },
})

function parseDateTime(s: string) {
  return parse(s, `${DATE_FORMAT} ${TIME_FORMAT}`, new Date())
}
</script>

<style lang="scss" scoped></style>
