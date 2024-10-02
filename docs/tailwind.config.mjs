import starlightPlugin from '@astrojs/starlight-tailwind'

const kibu = {
  light: '#CBF2AE',
}

// Generated color palettes
const accent = {
  200: '#878885', // dark accent
  600: '#a0e26f', // light accent
  900: '#262626',
  950: '#1e1e1e',
}
const gray = {
  100: '#f6f6f6',
  200: '#eeeeee',
  300: '#c2c2c2',
  400: '#8b8b8b',
  500: '#313131',
  700: '#383838',
  800: '#181818', // dark sidebar
  900: '#181818', // dark background
}

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
  theme: {
    extend: {
      colors: {
        kibu,
        // Your preferred accent color. Indigo is closest to Starlight’s defaults.
        accent,
        // Your preferred gray scale. Zinc is closest to Starlight’s defaults.
        gray,
      },
      // fontFamily: {
      //   // Your preferred text font. Starlight uses a system font stack by default.
      //   sans: ['"Atkinson Hyperlegible"'],
      //   // Your preferred code font. Starlight uses system monospace fonts by default.
      //   mono: ['"IBM Plex Mono"'],
      // },
    },
  },
  plugins: [starlightPlugin()],
}
