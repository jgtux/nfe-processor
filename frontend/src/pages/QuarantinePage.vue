<template>
  <div>
    <h1 class="text-2xl font-bold mb-1">Quarantine</h1>
    <p class="text-gray-500 text-sm mb-6">
      NF-es that failed XSD validation, mod 11 checks, or parsing.
      Records are automatically deleted after <strong>{{ ttlDays }} days</strong>.
    </p>

    <div v-if="loading" class="text-center py-16 text-gray-400">Loading...</div>

    <div v-else-if="errorMsg" class="bg-red-50 border border-red-300 text-red-700 rounded-lg px-4 py-3 text-sm">
      {{ errorMsg }}
    </div>

    <div v-else-if="list.length === 0" class="bg-green-50 border border-green-200 rounded-xl px-6 py-10 text-center">
      <p class="text-green-700 font-medium">No quarantined records.</p>
      <p class="text-green-600 text-sm mt-1">All uploaded files passed validation successfully.</p>
    </div>

    <div v-else class="bg-white rounded-xl border border-gray-200 overflow-x-auto">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 text-gray-500 text-xs uppercase tracking-wide">
          <tr>
            <th class="text-left px-5 py-3">Upload ID</th>
            <th class="text-left px-5 py-3">Reason</th>
            <th class="text-center px-5 py-3">Received</th>
            <th class="text-center px-5 py-3">Expires</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-for="nfe in list" :key="nfe.id" class="hover:bg-gray-50 transition-colors">
            <td class="px-5 py-3 font-mono text-xs text-gray-500">{{ nfe.upload_id }}</td>
            <td class="px-5 py-3 text-red-700 text-xs max-w-sm">{{ nfe.error_msg || '—' }}</td>
            <td class="px-5 py-3 text-center text-gray-500 text-xs">{{ formatDate(nfe.created_at) }}</td>
            <td class="px-5 py-3 text-center text-xs" :class="expiresClass(nfe.created_at)">
              {{ expiresIn(nfe.created_at) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { nfeApi } from '@/services/api'
import type { NFe } from '@/types'

const TTL_DAYS = 30

const list = ref<NFe[]>([])
const loading = ref(true)
const errorMsg = ref('')
const ttlDays = TTL_DAYS

function formatDate(iso: string) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString('pt-BR')
}

function expiresAt(createdAt: string): Date {
  const d = new Date(createdAt)
  d.setDate(d.getDate() + TTL_DAYS)
  return d
}

function expiresIn(createdAt: string): string {
  const now = new Date()
  const exp = expiresAt(createdAt)
  const diffDays = Math.ceil((exp.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diffDays <= 0) return 'Expires soon'
  if (diffDays === 1) return 'in 1 day'
  return `in ${diffDays} days`
}

function expiresClass(createdAt: string): string {
  const now = new Date()
  const exp = expiresAt(createdAt)
  const diffDays = Math.ceil((exp.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diffDays <= 3) return 'text-red-600 font-semibold'
  if (diffDays <= 7) return 'text-amber-600'
  return 'text-gray-500'
}

onMounted(async () => {
  try {
    list.value = await nfeApi.listQuarantine() ?? []
  } catch (err: unknown) {
    errorMsg.value = err instanceof Error ? err.message : 'Failed to load quarantine data'
  } finally {
    loading.value = false
  }
})
</script>
