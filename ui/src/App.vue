<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'

const { user, login, logout } = useAuth()
</script>

<template>
  <h1>Hello App!</h1>
  <template v-if="user">
    <p>Logged in as {{ user.email }}</p>
    <button v-on:click="logout">Logout</button>
  </template>
  <template v-else>
    <p>Not logged in</p>
    <button v-on:click="login">Login</button>
  </template>
  <p><strong>Current route path:</strong> {{ $route.fullPath }}</p>
  <nav>
    <RouterLink :to="{ name: 'home' }">Go to Home</RouterLink><br />
    <RouterLink :to="{ name: 'create' }">Go to Create</RouterLink>
  </nav>
  <main>
    <Suspense>
      <RouterView />
    </Suspense>
  </main>
</template>
