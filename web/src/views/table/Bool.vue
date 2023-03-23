<script setup lang="ts">
interface Props {
  value: boolean
}

defineProps<Props>()
</script>

<template>
  <template v-if="value">
    <div class="i-uiw:circle-check-o text-green-500" />
  </template>
  <template v-else>
    <div class="i-uiw:circle-close-o text-red-500" />
  </template>
</template>
