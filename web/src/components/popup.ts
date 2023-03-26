/* eslint-disable vue/one-component-per-file */
import { createApp } from 'vue'
import Confirm from './Confirm.vue'
import Alert from '@/components/Alert.vue'

interface ConfirmOptions {
  title: string
  content: string
  onNegativeClick?: Function
  onPositiveClick?: Function
}

export function confirm({
  title,
  content,
  onNegativeClick = () => {},
  onPositiveClick = () => {},
}: ConfirmOptions) {
  const parentNode = document.createElement('div')

  const instance = createApp(Confirm, {
    title,
    content,
    onNegativeClick,
    onPositiveClick,
    onClose: () => {
      instance.unmount()
      document.body.removeChild(parentNode)
    },
  })

  document.body.appendChild(parentNode)
  instance.mount(parentNode)
}

interface AlertOptions {
  type?: 'success' | 'error' | 'warning' | 'info'
  message: string
  delay?: number
}

export function showAlert({
  type = 'info',
  message,
  delay = 2 * 1000,
}: AlertOptions) {
  const parentNode = document.createElement('div')

  const instance = createApp(Alert, {
    type,
    message,
    delay,
    onClose: () => {
      instance.unmount()
      document.body.removeChild(parentNode)
    },
  })

  document.body.appendChild(parentNode)
  instance.mount(parentNode)
}

class Alerter {
  info(message: string) {
    showAlert({ type: 'info', message })
  }

  success(message: string) {
    showAlert({ type: 'success', message })
  }

  warning(message: string) {
    showAlert({ type: 'warning', message })
  }

  error(message: string) {
    showAlert({ type: 'error', message })
  }
}

export const alerter = new Alerter()
