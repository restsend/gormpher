import { ref } from 'vue'

import { alerter } from '@/popup'
import type { Filter, FilterOp, Order, OrderOp } from '@/types'

interface ListOptions {
  initForm: {}
  validateForm?: () => boolean
  queryFn: (params: QueryParams) => Promise<QueryResult>
  editFn: (item: EditItem) => Promise<void>
  addFn: (form: any) => Promise<void> // | boolean
  deleteFn: (id: string | number) => Promise<void>
  batchFn: (form: string[]) => Promise<void>
  queryParams: any // 额外的 query 参数
}

interface EditItem {
  id: string | number
  // ...
}

interface QueryParams {
  pos?: number
  limit?: number
  keyword?: string
  filters?: Array<Filter>
  orders?: Array<Order>
  // [key: string]: any
}

interface QueryResult {
  pos?: number
  limit?: number
  keyword?: string
  total: number
  items?: any[]
}

export default function useTable({
  initForm = {},
  validateForm = () => true,
  addFn,
  editFn,
  deleteFn,
  queryFn,
  batchFn,
  queryParams = {},
}: ListOptions) {
  // table
  const selectedIds = ref<number[] | string[]>([])
  const loading = ref(false)
  const form = ref<any>({})
  const modalVisible = ref(false)
  const modalLoading = ref(false)

  // queryParams
  const pos = ref(0)
  const limit = ref(10)
  const keyword = ref('')

  const filters = ref<Filter[]>([])
  const orders = ref<Order[]>([])

  // queryResult
  const total = ref(0)
  const list = ref<any[]>([])

  function handleShowAdd(): void {
    modalVisible.value = true
    form.value = { ...initForm }
  }

  function handleShowEdit(row: any): void {
    modalVisible.value = true
    form.value = { ...row }
  }

  async function handleQuery(): Promise<void> {
    selectedIds.value = []
    loading.value = true
    try {
      const resp = await queryFn({
        pos: pos.value,
        limit: limit.value,
        keyword: keyword.value,
        filters: filters.value,
        orders: orders.value,
        ...queryParams,
      })
      total.value = resp.total
      list.value = resp.items ?? []
    }
    catch (err: any) {
      alerter.error(err)
    }
    finally {
      setTimeout(() => loading.value = false, 100)
    }
  }

  function handleSearch(): void {
    pos.value = 0
    handleQuery()
  }

  async function handleEdit(item: EditItem): Promise<void> {
    if (!validateForm())
      return

    try {
      modalLoading.value = true
      await editFn(item)
      modalVisible.value = false
      handleQuery()
      alerter.success('Edit success!')
    }
    catch (err: any) {
      alerter.error(err)
    }
    finally {
      modalLoading.value = false
    }
  }

  async function handleAdd(): Promise<void> {
    if (!validateForm())
      return

    try {
      modalLoading.value = true
      await addFn(form.value)
      modalVisible.value = false
      setTimeout(() => handleQuery(), 100)
      alerter.success('Add success!')
    }
    catch (err: any) {
      alerter.error(err)
    }
    finally {
      modalLoading.value = false
    }
  }

  async function handleDelete(id: number | string): Promise<void> {
    try {
      await deleteFn(id)
      modalVisible.value = false
      handleQuery()
      alerter.success('Delete success!')
    }
    catch (err: any) {
      alerter.error(err)
    }
  }

  async function handleBatch(ids: number[] | string[]): Promise<void> {
    try {
      await batchFn(ids.map(e => String(e)))
      handleQuery()
    }
    catch (err: any) {
      alerter.error(err)
    }
  }

  function handleOrder(field: string, op: OrderOp) {
    orders.value = orders.value.filter(e => e.name !== field)
    orders.value.push({ name: field, op })
    handleQuery()
  }

  function handleRemoveOrder(field: string) {
    orders.value = orders.value.filter(e => e.name !== field)
    handleQuery()
  }

  function handleFilter(field: string, op: FilterOp, value: any) {
    filters.value = filters.value.filter(e => e.name !== field)
    filters.value.push({ name: field, op, value })
    handleQuery()
  }

  function handleRemoveFilter(field: string) {
    filters.value = filters.value.filter(e => e.name !== field)
    handleQuery()
  }

  function handleReset() {
    pos.value = 0
    limit.value = 10
    keyword.value = ''
    filters.value = []
    orders.value = []

    handleQuery()
  }

  return {
    pos,
    limit,
    keyword,
    filters,
    orders,
    total,
    list,
    loading,
    selectedIds,
    modalVisible,
    modalLoading,
    form,
    handleQuery,
    handleSearch,
    handleOrder,
    handleRemoveOrder,
    handleFilter,
    handleRemoveFilter,
    handleEdit,
    handleShowEdit,
    handleAdd,
    handleShowAdd,
    handleDelete,
    handleBatch,
    handleReset,
  }
}
