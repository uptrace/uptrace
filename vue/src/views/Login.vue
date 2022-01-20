<template>
  <v-container fluid class="fill-height grey lighten-5">
    <v-row>
      <v-col>
        <v-card max-width="500" class="mx-auto">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>Log in</v-toolbar-title>
          </v-toolbar>

          <v-form v-model="isValid" @submit.prevent="submit">
            <v-card flat class="px-14 py-8">
              <v-text-field
                v-model="username"
                prepend-inner-icon="mdi-account"
                label="Username"
                :rules="rules.username"
                required
                filled
              ></v-text-field>
              <v-text-field
                v-model="password"
                prepend-inner-icon="mdi-lock"
                label="Password"
                type="password"
                :rules="rules.password"
                required
                filled
              ></v-text-field>

              <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn :loading="loading" :disabled="!isValid" type="submit" color="primary">
                  Sign in
                </v-btn>
              </v-card-actions>
            </v-card>
          </v-form>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, ref, watch } from '@vue/composition-api'

// Composables
import { useTitle } from '@vueuse/core'
import { useAxios } from '@/use/axios'
import { useRouter } from '@/use/router'
import { useUser } from '@/use/org'

const requiredRule = (v: string) => (v && v.length != 0) || 'Field is required'

export default defineComponent({
  name: 'Login',

  setup() {
    useTitle('Log in')
    const { router } = useRouter()
    const user = useUser()

    const isValid = ref(false)
    const rules = {
      username: [requiredRule],
      password: [requiredRule],
    }

    const username = ref('uptrace')
    const password = ref('uptrace')

    const { loading, request } = useAxios()

    watch(
      () => user.current,
      () => {
        if (user.isAuth) {
          router.push({ name: 'Home' })
        }
      },
    )

    function submit() {
      login().then(() => {
        user.reload()
      })
    }

    function login() {
      const data = {
        username: username.value,
        password: password.value,
      }

      const url = `/api/users/login`
      return request({ method: 'POST', url, data })
    }

    return {
      rules,
      isValid,

      username,
      password,

      loading,
      submit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
