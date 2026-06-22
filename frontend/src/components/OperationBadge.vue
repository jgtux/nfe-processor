<template>
  <span class="inline-block px-2 py-0.5 rounded-full text-xs font-semibold" :class="cls">{{ label }}</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { OperationType } from '@/types'

const props = defineProps<{ operation: OperationType }>()

const cls = computed(() => ({
  purchase:     'bg-green-100 text-green-700',
  sale:         'bg-purple-100 text-purple-700',
  unidentified: 'bg-amber-100 text-amber-700'
}[props.operation]))

const label = computed(() => ({
  purchase:     'Purchase',
  sale:         'Sale',
  unidentified: 'Unidentified'
}[props.operation]))
</script>
