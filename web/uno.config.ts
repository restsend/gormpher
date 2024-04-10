import {
  defineConfig,
  presetIcons,
  presetTypography,
  presetUno,
  transformerDirectives,
} from 'unocss'
import { presetDaisy } from 'unocss-preset-daisy'

export default defineConfig({
  transformers: [transformerDirectives()],
  presets: [
    presetUno(),
    presetDaisy(),
    presetTypography(),
    presetIcons({
      scale: 1.2,
      warn: true,
      extraProperties: {
        cursor: 'pointer',
      },
    }),
  ],
})
