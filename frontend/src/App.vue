<template>
  <div class="min-h-screen flex flex-col">
    <header class="bg-blue-700 text-white shadow-md">
      <div class="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <svg class="w-7 h-7" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round"
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <span class="text-lg font-semibold tracking-wide">NF-e Processor</span>
        </div>
        <nav class="flex gap-1">
          <RouterLink
            v-for="item in nav"
            :key="item.to"
            :to="item.to"
            class="px-3 py-1.5 rounded text-sm font-medium transition-colors"
            :class="$route.path === item.to
              ? 'bg-white text-blue-700'
              : 'text-blue-100 hover:bg-blue-600'"
          >
            {{ item.label }}
          </RouterLink>
        </nav>
      </div>
    </header>

    <main class="flex-1 max-w-6xl mx-auto w-full px-4 py-8">
      <RouterView />
    </main>

    <footer class="text-center text-xs text-gray-400 py-4">
      NF-e Processor · Go + Gin · PostgreSQL · RabbitMQ · Vue 3
    </footer>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router'

const $route = useRoute()

const nav = [
  { to: '/',             label: 'Upload' },
  { to: '/dashboard',    label: 'Dashboard' },
  { to: '/unidentified', label: 'Unidentified' },
  { to: '/quarantine',   label: 'Quarantine' }
]
</script>
