import { useDateFormat } from '@vueuse/core'

export function formatDate(time: string, format = 'YYYY-MM-DD HH:mm:ss') {
  return useDateFormat(time, format).value
}
