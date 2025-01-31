import {defineConfig} from 'vite'
import {svelte} from '@sveltejs/vite-plugin-svelte'
import { ViteAliases } from 'vite-aliases'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte(), ViteAliases()],
})
