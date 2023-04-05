<template>
  <div>
    <v-container v-if="!activeSpan" :fluid="fluid">
      <v-skeleton-loader type="article, table" />
    </v-container>

    <SpanCard
      v-else
      :date-range="effectiveDateRange"
      :loading="loading"
      :span="activeSpan"
      :fluid="fluid"
      :show-toolbar="showToolbar"
      @find:span="findSpan"
    >
      <template slot="append-column">
        <slot name="append-column" />
      </template>
      <template slot="action-prepend">
        <slot name="action-prepend" />
      </template>
      <template slot="append-action">
        <slot name="append-action" />
      </template>
    </SpanCard>
  </div>
</template>
