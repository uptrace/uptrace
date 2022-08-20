<template>
  <v-row align="end">
    <v-col v-for="item in items" :key="item.name" cols="6" sm="3" lg="2" class="flex-grow-1">
      <Devicon :name="item.name" :icon="item.icon" :to="item.to" />
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Composables
import { useRouter } from '@/use/router'

// Components
import Devicon from '@/components/Devicon.vue'

export default defineComponent({
  name: 'SoftwareIcons',
  components: { Devicon },

  props: {
    showFrameworks: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const items = computed(() => {
      const items = []

      if (props.showFrameworks) {
        items.push(
          {
            name: 'net/http',
            icon: '/devicon/net-http.svg',
            to: instrumentationLink('go-net-http.html'),
          },
          {
            name: 'Gin',
            icon: '/devicon/gin.svg',
            to: instrumentationLink('go-gin.html'),
          },
          {
            name: 'Beego',
            icon: '/devicon/beego.svg',
            to: instrumentationLink('go-beego.html'),
          },
          {
            name: 'Django',
            icon: '/devicon/django.svg',
            to: instrumentationLink('python-django.html'),
          },
          {
            name: 'Flask',
            icon: '/devicon/flask.svg',
            to: instrumentationLink('python-flask.html'),
          },
          {
            name: 'FastAPI',
            icon: '/devicon/fastapi-original.svg',
            to: instrumentationLink('python-fastapi.html'),
          },
          {
            name: 'Rails',
            icon: '/devicon/rails.svg',
            to: instrumentationLink('ruby-rails.html'),
          },
          {
            name: 'Express',
            icon: '/devicon/express.svg',
            to: instrumentationLink('node-express.html'),
          },
        )
      }
      items.push(
        {
          name: 'PostgreSQL',
          icon: '/devicon/postgresql-original.svg',
          to: receiverLink('postgresql'),
        },
        {
          name: 'MySQL',
          icon: '/devicon/mysql-original.svg',
          to: receiverLink('mysql'),
        },
        {
          name: 'Redis',
          icon: '/devicon/redis-original.svg',
          to: receiverLink('redis'),
        },
        {
          name: 'MongoDB',
          icon: '/devicon/mongodb-original.svg',
          to: receiverLink('mongodb'),
        },
        {
          name: 'Nginx',
          icon: '/devicon/nginx-original.svg',
          to: receiverLink('nginx'),
        },
        {
          name: 'Kafka',
          icon: '/devicon/apachekafka-original.svg',
          to: receiverLink('kafkametrics'),
        },
        {
          name: 'Docker',
          icon: '/devicon/docker-original.svg',
          to: receiverLink('dockerstats'),
        },
        {
          name: 'Zookeeper',
          icon: '/devicon/devicon-original.svg',
          to: receiverLink('zookeeper'),
        },
        {
          name: 'Memcached',
          icon: '/devicon/devicon-original.svg',
          to: receiverLink('memcached'),
        },
      )

      const projectId = route.value.params.projectId
      if (projectId) {
        items.push({
          name: 'Slack',
          icon: '/devicon/slack-original.svg',
          to: `/projects/${projectId}/notifications/slack`,
        })

        items.push({
          name: 'PagerDuty',
          icon: '/devicon/pagerduty-original.svg',
          to: `/projects/${projectId}/notifications/pagerduty`,
        })
      }

      return items
    })

    function instrumentationLink(file: string): string {
      return `https://uptrace.dev/opentelemetry/instrumentations/${file}`
    }

    function receiverLink(receiver: string): string {
      return `https://uptrace.dev/opentelemetry/collector-config.html?receiver=${receiver}`
    }

    return { items }
  },
})
</script>

<style lang="scss" scoped></style>
