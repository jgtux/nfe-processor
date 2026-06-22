<template>
  <div>
    <h1 class="text-2xl font-bold mb-1">Unidentified NF-es</h1>
    <p class="text-gray-500 text-sm mb-6">
      XMLs whose issuer and recipient do not match any registered internal client.
    </p>

    <div v-if="loading" class="text-center py-16 text-gray-400">Loading...</div>

    <div v-else-if="errorMsg" class="bg-red-50 border border-red-300 text-red-700 rounded-lg px-4 py-3 text-sm">
      {{ errorMsg }}
    </div>

    <div v-else-if="list.length === 0" class="bg-green-50 border border-green-200 rounded-xl px-6 py-10 text-center">
      <p class="text-green-700 font-medium">No unidentified NF-es found.</p>
      <p class="text-green-600 text-sm mt-1">All notes were linked to internal clients.</p>
    </div>

    <div v-else class="bg-white rounded-xl border border-gray-200 overflow-x-auto">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 text-gray-500 text-xs uppercase tracking-wide">
          <tr>
            <th class="text-left px-5 py-3">Issuer</th>
            <th class="text-left px-5 py-3">Issuer CNPJ</th>
            <th class="text-left px-5 py-3">Recipient</th>
            <th class="text-left px-5 py-3">Recipient CNPJ</th>
            <th class="text-left px-5 py-3">Reason</th>
            <th class="text-right px-5 py-3">Issued At</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-for="nfe in list" :key="nfe.id" class="hover:bg-gray-50 transition-colors">
            <td class="px-5 py-3 font-medium text-gray-800">{{ nfe.issuer_name || '—' }}</td>
            <td class="px-5 py-3 font-mono text-xs text-gray-500">{{ formatCNPJ(nfe.issuer_cnpj) }}</td>
            <td class="px-5 py-3 text-gray-800">{{ nfe.recipient_name || '—' }}</td>
            <td class="px-5 py-3 font-mono text-xs text-gray-500">{{ formatCNPJ(nfe.recipient_cnpj) }}</td>
            <td class="px-5 py-3 text-amber-700 text-xs max-w-xs">{{ nfe.unidentified_note || nfe.error_msg || '—' }}</td>
            <td class="px-5 py-3 text-right text-gray-500 text-xs">{{ formatDate(nfe.issued_at) }}</td>
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

const list = ref<NFe[]>([])
const loading = ref(true)
const errorMsg = ref('')

function formatCNPJ(cnpj: string) {
  if (!cnpj || cnpj.length !== 14) return cnpj || '—'
  return cnpj.replace(/(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})/, '$1.$2.$3/$4-$5')
}

function formatDate(iso: string) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('pt-BR')
}

onMounted(async () => {
  try {
    list.value = await nfeApi.listUnidentified() ?? []
  } catch (err: unknown) {
    errorMsg.value = err instanceof Error ? err.message : 'Failed to load data'
  } finally {
    loading.value = false
  }
})
</script>
