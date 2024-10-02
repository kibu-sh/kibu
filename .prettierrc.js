module.exports = {
  tabWidth: 2,
  trailingComma: 'all',
  singleQuote: true,
  printWidth: 80,
  semi: false,
  plugins: [require('prettier-plugin-tailwindcss')],
  tailwindConfig: './pkg/wiretap/ui/tailwind.config.js',
}
