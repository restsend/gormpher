<script setup>
import { onMounted, ref } from 'vue'

const MONTH_NAMES = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December']
const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const showDatepicker = ref(false)
const datepickerValue = ref('')
const month = ref('')
const year = ref('')
const no_of_days = ref([])
const blankdays = ref([])
const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

function initDate() {
  const today = new Date()
  month.value = today.getMonth()
  year.value = today.getFullYear()
  datepickerValue.value = new Date(year.value, month.value, today.getDate()).toDateString()
}

function isToday(date) {
  const today = new Date()
  const d = new Date(year.value, month.value, date.value)
  return today.toDateString() === d.toDateString()
}

function getDateValue(date) {
  const selectedDate = new Date(year.value, month.value, date)
  datepickerValue.value = selectedDate.toDateString()

  // this.$refs.date.value = `${selectedDate.getFullYear()}-${(`0${selectedDate.getMonth()}`).slice(-2)}-${(`0${selectedDate.getDate()}`).slice(-2)}`
  // console.log(this.$refs.date.value)
  showDatepicker.value = false
}

function getNoOfDays() {
  const daysInMonth = new Date(year.value, month.value + 1, 0).getDate()

  // find where to start calendar day of week
  const dayOfWeek = new Date(year.value, month.value).getDay()
  const blankdaysArray = []
  for (let i = 1; i <= dayOfWeek; i++)
    blankdaysArray.push(i)

  const daysArray = []
  for (let i = 1; i <= daysInMonth; i++)
    daysArray.push(i)

  blankdays.value = blankdaysArray
  no_of_days.value = daysArray
}

onMounted(() => {
  initDate()
  getNoOfDays()
})
</script>

<template>
  <!-- component -->
  <div class="h-screen w-screen flex items-center justify-center bg-gray-200 ">
    <div class="antialiased sans-serif">
      <div class="container mx-auto px-4 py-2 md:py-10">
        <div class="mb-5 w-64">
          <label for="datepicker" class="font-bold mb-1 text-gray-700 block">Select Date</label>
          <div class="relative">
            <input ref="date" type="hidden" name="date">
            <input
              v-model="datepickerValue"
              type="text"
              readonly
              class="w-full pl-4 pr-10 py-3 leading-none rounded-lg shadow-sm focus:outline-none focus:shadow-outline text-gray-600 font-medium"
              placeholder="Select date"
              @click="showDatepicker = !showDatepicker"
              @keydown.escape="showDatepicker = false"
            >

            <div class="absolute top-0 right-0 px-3 py-2">
              <svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>

            <!-- <div v-text="no_of_days.length"></div>
                          <div v-text="32 - new Date(year, month, 32).getDate()"></div>
                          <div v-text="new Date(year, month).getDay()"></div> -->

            <div
              v-show="showDatepicker"
              class="bg-white mt-12 rounded-lg shadow p-4 absolute top-0 left-0"
              style="width: 17rem"
              @click="showDatepicker = false"
            >
              <div class="flex justify-between items-center mb-2">
                <div>
                  <span class="text-lg font-bold text-gray-800" v-text="MONTH_NAMES[month]" />
                  <span class="ml-1 text-lg text-gray-600 font-normal" v-text="year" />
                </div>
                <div>
                  <button
                    type="button"
                    class="transition ease-in-out duration-100 inline-flex cursor-pointer hover:bg-gray-200 p-1 rounded-full"
                    :class="{ 'cursor-not-allowed opacity-25': month === 0 }"
                    :disabled="month === 0 ? true : false"
                    @click="month--; getNoOfDays()"
                  >
                    <svg class="h-6 w-6 text-gray-500 inline-flex" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                    </svg>
                  </button>
                  <button
                    type="button"
                    class="transition ease-in-out duration-100 inline-flex cursor-pointer hover:bg-gray-200 p-1 rounded-full"
                    :class="{ 'cursor-not-allowed opacity-25': month === 11 }"
                    :disabled="month === 11 ? true : false"
                    @click="month++; getNoOfDays()"
                  >
                    <svg class="h-6 w-6 text-gray-500 inline-flex" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                </div>
              </div>

              <div class="flex flex-wrap mb-3 -mx-1">
                <template v-for="(day, index) in DAYS" :key="index">
                  <div style="width: 14.26%" class="px-1">
                    <div
                      class="text-gray-800 font-medium text-center text-xs"
                      v-text="day"
                    />
                  </div>
                </template>
              </div>

              <div class="flex flex-wrap -mx-1">
                <template v-for="blankday in blankdays" :key="blankday">
                  <div
                    style="width: 14.28%"
                    class="text-center border p-1 border-transparent text-sm"
                  />
                </template>
                <template v-for="(date, dateIndex) in no_of_days" :key="dateIndex">
                  <div style="width: 14.28%" class="px-1 mb-1">
                    <div
                      class="cursor-pointer text-center text-sm leading-none rounded-full leading-loose transition ease-in-out duration-100"
                      :class="{ 'bg-blue-500 text-white': isToday(date) === true, 'text-gray-700 hover:bg-blue-200': isToday(date) === false }"
                      @click="getDateValue(date)"
                      v-text="date"
                    />
                  </div>
                </template>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
