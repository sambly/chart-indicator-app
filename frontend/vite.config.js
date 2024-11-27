import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
  root: path.resolve(__dirname),
  build: {
    outDir: path.resolve(__dirname, 'dist'),
    manifest: true, 
    rollupOptions: {
      input: path.resolve(__dirname, 'src/main.ts'),
    },
  },
});
