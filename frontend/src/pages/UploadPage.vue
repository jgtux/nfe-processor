<template>
  <div class="max-w-2xl mx-auto">
    <h1 class="text-2xl font-bold mb-1">Upload NF-e</h1>
    <p class="text-gray-500 mb-6 text-sm">Select one or more XML files. Processing is asynchronous.</p>

    <div
      class="border-2 border-dashed rounded-xl p-10 text-center transition-colors cursor-pointer"
      :class="isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-blue-400 hover:bg-gray-50'"
      @dragover.prevent="isDragging = true"
      @dragleave="isDragging = false"
      @drop.prevent="onDrop"
      @click="fileInput?.click()"
    >
      <svg class="w-12 h-12 mx-auto text-gray-400 mb-3" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round"
          d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5" />
      </svg>
      <p class="text-sm font-medium text-gray-700">Drag XML files here or <span class="text-blue-600 underline">click to select</span></p>
      <p class="text-xs text-gray-400 mt-1">Only NF-e .xml files</p>
      <input ref="fileInput" type="file" multiple accept=".xml" class="hidden" @change="onFileChange" />
    </div>

    <div v-if="selectedFiles.length" class="mt-4 space-y-2">
      <div
        v-for="(f, i) in selectedFiles"
        :key="i"
        class="flex items-center justify-between bg-white border border-gray-200 rounded-lg px-4 py-2 text-sm"
      >
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6M5 6h14M5 10h14" />
          </svg>
          <span class="font-medium truncate max-w-xs">{{ f.name }}</span>
          <span class="text-gray-400">({{ (f.size / 1024).toFixed(1) }} KB)</span>
        </div>
        <button class="text-gray-400 hover:text-red-500 transition-colors" @click="removeFile(i)">✕</button>
      </div>
    </div>

    <div class="mt-5 flex gap-3">
      <button
        :disabled="!selectedFiles.length || uploading"
        class="px-6 py-2 bg-blue-600 text-white rounded-lg font-medium text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-blue-700 transition-colors flex items-center gap-2"
        @click="upload"
      >
        <svg v-if="uploading" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z"/>
        </svg>
        {{ uploading ? 'Sending...' : 'Send XMLs' }}
      </button>
      <button
        v-if="selectedFiles.length"
        class="px-4 py-2 border border-gray-300 text-gray-600 rounded-lg text-sm hover:bg-gray-100 transition-colors"
        @click="clear"
      >
        Clear
      </button>
    </div>

    <div v-if="results.length" class="mt-8">
      <h2 class="text-base font-semibold mb-3">Upload results</h2>
      <div class="space-y-2">
        <div
          v-for="(r, i) in results"
          :key="i"
          class="flex items-start gap-3 rounded-lg border px-4 py-3 text-sm"
          :class="r.error ? 'border-red-200 bg-red-50' : 'border-green-200 bg-green-50'"
        >
          <span class="mt-0.5 text-base">{{ r.error ? '❌' : '✅' }}</span>
          <div>
            <p class="font-medium">{{ r.file }}</p>
            <p v-if="r.error" class="text-red-600 text-xs mt-0.5">{{ r.error }}</p>
            <p v-else class="text-green-700 text-xs mt-0.5">Queued · ID: <code>{{ r.id }}</code></p>
          </div>
        </div>
      </div>
      <p class="text-xs text-gray-400 mt-3">
        Processing is async. Check the <RouterLink to="/dashboard" class="text-blue-600 underline">Dashboard</RouterLink> in a moment.
      </p>
    </div>

    <div v-if="errorMsg" class="mt-5 bg-red-50 border border-red-300 text-red-700 rounded-lg px-4 py-3 text-sm">
      {{ errorMsg }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { nfeApi } from '@/services/api'
import type { UploadFileResult } from '@/types'

const fileInput = ref<HTMLInputElement | null>(null)
const selectedFiles = ref<File[]>([])
const isDragging = ref(false)
const uploading = ref(false)
const results = ref<UploadFileResult[]>([])
const errorMsg = ref('')

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files) return
  addFiles(Array.from(input.files))
  input.value = ''
}

function onDrop(e: DragEvent) {
  isDragging.value = false
  if (!e.dataTransfer?.files) return
  addFiles(Array.from(e.dataTransfer.files))
}

function addFiles(files: File[]) {
  const xmls = files.filter(f => f.name.endsWith('.xml'))
  const invalid = files.filter(f => !f.name.endsWith('.xml'))
  errorMsg.value = invalid.length
    ? `Ignored (not XML): ${invalid.map(f => f.name).join(', ')}`
    : ''
  const existing = new Set(selectedFiles.value.map(f => f.name))
  selectedFiles.value.push(...xmls.filter(f => !existing.has(f.name)))
}

function removeFile(i: number) { selectedFiles.value.splice(i, 1) }

function clear() {
  selectedFiles.value = []
  results.value = []
  errorMsg.value = ''
}

async function upload() {
  uploading.value = true
  results.value = []
  errorMsg.value = ''
  try {
    const res = await nfeApi.upload(selectedFiles.value)
    results.value = res.data.results
    selectedFiles.value = []
  } catch (err: unknown) {
    errorMsg.value = err instanceof Error ? err.message : 'Failed to upload files'
  } finally {
    uploading.value = false
  }
}
</script>
