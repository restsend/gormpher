<script setup lang="ts">
import { ref, watch } from 'vue'
import { formatDate } from '@/helper'

interface Props {
  operation: 'filter' | 'add' | 'edit'
  value?: string | number | boolean
  field?: 'id' | 'createdAt' | 'updatedAt' | string
  type?: 'number' | 'string' | 'boolean' | string
  disabled?: boolean
}

const props = defineProps<Props>()
defineEmits(['update:value'])

const value = ref(props.value)
watch(() => props.value, val => value.value = val)
</script>

<template>
  <!-- TODO: -->
  <div class="w-full flex items-center justify-end">
    <template v-if="field === 'id'">
      <input
        type="text"
        class="input input-sm text-gray-800"
        :value="value || 'auto fill'"
        disabled
      >
    </template>
    <template v-else-if="field === 'createdAt' || field === 'updatedAt'">
      <template v-if="operation === 'filter'">
        <!-- TODO: DatePicker -->
        <input
          v-model="value"
          type="text"
          class="input input-sm input-bordered w-full"
          @change="$emit('update:value', value)"
        >
      </template>
      <template v-else>
        <input
          type="text"
          class="input input-sm text-gray-800"
          :value="value ? formatDate(value as string) : 'auto fill'"
          disabled
        >
      </template>
    </template>
    <template v-else-if="type === 'number'">
      <input
        v-model.number="value"
        :disabled="operation === 'edit' && disabled"
        type="number"
        class="input input-sm input-bordered w-full"
        @change="$emit('update:value', value)"
      >
    </template>
    <template v-else-if="type === 'boolean'">
      <input
        v-model="value"
        :disabled="operation === 'edit' && disabled"
        type="checkbox"
        class="toggle toggle-success"
        @change="$emit('update:value', value)"
      >
    </template>
    <template v-else-if="type === 'string'">
      <input
        v-model="value"
        :disabled="operation === 'edit' && disabled"
        type="text"
        class="input input-sm input-bordered w-full"
        @change="$emit('update:value', value)"
      >
    </template>
    <template v-else>
      <input
        v-model="value"
        :disabled="operation === 'edit' && disabled"
        type="text"
        class="input input-sm input-bordered w-full"
        @change="$emit('update:value', value)"
      >
    </template>
  </div>
</template>
