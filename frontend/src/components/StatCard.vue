<template>
  <div class="bg-white rounded-xl border border-gray-200 px-5 py-4">
    <p class="text-xs text-gray-500 uppercase tracking-wide mb-1">{{ label }}</p>
    <p class="text-3xl font-bold" :class="colorClass">{{ value }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ label: string; value: number; color: 'blue' | 'green' | 'purple' }>()

const colorClass = computed(() => ({
  blue:   'text-blue-600',
  green:  'text-green-600',
  purple: 'text-purple-600'
}[props.color]))
</script>
