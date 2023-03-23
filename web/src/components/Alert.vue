<script setup lang="ts">
import { h, ref } from 'vue'

interface Props {
  type: 'info' | 'success' | 'warning' | 'error'
  message: string
  delay?: number
  onClose?: Function
}

const props = withDefaults(defineProps<Props>(), {
  type: 'info',
  delay: 3 * 1000,
  onClose: () => {},
})

const show = ref(true)

function close() {
  show.value = false
  setTimeout(() => {
    props.onClose()
  }, 2000)
}

setTimeout(() => close(), props.delay)

function renderTypeIcon() {
  switch (props.type) {
    case 'success':
      return h('div', { class: 'i-clarity:success-standard-line' })
    case 'error':
      return h('div', { class: 'i-clarity:error-standard-line' })
    case 'warning':
      return h('div', { class: 'i-clarity:warning-standard-line' })
    default:
      return h('div', { class: 'i-clarity:info-standard-line' })
  }
}
</script>

<template>
  <Transition>
    <div v-if="show" class="fixed z-100 top-0 right-0 p-5 w-120">
      <div class="alert shadow-lg" :class="`alert-${type}`">
        <div>
          <component :is="renderTypeIcon" />
          <span> {{ message }} </span>
        </div>
        <div class="flex-none">
          <button
            class="btn btn-sm"
            :class="`btn-${type}`"
            @click="show = false"
          >
            Accept
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.v-enter-active,
.v-leave-active {
  transition: opacity 0.5s ease;
}

.v-enter-from,
.v-leave-to {
  opacity: 0;
}
</style>
