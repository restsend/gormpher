<script setup lang="ts">
import { computed } from 'vue'
import { formatDate } from '@/helper'

const props = defineProps<{
  operation: 'add' | 'edit'
  value?: any
  field?: 'id' | 'createdAt' | 'updatedAt' | string
  type?: 'number' | 'string' | 'boolean' | string
  goType: string
  disabled?: boolean
  primaryKey?: string
}>()

const emit = defineEmits(['update:value'])

const value = computed({
  get: () => props.value,
  set: val => emit('update:value', val),
})
</script>

<template>
  <div class="w-full flex items-center justify-end">
    <template v-if="field === primaryKey">
      {{ value || 'auto fill' }}
    </template>
    <template v-else-if="goType === 'time.Time' || goType === '*time.Time'">
      <template v-if="disabled">
        {{ value ? formatDate(value as string) : 'can not edit' }}
      </template>
      <template v-else>
        <input
          v-model="value"
          class="input input-sm input-bordered w-full"
          type="datetime-local"
          :disabled="disabled"
        >
      </template>
    </template>
    <template v-else-if="type === 'number'">
      <input
        v-model.number="value"
        :disabled="disabled"
        type="number"
        class="input input-sm input-bordered w-full"
      >
    </template>
    <template v-else-if="type === 'boolean'">
      <input
        v-model="value"
        :disabled="disabled"
        type="checkbox"
        class="toggle toggle-success"
      >
    </template>
    <template v-else-if="type === 'string'">
      <input
        v-model="value"
        :disabled="disabled"
        type="text"
        class="input input-sm input-bordered w-full"
      >
    </template>
    <template v-else>
      <input
        v-model="value"
        :disabled="disabled"
        type="text"
        class="input input-sm input-bordered w-full"
      >
      <!-- <textarea
        v-model="value"
        :disabled="disabled"
        type="text"
        class="textarea"
      /> -->
    </template>
  </div>
</template>
