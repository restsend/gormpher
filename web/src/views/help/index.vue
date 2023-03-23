<script setup lang="ts">
import { marked } from 'marked'
import helpmd from './help.md?raw'

const markdown = marked.parse(helpmd)
</script>

<template>
  <article class="text-base prose prose-truegray">
    <div v-html="markdown" />
  </article>
</template>
