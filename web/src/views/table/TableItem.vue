<script setup lang="ts">
import { useTimeAgo } from '@vueuse/core'
import { formatDate } from '@/helper'

defineProps<{
  value?: string | number | boolean
  field?: 'id' | 'createdAt' | 'updatedAt' | string
  type?: 'number' | 'string' | 'boolean' | string
  goType: string
  primaryKey?: string
  toggleBool: Function
  openModal: Function
}>()
</script>

<template>
  <template v-if="field === primaryKey">
    <button class="text-teal-600 hover:text-blue-600" @click="openModal()">
      {{ value }}
    </button>
  </template>
  <template v-else-if="goType === 'time.Time' || goType === '*time.Time'">
    <div class="tooltip" :data-tip="useTimeAgo(value as string).value">
      <span class="cursor-default text-sm text-gray-500">
        {{ formatDate(value as string) }}
      </span>
    </div>
  </template>
  <template v-else-if="type === 'boolean'">
    <button @click="toggleBool()">
      <div v-if="value" class="i-uiw:circle-check-o text-green-500" />
      <div v-else class="i-uiw:circle-close-o text-red-500" />
    </button>
  </template>
  <template v-else-if="type === 'string'">
    {{ value }}
  </template>
  <template v-else-if="type === 'number'">
    {{ value }}
  </template>
  <!-- TODO:  -->
  <template v-else>
    <span class="cursor-default">
      {{ value }}
    </span>
  </template>
</template>
