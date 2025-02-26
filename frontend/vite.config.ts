import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      // This resolves the tilde paths
      '~@ibm/plex': path.resolve(__dirname, 'node_modules/@ibm/plex'),
      // You might also need a general alias for node_modules
      '~': path.resolve(__dirname, 'node_modules')
    }
  },
  // Ensure Vite properly handles font files
  assetsInclude: ['**/*.woff2', '**/*.woff', '**/*.ttf'],
})
