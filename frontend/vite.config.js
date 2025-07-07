import { defineConfig } from 'vite';
import { resolve } from 'path';
import fg from 'fast-glob';

export default defineConfig(async () => {
  const files = await fg(['**/*.html', '!node_modules/**', '!dist/**']);
  const input = Object.fromEntries(
    files.map(f => [
      f.replace(/\.html$/, '').replace(/[\/\\]/g, '_'),
      resolve(__dirname, f)
    ])
  );

  return {
    root: '.',
    publicDir: 'public/',
    build: {
      outDir: 'dist',
      rollupOptions: { input }
    }
  };
});
