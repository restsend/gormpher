<script setup lang="ts">
import { ref } from 'vue'
import DynaticInput from './DynaticInput.vue'
import type { Filter } from '@/types'

interface Props {
  filter: Filter
  filters: string[] // filterable fields
  type: string
}

const props = defineProps<Props>()
const emits = defineEmits(['handleQuery', 'removeFilter'])

const filter = ref(props.filter)

function handleQueryFilter(filter: Filter) {
  if (filter.name && filter.value && filter.op)
    emits('handleQuery')
}

const filterOptions = ['=', '<>', 'in', 'not_in', '>', '>=', '<', '<=']
</script>

<template>
  <div class="flex space-x-4 items-center">
    <select
      v-model="filter.name"
      class="select select-sm"
      @change="handleQueryFilter(filter)"
    >
      <option v-for="option of filters" :key="option">
        {{ option }}
      </option>
    </select>
    <select
      v-model="filter.op"
      class="select select-sm"
      @change="handleQueryFilter(filter)"
    >
      <option v-for="option of filterOptions" :key="option">
        {{ option }}
      </option>
    </select>
    <DynaticInput
      v-model:value="filter.value"
      operation="filter"
      :type="type"
      :field="filter.name"
      @update:value="$emit('handleQuery')"
    />
    <span
      class="i-uiw:circle-close-o text-3xl"
      @click="$emit('removeFilter', filter.name)"
    />
  </div>
</template>
