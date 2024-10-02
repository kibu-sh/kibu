// @ts-check
import starlight from '@astrojs/starlight'
import tailwind from '@astrojs/tailwind'
import { defineConfig } from 'astro/config'

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      title: 'Kibu Documentation',
      logo: {
        replacesTitle: true,
        dark: './src/assets/logo dark.svg',
        light: './src/assets/logo light.svg',
      },
      social: {
        github: 'https://github.com/kibu-sh/kibu',
        youtube: 'https://www.youtube.com/@kibu-sh',
        twitter: 'https://twitter.com/kibu_sh',
        discord: 'https://discord.gg/5sga863FVB',
      },
      sidebar: [
        {
          label: 'Guides',
          autogenerate: { directory: 'guides' },
          // items: [
          //   // Each item here is one entry in the navigation menu.
          //   // { label: 'Example Guide', slug: 'guides/example' },
          // ],
        },
        {
          label: 'Reference',
          autogenerate: { directory: 'reference' },
        },
      ],
      customCss: ['./src/tailwind.css'],
    }),
    tailwind({ applyBaseStyles: false }),
  ],
})
