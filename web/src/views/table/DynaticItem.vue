<script setup lang="ts">
import { useTimeAgo } from '@vueuse/core'
import { formatDate } from '@/helper'

interface Props {
  value?: string | number | boolean
  field?: 'id' | 'createdAt' | 'updatedAt' | string
  type?: 'number' | 'string' | 'boolean' | string
  toggleBool: Function
}

defineProps<Props>()
</script>

<template>
  <template v-if="field === 'createdAt' || field === 'updatedAt'">
    <div class="tooltip" :data-tip="formatDate(value as string)">
      <span class="text-gray cursor-default">
        {{ useTimeAgo(value as string).value }}
      </span>
    </div>
  </template>
  <template v-else-if="type === 'boolean'">
    <div @click="toggleBool()">
      <div v-if="value" class="i-uiw:circle-check-o text-green-500" />
      <div v-else class="i-uiw:circle-close-o text-red-500" />
    </div>
  </template>
  <!-- <template v-else-if="type === 'number'">
    <span class="badge badge-accent cursor-default">
      {{ value }}
    </span>
  </template> -->
  <template v-else>
    {{ value }}
  </template>
</template>
