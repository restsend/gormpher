<script setup lang="ts">
import { marked } from 'marked'
import help from './help.md?raw'

const markdown = marked.parse(help)
</script>

<template>
  <article class="mx-3 text-base prose prose-truegray">
    <div v-html="markdown" />
  </article>
</template>
