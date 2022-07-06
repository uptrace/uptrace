<template>
  <v-row align="end">
    <v-col
      v-for="framework in frameworks"
      :key="framework.name"
      cols="6"
      sm="3"
      lg="2"
      class="flex-grow-1"
    >
      <Devicon :name="framework.name" :icon="framework.icon" :to="framework.to" />
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Components
import Devicon from '@/components/Devicon.vue'

export default defineComponent({
  name: 'FrameworkIcons',
  components: { Devicon },

  setup() {
    const frameworks = computed(() => {
      return [
        {
          name: 'net/http',
          icon: '/devicon/net-http.svg',
          to: instrumentationLink('go-net-http.html'),
        },
        {
          name: 'Gin',
          icon: '/devicon/gin.svg',
          to: tutorialLink('gin-gorm.html'),
        },
        {
          name: 'Beego',
          icon: '/devicon/beego.svg',
          to: instrumentationLink('go-beego.html'),
        },
        {
          name: 'Django',
          icon: '/devicon/django.svg',
          to: tutorialLink('django.html'),
        },
        {
          name: 'Flask',
          icon: '/devicon/flask.svg',
          to: tutorialLink('flask.html'),
        },
        {
          name: 'FastAPI',
          icon: '/devicon/fastapi-original.svg',
          to: instrumentationLink('python-fastapi.html'),
        },
        {
          name: 'Rails',
          icon: '/devicon/rails.svg',
          to: tutorialLink('rails.html'),
        },
        {
          name: 'Express',
          icon: '/devicon/express.svg',
          to: instrumentationLink('node-express.html'),
        },
      ]
    })

    function instrumentationLink(file: string): string {
      return `https://opentelemetry.uptrace.dev/instrumentations/${file}`
    }

    function tutorialLink(name: string): string {
      return `https://get.uptrace.dev/opentelemetry/${name}`
    }

    return { frameworks }
  },
})
</script>

<style lang="scss" scoped></style>
