<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps({
  title: { type: String },
  content: { type: String },
  open: { type: Boolean, default: true },
  onPositiveClick: { type: Function, default: () => {} },
  onNegativeClick: { type: Function, default: () => {} },
  onClose: { type: Function, default: () => {} },
})

const show = ref(false)

setTimeout(() => {
  show.value = props.open
}, 10)

function closeModal(flag: boolean) {
  show.value = false
  flag ? props.onPositiveClick() : props.onNegativeClick()
  setTimeout(() => props.onClose(flag), 300)
}
</script>

<template>
  <div>
    <input v-model="show" type="checkbox" class="modal-toggle">
    <div class="modal" @click.self="show = false">
      <div class="modal-box relative max-w-lg rounded-md" @click="() => {}">
        <h3 class="text-lg font-bold">
          {{ title }}
        </h3>
        <p class="py-4">
          {{ content }}
        </p>
        <div class="modal-action">
          <button class="btn btn-sm" @click="closeModal(false)">
            NO
          </button>
          <button class="btn btn-sm btn-error" @click="closeModal(true)">
            YES
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
