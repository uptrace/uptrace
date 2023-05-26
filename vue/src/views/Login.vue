<template>
  <v-container fluid class="fill-height grey lighten-5">
    <v-row>
      <v-col>
        <v-card max-width="500" class="mx-auto">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>Log in</v-toolbar-title>

            <v-spacer />

            <v-btn
              href="https://uptrace.dev/get/config.html#managing-users"
              target="_blank"
              outlined
              small
              >Create new user</v-btn
            >
          </v-toolbar>

          <template v-if="sso.methods.length">
            <v-card flat class="pa-8">
              <v-btn
                v-for="sso in sso.methods"
                :key="sso.id"
                :loading="loading"
                :href="sso.url"
                color="red darken-3"
                dark
                large
                width="100%"
              >
                {{ sso.name || 'OpenID Connect' }}
              </v-btn>
            </v-card>

            <div class="d-flex align-center">
              <v-divider />
              <div class="mx-2 grey--text text--lighten-1">or</div>
              <v-divider />
            </div>
          </template>

          <v-form v-model="isValid" @submit.prevent="submit">
            <v-card flat class="pa-8">
              <v-alert v-if="error" type="error">{{ error }}</v-alert>

              <p>
                The default login is <code>uptrace@localhost</code> with pass <code>uptrace</code>.
              </p>

              <!-- Basic Login (email/password) -->
              <v-text-field
                v-model="email"
                prepend-inner-icon="mdi-account"
                label="Email"
                :rules="rules.email"
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
import { upperFirst } from 'lodash-es'
import { defineComponent, shallowRef, watch } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useAxios } from '@/use/axios'
import { useRouter } from '@/use/router'
import { useUser, useSso } from '@/org/use-users'

const requiredRule = (v: string) => (v && v.length != 0) || 'Field is required'

export default defineComponent({
  name: 'Login',

  setup() {
    useTitle('Log in')

    const { router } = useRouter()
    const user = useUser()
    const sso = useSso()

    const isValid = shallowRef(false)
    const rules = {
      email: [requiredRule],
      password: [requiredRule],
    }
    const error = shallowRef('')

    const email = shallowRef('')
    const password = shallowRef('')

    const { loading, request } = useAxios()

    watch(
      () => user.current,
      () => {
        if (user.isAuth) {
          router.push({ name: 'Home' }).catch(() => {})
        }
      },
    )

    function submit() {
      login()
        .then(() => {
          error.value = ''
          user.reload()
        })
        .catch((err) => {
          const msg = err.response?.data?.message
          if (msg) {
            error.value = upperFirst(msg)
          }
        })
    }

    function login() {
      const data = {
        email: email.value,
        password: password.value,
      }

      const url = `/api/v1/users/login`
      return request({ method: 'POST', url, data })
    }

    return {
      sso,

      isValid,
      rules,
      error,

      email,
      password,

      loading,
      submit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
