<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">Dashboard</h1>
      <button class="text-sm text-blue-600 hover:underline flex items-center gap-1" :disabled="loading" @click="load">
        <svg class="w-4 h-4" :class="loading ? 'animate-spin' : ''" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Refresh
      </button>
    </div>

    <div class="grid grid-cols-3 gap-4 mb-8">
      <StatCard label="Total NF-es" :value="nfes.length" color="blue" />
      <StatCard label="Purchases" :value="nfes.filter(n => n.operation === 'purchase').length" color="green" />
      <StatCard label="Sales" :value="nfes.filter(n => n.operation === 'sale').length" color="purple" />
    </div>

    <section class="bg-white rounded-xl border border-gray-200 mb-8">
      <div class="px-5 py-4 border-b border-gray-100">
        <h2 class="font-semibold text-gray-800">Client Summary</h2>
      </div>
      <div v-if="summary.length === 0" class="px-5 py-8 text-center text-gray-400 text-sm">No data available yet.</div>
      <table v-else class="w-full text-sm">
        <thead class="bg-gray-50 text-gray-500 text-xs uppercase tracking-wide">
          <tr>
            <th class="text-left px-5 py-3">Client</th>
            <th class="text-center px-5 py-3">Purchases</th>
            <th class="text-center px-5 py-3">Sales</th>
            <th class="text-center px-5 py-3">Total</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-for="row in summary" :key="row.client" class="hover:bg-gray-50 transition-colors">
            <td class="px-5 py-3 font-medium text-gray-800">{{ row.client }}</td>
            <td class="px-5 py-3 text-center">
              <span class="inline-block bg-green-100 text-green-700 px-2 py-0.5 rounded-full font-semibold">{{ row.purchases }}</span>
            </td>
            <td class="px-5 py-3 text-center">
              <span class="inline-block bg-purple-100 text-purple-700 px-2 py-0.5 rounded-full font-semibold">{{ row.sales }}</span>
            </td>
            <td class="px-5 py-3 text-center text-gray-600">{{ row.purchases + row.sales }}</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="bg-white rounded-xl border border-gray-200">
      <div class="px-5 py-4 border-b border-gray-100">
        <h2 class="font-semibold text-gray-800">All NF-es</h2>
      </div>
      <div v-if="nfes.length === 0" class="px-5 py-8 text-center text-gray-400 text-sm">
        No NF-es processed yet. <RouterLink to="/" class="text-blue-600 underline">Upload</RouterLink> an XML file.
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-gray-50 text-gray-500 text-xs uppercase tracking-wide">
            <tr>
              <th class="text-left px-4 py-3">Access Key</th>
              <th class="text-left px-4 py-3">Issuer</th>
              <th class="text-left px-4 py-3">Recipient</th>
              <th class="text-right px-4 py-3">Amount</th>
              <th class="text-center px-4 py-3">Operation</th>
              <th class="text-center px-4 py-3">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-for="nfe in nfes" :key="nfe.id" class="hover:bg-gray-50 transition-colors">
              <td class="px-4 py-3 font-mono text-xs text-gray-500 truncate max-w-[160px]">{{ nfe.access_key }}</td>
              <td class="px-4 py-3 text-gray-800 truncate max-w-[140px]">{{ nfe.issuer_name || '—' }}</td>
              <td class="px-4 py-3 text-gray-800 truncate max-w-[140px]">{{ nfe.recipient_name || '—' }}</td>
              <td class="px-4 py-3 text-right text-gray-700">{{ formatCurrency(nfe.total_amount) }}</td>
              <td class="px-4 py-3 text-center"><OperationBadge :operation="nfe.operation" /></td>
              <td class="px-4 py-3 text-center"><StatusBadge :status="nfe.status" /></td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <p v-if="errorMsg" class="mt-4 text-sm text-red-600">{{ errorMsg }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { nfeApi } from '@/services/api'
import type { NFe, ClientSummary } from '@/types'
import StatCard from '@/components/StatCard.vue'
import OperationBadge from '@/components/OperationBadge.vue'
import StatusBadge from '@/components/StatusBadge.vue'

const nfes = ref<NFe[]>([])
const summary = ref<ClientSummary[]>([])
const loading = ref(false)
const errorMsg = ref('')

async function load() {
  loading.value = true
  errorMsg.value = ''
  try {
    const [n, s] = await Promise.all([nfeApi.listAll(), nfeApi.clientSummary()])
    nfes.value = n ?? []
    summary.value = s ?? []
  } catch (err: unknown) {
    errorMsg.value = err instanceof Error ? err.message : 'Failed to load data'
  } finally {
    loading.value = false
  }
}

function formatCurrency(val: number) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val)
}

onMounted(load)
</script>
