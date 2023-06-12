<script setup lang="ts">
import { ref } from 'vue'
import type { Filter } from '@/types'

const props = defineProps<{
  primaryKey: string
  filter: Filter
  filters: string[] // filterable fields
  type: string
  goType: string
}>()

const emit = defineEmits(['handleQuery', 'removeFilter'])

const filter = ref(props.filter)

function handleQueryFilter(filter: Filter) {
  if (filter.name && filter.value && filter.op)
    emit('handleQuery')
}

const filterOptions = ['=', '<>', 'in', 'not_in', '>', '>=', '<', '<=']
</script>

<template>
  <div class="flex items-center space-x-4">
    <!-- name -->
    <select
      v-model="filter.name"
      class="select select-sm"
      @change="handleQueryFilter(filter)"
    >
      <option v-for="option of filters" :key="option">
        {{ option }}
      </option>
    </select>
    <!-- op -->
    <select
      v-model="filter.op"
      class="select select-sm"
      @change="handleQueryFilter(filter)"
    >
      <option v-for="option of filterOptions" :key="option">
        {{ option }}
      </option>
    </select>
    <!-- value -->
    <div class="w-full flex items-center justify-end">
      <template v-if="goType === 'time.Time' || goType === '*time.Time'">
        <input
          v-model="filter.value"
          type="datetime-local"
          class="input input-sm input-bordered w-full"
        >
      </template>
      <template v-else-if="type === 'number'">
        <input
          v-model.number="filter.value"
          type="number"
          class="input input-sm input-bordered w-full"
        >
      </template>
      <template v-else-if="type === 'boolean'">
        <input
          v-model="filter.value"
          type="checkbox"
          class="toggle toggle-success"
        >
      </template>
      <template v-else-if="type === 'string'">
        <input
          v-model="filter.value"
          type="text"
          class="input input-sm input-bordered w-full"
        >
      </template>
      <template v-else>
        <textarea
          v-model="filter.value"
          type="text"
          class="textarea"
        />
      </template>
    </div>
    <!-- X -->
    <span
      class="i-uiw:circle-close-o inline-block h-9 w-9 text-gray-300 hover:text-gray-800"
      @click="$emit('removeFilter', filter.name)"
    />
  </div>
</template>
