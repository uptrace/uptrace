<template>
  <v-row align="end">
    <v-col v-for="distro in distros" :key="distro.name" cols="6" sm="3" lg="2" class="flex-grow-1">
      <Devicon :name="distro.name" :icon="distro.icon" :to="distro.to" />
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, computed } from '@vue/composition-api'

// Components
import Devicon from '@/components/Devicon.vue'

export default defineComponent({
  name: 'UptraceDistroIcons',
  components: { Devicon },

  setup() {
    const distros = computed(() => {
      return [
        {
          name: 'Go',
          icon: '/devicon/go-original.svg',
          to: link('go'),
        },
        {
          name: 'Python',
          icon: '/devicon/python-original.svg',
          to: link('python'),
        },
        {
          name: 'Ruby',
          icon: '/devicon/ruby-original.svg',
          to: link('ruby'),
        },
        {
          name: 'Node.js',
          icon: '/devicon/nodejs-original.svg',
          to: link('js-node'),
        },
        {
          name: '.NET',
          icon: '/devicon/dot-net-original.svg',
          to: link('dotnet'),
        },
        {
          name: 'Java',
          icon: '/devicon/java-original.svg',
          to: link('java'),
        },
      ]
    })

    function link(lang: string): string {
      return `https://docs.uptrace.dev/guide/${lang}.html`
    }

    return { distros }
  },
})
</script>

<style lang="scss" scoped></style>
