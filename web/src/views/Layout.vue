<script setup lang="ts">
import { ref } from 'vue'
import api from '@/api'
import Table from '@/views/table/index.vue'
import Help from '@/views/help/index.vue'

const name = ref('')
const menus = ref<string[]>([])
menus.value = await api.getObjectNames()
</script>

<template>
  <div class="flex flex-col h-screen">
    <nav class="navbar bg-base-100 shadow-sm fixed top-0 z-100">
      <div class="navbar-start">
        <div class="dropdown">
          <!-- Mobile -->
          <div class="block sm:hidden">
            <label tabindex="0" class="btn btn-ghost btn-circle">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" /></svg>
            </label>
            <ul tabindex="0" class="menu menu-compact dropdown-content mt-3 p-2 shadow bg-base-100 rounded-box w-52">
              <li v-for="menu in menus" :key="menu">
                <a @click="name = menu">
                  <div class="i-bi:lightning text-lg" />
                  <span class="text-lg">
                    {{ menu }}
                  </span>
                </a>
              </li>
            </ul>
          </div>
          <!-- Desktop -->
          <div class="mx-4 hidden sm:block">
            <a class="btn btn-ghost normal-case text-2xl" @click="name = ''">
              Gormpher Admin
            </a>
          </div>
        </div>
      </div>
      <div class="navbar-end">
        <button class="btn btn-ghost btn-circle">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
        </button>
        <button class="btn btn-ghost btn-circle">
          <div class="indicator">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>
            <span class="badge badge-xs badge-primary indicator-item" />
          </div>
        </button>
      </div>
    </nav>

    <main class="flex-1 flex flex-row my-20">
      <div class="shrink-0 hidden sm:block sm:w-1/5">
        <!-- Side menu -->
        <ul class="sticky top-20 menu mx-2">
          <li class="menu-title">
            <span>Index</span>
          </li>
          <li>
            <a @click="name = ''">
              <div class="i-uiw:question-circle-o" />
              Help
            </a>
          </li>
          <li class="menu-title">
            <span>Webobjects</span>
          </li>
          <li v-for="menu in menus" :key="menu">
            <a @click="name = menu">
              <div class="i-bi:lightning text-base" />
              <span class="text-lg">
                {{ menu }}
              </span>
            </a>
          </li>
        </ul>
      </div>
      <div class="mr-2 w-full sm:w-4/5">
        <template v-if="name">
          <div class="px-4 text-sm breadcrumbs">
            <ul>
              <li>Webobjects</li>
              <li>{{ name }}</li>
            </ul>
          </div>
          <Table :name="name" />
        </template>
        <template v-else>
          <Help />
        </template>
      </div>
    </main>

    <footer class="footer footer-center p-4 bg-base-300 text-base-content">
      <div>
        <p>
          This project is powered by
          <a class="link" href="https://github.com/restsend/gormpher" target="_blank"> Gormpher </a>
          .
        </p>
      </div>
    </footer>
  </div>
</template>
