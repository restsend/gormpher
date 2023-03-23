import {
  defineConfig,
  presetIcons,
  presetTypography,
  presetUno,
  // presetWebFonts,
  transformerDirectives,
} from 'unocss'
import { presetDaisy } from 'unocss-preset-daisy'

export default defineConfig({
  shortcuts: [
    ['f-c-c', 'flex items-center justify-center'],
    ['wh-full', 'w-full h-full'],
    ['wh-screen', 'w-screen h-screen'],
  ],
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
        // display: 'inline-block',
      },
    }),
    // presetWebFonts({
    // fonts: {
    //   sans: 'DM Sans',
    //   serif: 'DM Serif Display',
    //   mono: 'DM Mono',
    // },
    // }),
  ],
})
