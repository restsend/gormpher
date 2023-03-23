<script setup lang="ts">
import type { ActionType } from '@/types'

interface Props {
  actions?: Array<ActionType>
}

defineProps<Props>()
</script>

<template>
  <template v-if="actions?.includes('filter')">
    <div class="tooltip tooltip-bottom tooltip-info" data-tip="Filterable">
      <span class="badge badge-info badge-sm cursor-default"> F </span>
    </div>
  </template>
  <template v-if="actions?.includes('search')">
    <div class="tooltip tooltip-bottom tooltip-error" data-tip="Searchable">
      <span class="badge badge-error badge-sm cursor-default"> S </span>
    </div>
  </template>
  <!-- <template v-if="actions?.includes('order')">
    <div class="tooltip tooltip-warning" data-tip="Orderable">
      <span class="badge badge-warning badge-sm cursor-default"> O </span>
    </div>
  </template> -->
  <template v-if="actions?.includes('edit')">
    <div class="tooltip tooltip-bottom tooltip-success" data-tip="Editable">
      <span class="badge badge-success badge-sm cursor-default"> E </span>
    </div>
  </template>
</template>
