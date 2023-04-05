<template>
  <div>
    <ListViewsDrawer :views="views" />
    <SaveViewDrawer :views="views" :disabled="!user.isAuth || Boolean($route.params.eventSystem)" />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useUser } from '@/org/use-users'
import { useSavedViews } from '@/tracing/views/use-saved-views'

// Components
import ListViewsDrawer from '@/tracing/views/ListViewsDrawer.vue'
import SaveViewDrawer from '@/tracing/views/SaveViewDrawer.vue'

export default defineComponent({
  name: 'SavedViews',
  components: { ListViewsDrawer, SaveViewDrawer },

  setup() {
    const user = useUser()

    const views = useSavedViews(() => {
      return {}
    })

    return {
      user,
      views,
    }
  },
})
</script>

<style lang="scss" scoped></style>
