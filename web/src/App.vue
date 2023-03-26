<script setup lang="ts">
import Layout from './views/Layout.vue'
</script>

<template>
  <Suspense>
    <template #fallback>
      <div class="hero min-h-screen bg-base-200">
        <div class="hero-content text-center">
          <div class="max-w-md">
            <h1 class="text-5xl font-bold">
              Hello there
            </h1>
            <p class="py-6">
              Sorry, it looks like <strong>the backend server is not running</strong> . Please make sure to start the backend server before running the frontend to see the full functionality of the project.
            </p>
            <a class="link" href="https://github.com/restsend/gormpher" target="_blank">
              View On Github
            </a>
          </div>
        </div>
      </div>
    </template>
    <Layout />
  </Suspense>
</template>
