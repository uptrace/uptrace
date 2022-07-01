<template>
  <div>
    <v-text-field
      v-model="formattedDate"
      label="Date"
      class="d-inline-block"
      :rules="rules.date"
      @blur="onBlur"
    ></v-text-field>
    <v-text-field
      v-model="formattedTime"
      :rules="rules.time"
      label="Time"
      class="d-inline-block"
      @blur="onBlur"
    ></v-text-field>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from '@vue/composition-api'

// Utilities
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'DateRangePickerInput',

  props: {
    value: {
      type: Date,
      required: true,
    },
  },

  setup(props, ctx) {
    const menu = shallowRef(false)

    const date = shallowRef(props.value)
    const formattedDate = shallowRef(formatDate(props.value))
    const formattedTime = shallowRef(
      `${padTo2Digits(props.value.getHours())}:${padTo2Digits(props.value.getMinutes())}`,
    )

    const timeRule = (v: string) => isValidTime(v) || 'Time must be valid'
    function isValidTime(v: string): boolean {
      return /^([0-1]?[0-9]|2[0-4]):([0-5][0-9])$/.test(v)
    }
    const dateRule = (v: string) => isValidDate(v) || 'Date must be valid'
    function isValidDate(v: string) {
      return /^\d{2}-\d{2}-\d{4}$/.test(v)
    }
    const rules = {
      date: [requiredRule, dateRule],
      time: [requiredRule, timeRule],
    }

    function onBlur() {
      if (!isValidTime(formattedTime.value) || !isValidDate(formattedDate.value)) {
        return
      }

      const [dd, mm, yyyy] = formattedDate.value.split('-')
      const [hours, minutes] = formattedTime.value.split(':')
      date.value = new Date(
        parseInt(yyyy),
        parseInt(mm) - 1,
        parseInt(dd),
        parseInt(hours),
        parseInt(minutes),
      )

      ctx.emit('input', date.value)
    }

    function formatDate(val: Date) {
      return [
        padTo2Digits(val.getDate()),
        padTo2Digits(val.getMonth() + 1),
        val.getFullYear(),
      ].join('-')
    }

    function padTo2Digits(num: number) {
      return num.toString().padStart(2, '0')
    }

    return {
      menu,
      date,
      formattedDate,
      formattedTime,
      rules,

      formatDate,
      onBlur,
    }
  },
})
</script>

<style lang="scss" scoped></style>
