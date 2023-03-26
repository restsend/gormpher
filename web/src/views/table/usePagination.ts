import type { Ref } from 'vue'
import { computed, ref } from 'vue'

interface PaginationOptions {
  pos: Ref<number>
  limit: Ref<number>
  total: Ref<number>
  callback?: Function
}

export default function usePagination({
  pos = ref(0),
  limit = ref(10),
  total = ref(0),
  callback = () => {},
}: PaginationOptions) {
  const currentPage = computed(() => 1 + pos.value / limit.value + (pos.value % limit.value))

  function handleNext() {
    if (pos.value + limit.value >= total.value)
      return
    pos.value += limit.value
    callback()
  }
  function handlePrev() {
    if (pos.value === 0 || pos.value < limit.value)
      return
    pos.value -= limit.value
    callback()
  }

  return {
    currentPage,
    handleNext,
    handlePrev,
  }
}
