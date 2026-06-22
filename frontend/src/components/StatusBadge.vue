<template>
  <span class="inline-block px-2 py-0.5 rounded-full text-xs font-semibold" :class="cls">{{ label }}</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProcessingStatus } from '@/types'

const props = defineProps<{ status: ProcessingStatus }>()

const cls = computed(() => ({
  pending:   'bg-yellow-100 text-yellow-700',
  processed: 'bg-blue-100 text-blue-700',
  error:     'bg-red-100 text-red-700'
}[props.status]))

const label = computed(() => ({
  pending:   'Pending',
  processed: 'Processed',
  error:     'Error'
}[props.status]))
</script>
