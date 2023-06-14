<script setup lang="ts">
import { computed, onMounted, reactive, watch } from 'vue'

import Badge from './Badge.vue'
import DynamicInput from './DynamicInput.vue'
import TableItem from './TableItem.vue'
import FilterItem from './FilterItem.vue'
import type { ActionType, TableState } from '@/types'

import api from '@/api'
import useTable from '@/views/table/useTable'
import usePagination from '@/views/table/usePagination'
import { alerter } from '@/components/popup'

interface Props {
  name: string
}

const props = defineProps<Props>()

const state = reactive<TableState>({
  name: props.name,
  fields: [],
  types: [],
  mapping: {},
  goMapping: {},
  filters: [],
  edits: [],
  orders: [],
  searchs: [],
  primaryKey: '',
})

watch(() => props.name, (name) => {
  state.name = name
  initialize()
})

const {
  pos,
  limit,
  keyword,
  filters,
  orders,
  total,
  list,
  form,
  loading,
  selectedIds,
  modalVisible,
  handleQuery,
  handleSearch,
  handleOrder,
  handleRemoveFilter,
  handleShowAdd,
  handleAdd,
  handleShowEdit,
  handleEdit,
  handleDelete,
  handleBatchDelete,
  handleReset,
} = useTable({
  initForm: {},
  addFn: item => api.handleAdd(state.name, item),
  editFn: item => api.handleEdit(state.name, item),
  deleteFn: id => api.handleDelete(state.name, id),
  queryFn: params => api.handleQuery(state.name, params),
  batchFn: ids => api.handleBatch(state.name, ids),
  extraParams: {},
  validateForm: () => true, // no form validate
})

function handleToggleBool(item: any, field: string) {
  try {
    handleEdit({ id: item.id, [field]: !item[field] })
  }
  catch (err: any) {
    alerter.error(err)
  }
}

const indeterminate = computed(() => selectedIds.value.length > 0 && selectedIds.value.length < list.value.length)

const {
  currentPage,
  handleNext,
  handlePrev,
} = usePagination({
  pos,
  limit,
  total,
  callback: handleQuery,
})

function canFilter(field: string) {
  return state.filters.includes(field)
}

function canOrder(field: string) {
  return state.orders.includes(field)
}

function canSearch(field: string) {
  return state.searchs.includes(field)
}

function canEdit(field: string) {
  return state.edits.includes(field)
}

function getActions(field: string) {
  const actions: Array<ActionType> = []
  canFilter(field) && actions.push('filter')
  canOrder(field) && actions.push('order')
  canSearch(field) && actions.push('search')
  canEdit(field) && actions.push('edit')
  return actions
}

async function initialize() {
  if (!state.name)
    return

  loading.value = true

  try {
    const { fields, types, goTypes, filters, orders, searchs, edits, primaryKey }
    = await api.getObject(state.name)

    state.fields = fields
    state.types = types
    state.mapping = state.fields.reduce((acc: any, str, i) => {
      acc[str] = types[i]
      return acc
    }, {})
    state.goMapping = state.fields.reduce((acc: any, str, i) => {
      acc[str] = goTypes[i]
      return acc
    }, {})
    state.filters = filters
    state.orders = orders
    state.searchs = searchs
    state.edits = edits
    state.primaryKey = primaryKey

    handleReset()
  }
  catch (err: any) {
    alerter.error(err)
    console.error(err)
  }
}

onMounted(async () => {
  initialize()
})
</script>

<template>
  <div class="h-full w-full rounded-md p-4 pb-0 space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-4">
        <div class="relative flex">
          <input
            v-model="keyword"
            :disabled="!state.searchs.length"
            type="text"
            placeholder="Keyword search"
            class="input input-bordered input-sm max-w-xs w-full"
            @keyup.enter="handleSearch"
          >
          <button
            class="btn btn-sm btn-square i-uiw:search absolute right-2"
            :disabled="!state.searchs.length"
            @click="handleSearch"
          />
        </div>
        <div class="dropdown dropdown-bottom">
          <label tabindex="0">
            <button class="btn btn-sm btn-circle btn-ghost m-1" :disabled="!state.filters.length">
              <div class="i-uiw:filter p-1" />
            </button>
          </label>
          <ul tabindex="0" class="dropdown-content min-w-120 bg-base-100 p-2 shadow rounded-box space-y-2">
            <li v-for="filter of filters" :key="`${filter}`">
              <FilterItem
                :filter="filter"
                :filters="state.filters"
                :type="state.mapping[filter.name]"
                :go-type="state.goMapping[filter.name]"
                :primary-key="state.primaryKey"
                @handle-query="handleQuery"
                @remove-filter="handleRemoveFilter"
              />
            </li>
            <li class="flex cursor-pointer items-center justify-between">
              <button
                class="btn btn-sm btn-circle btn-ghost"
                :disabled="filters.length >= state.filters.length"
                @click="filters.push({ name: state.filters[filters.length], op: '=', value: '' })"
              >
                <div class="i-uiw:plus" />
              </button>
              <button
                class="btn btn-sm btn-circle btn-ghost"
                @click="handleQuery"
              >
                <div class="i-mdi:magnify" />
              </button>
            </li>
          </ul>
        </div>
        <div class="dropdown">
          <label tabindex="0">
            <button
              class="btn btn-sm btn-circle btn-ghost"
              :disabled="!state.orders.length"
            >
              <div class="i-uiw:bar-chart p-1" />
            </button>
          </label>
          <ul tabindex="0" class="dropdown-content menu w-64 bg-base-100 text-sm shadow rounded-box">
            <div class="card bg-base-100 shadow-xl">
              <div class="card-body">
                orders:
                <pre><code>{{ JSON.stringify(orders, null, '\t') }}</code></pre>
                filters:
                <pre><code>{{ JSON.stringify(filters, null, '\t') }}</code></pre>
              </div>
            </div>
          </ul>
        </div>
        <button
          class="btn btn-sm btn-circle btn-ghost"
          @click="handleReset"
        >
          <div class="i-uiw:reload" />
        </button>
      </div>

      <button class="btn btn-sm btn-primary" @click="handleShowAdd">
        ADD
      </button>
    </div>
    <!-- Table -->
    <div class="relative min-h-md w-full overflow-x-auto">
      <!-- Bulk action -->
      <div
        v-if="selectedIds.length > 0"
        class="absolute left-16 top-0 z-10 h-16 flex items-center sm:left-12 space-x-3"
      >
        <button
          type="button"
          class="btn btn-sm ml-5"
          @click="handleBatchDelete(selectedIds)"
        >
          Delete all
        </button>
      </div>
      <table class="table-normal h-full w-full table text-sm">
        <thead>
          <tr>
            <template v-if="loading">
              <th :colspan="99">
                <div class="w-full flex items-center justify-center text-sm opacity-90">
                  <div class="i-eos-icons:three-dots-loading text-2xl" />
                </div>
              </th>
            </template>
            <template v-else>
              <th>
                <input
                  type="checkbox"
                  class="checkbox"
                  :checked="indeterminate || (selectedIds.length > 0 && selectedIds.length === list.length)"
                  :indeterminate="indeterminate"
                  @change="selectedIds = ($event.target as HTMLInputElement).checked ? list.map(e => e.id) : []"
                >
              </th>
              <th v-for="field of state.fields" :key="field">
                <div class="flex cursor-default items-center space-x-1">
                  <div v-if="canOrder(field)" class="flex flex-col">
                    <div
                      class="i-typcn:arrow-sorted-up transition-500 hover:scale-130"
                      @click="handleOrder(field, 'asc')"
                    />
                    <div
                      class="i-typcn:arrow-sorted-down transition-500 hover:scale-130"
                      @click="handleOrder(field, 'desc')"
                    />
                  </div>
                  <span> {{ field }} </span>
                  <Badge :actions="getActions(field)" />
                </div>
              </th>
              <th />
            </template>
          </tr>
        </thead>
        <tbody>
          <template v-if="loading">
            <tr>
              <td :colspan="99">
                <div class="h-96 w-full flex items-center justify-center text-sm opacity-90">
                  <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"><rect width="6" height="14" x="1" y="4" fill="#888888"><animate id="svgSpinnersBarsScaleFade0" fill="freeze" attributeName="y" begin="0;svgSpinnersBarsScaleFade1.end-0.25s" dur="0.75s" values="1;5" /><animate fill="freeze" attributeName="height" begin="0;svgSpinnersBarsScaleFade1.end-0.25s" dur="0.75s" values="22;14" /><animate fill="freeze" attributeName="opacity" begin="0;svgSpinnersBarsScaleFade1.end-0.25s" dur="0.75s" values="1;.2" /></rect><rect width="6" height="14" x="9" y="4" fill="currentColor" opacity=".4"><animate fill="freeze" attributeName="y" begin="svgSpinnersBarsScaleFade0.begin+0.15s" dur="0.75s" values="1;5" /><animate fill="freeze" attributeName="height" begin="svgSpinnersBarsScaleFade0.begin+0.15s" dur="0.75s" values="22;14" /><animate fill="freeze" attributeName="opacity" begin="svgSpinnersBarsScaleFade0.begin+0.15s" dur="0.75s" values="1;.2" /></rect><rect width="6" height="14" x="17" y="4" fill="currentColor" opacity=".3"><animate id="svgSpinnersBarsScaleFade1" fill="freeze" attributeName="y" begin="svgSpinnersBarsScaleFade0.begin+0.3s" dur="0.75s" values="1;5" /><animate fill="freeze" attributeName="height" begin="svgSpinnersBarsScaleFade0.begin+0.3s" dur="0.75s" values="22;14" /><animate fill="freeze" attributeName="opacity" begin="svgSpinnersBarsScaleFade0.begin+0.3s" dur="0.75s" values="1;.2" /></rect></svg>
                </div>
              </td>
            </tr>
          </template>
          <template v-else-if="!list.length">
            <tr>
              <td :colspan="99">
                <div class="min-h-96 flex items-center justify-center bg-base-100">
                  <!-- <div class="i-simple-icons:protodotio text-5xl" /> -->
                  <span class="select-none font-mono text-5xl text-base-300">
                    Empty
                  </span>
                </div>
              </td>
            </tr>
          </template>
          <template v-else>
            <tr v-for="item of list" :key="item.id" class="group hover">
              <th>
                <input
                  v-model="selectedIds"
                  :value="item.id"
                  type="checkbox"
                  class="checkbox"
                >
              </th>
              <td v-for="field of state.fields" :key="field">
                <div class="">
                  <TableItem
                    :field="field"
                    :value="item[field]"
                    :type="state.mapping[field]"
                    :primary-key="state.primaryKey"
                    :go-type="state.goMapping[field]"
                    :toggle-bool="canEdit(field) ? () => handleToggleBool(item, field) : () => {}"
                    :open-modal="() => handleShowEdit(item)"
                  />
                </div>
              </td>
              <td class="py-0.5">
                <div class="flex items-center space-x-4">
                  <div
                    class="btn btn-circle btn-ghost btn-sm"
                    @click="handleShowEdit(item)"
                  >
                    <span class="i-uiw:edit text-green-400 hover:text-green-400" />
                  </div>
                  <div
                    class="btn btn-circle btn-ghost btn-sm"
                    @click="handleDelete(item.id)"
                  >
                    <span class="i-uiw:delete text-red-500 hover:text-red-500" />
                  </div>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
    <!-- Pagination -->
    <div class="mt-5 flex items-center justify-between">
      <div class="space-x-3">
        <span>
          Total: {{ total }}
        </span>
        <select
          v-model.number="limit"
          class="select select-sm select-bordered"
          @change="handleQuery"
        >
          <option>5</option>
          <option>10</option>
          <option>20</option>
          <option>50</option>
          <option>80</option>
        </select>
      </div>
      <div class="btn-group">
        <button
          class="btn"
          @click="handlePrev"
        >
          <div class="i-ic:sharp-arrow-back-ios" />
        </button>
        <button class="btn">
          PAGE {{ currentPage }}
        </button>
        <button
          class="btn"
          @click="handleNext"
        >
          <div class="i-ic:baseline-arrow-forward-ios" />
        </button>
      </div>
    </div>
  </div>

  <!-- Edit/Add Modal -->
  <div>
    <input v-model="modalVisible" type="checkbox" class="modal-toggle">
    <div class="modal" @click.self="modalVisible = false">
      <div class="modal-box relative max-w-lg" @click="() => {}">
        <h3 class="text-lg font-bold">
          View info
        </h3>
        <div class="py-4">
          <template v-for="field of state.fields" :key="field">
            <div class="my-4 flex justify-between">
              <div class="space-x-1">
                {{ field }}
                <span class="badge badge-outline cursor-default">
                  {{ state.goMapping[field] }}
                </span>
                <Badge v-if="canEdit(field)" :actions="['edit']" />
              </div>
              <div>
                <DynamicInput
                  v-model:value="form[field]"
                  :operation="form.id ? 'edit' : 'add'"
                  :field="field"
                  :go-type="state.goMapping[field]"
                  :type="state.mapping[field]"
                  :primary-key="state.primaryKey"
                  :disabled="form.id && !canEdit(field)"
                />
              </div>
            </div>
          </template>
        </div>
        <div class="modal-action">
          <button
            v-if="!form.id"
            class="btn btn-sm btn-success"
            @click="handleAdd"
          >
            Save
          </button>
          <button
            v-if="form.id"
            class="btn btn-sm btn-warning"
            @click="handleEdit(form)"
          >
            Update
          </button>
          <button class="btn btn-sm" @click="modalVisible = false">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
