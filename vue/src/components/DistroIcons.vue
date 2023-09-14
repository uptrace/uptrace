<template>
  <v-row align="end">
    <v-col v-for="distro in distros" :key="distro.name" cols="6" sm="3" class="flex-grow-1">
      <DevIcon :name="distro.name" :icon="distro.icon" :href="distro.href" />
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Components
import DevIcon from '@/components/DevIcon.vue'

export default defineComponent({
  name: 'DistroIcons',
  components: { DevIcon },

  setup() {
    const distros = computed(() => {
      const items = [
        {
          name: 'Go',
          icon: 'devicon/go-original.svg',
          href: link('go'),
        },
        {
          name: 'Python',
          icon: 'devicon/python-original.svg',
          href: link('python'),
        },
        {
          name: 'Ruby',
          icon: 'devicon/ruby-original.svg',
          href: link('ruby'),
        },
        {
          name: 'Node.js',
          icon: 'devicon/nodejs-original.svg',
          href: link('js-node'),
        },
        {
          name: '.NET',
          icon: 'devicon/dot-net-original.svg',
          href: link('dotnet'),
        },
        {
          name: 'Java',
          icon: 'devicon/java-original.svg',
          href: link('java'),
        },
        {
          name: 'Erlang',
          icon: 'devicon/erlang-original.svg',
          href: link('erlang-elixir'),
        },
        {
          name: 'Elixir',
          icon: 'devicon/elixir-original.svg',
          href: link('erlang-elixir'),
        },
        {
          name: 'Rust',
          icon: 'devicon/rust-plain.svg',
          href: link('rust'),
        },
        {
          name: 'PHP',
          icon: 'devicon/php-original.svg',
          href: link('php'),
        },
      ]

      const publicPath = process.env.BASE_URL
      for (let item of items) {
        item.icon = publicPath + item.icon
      }

      return items
    })

    function link(lang: string): string {
      return `https://uptrace.dev/get/opentelemetry-${lang}.html`
    }

    return { distros }
  },
})
</script>

<style lang="scss" scoped></style>
