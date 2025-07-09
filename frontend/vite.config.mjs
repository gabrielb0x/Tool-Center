import { defineConfig } from 'vite';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import fg from 'fast-glob';
import tailwindcss from '@tailwindcss/vite';
import obfuscator from 'vite-plugin-javascript-obfuscator';

const __dirname = dirname(fileURLToPath(import.meta.url));

const OBF_LEVEL   = 'hard';   // 'none' | 'simple' | 'medium' | 'hard'
const MINIFY_HARD = true;     // true = terser hard, false = esbuild
const ghost_names = true;    // true = noms hashéss, false = Noms normaux

const obfPresets = {
  simple: {
    compact: true,
    identifierNamesGenerator: 'hexadecimal',
  },
  medium: {
    compact: true,
    controlFlowFlattening: true,
    identifierNamesGenerator: 'hexadecimal',
    stringArray: true,
    stringArrayEncoding: ['rc4'],
    stringArrayThreshold: 0.6,
  },
  hard: {
    compact: true,
    controlFlowFlattening: true,
    deadCodeInjection: true,
    debugProtection: false, // Fait laguer le navigateur de 1 seconde entière !
    selfDefending: true,
    identifierNamesGenerator: 'hexadecimal',
    stringArray: true,
    stringArrayEncoding: ['rc4'],
    stringArrayThreshold: 1,
  },
};

if (OBF_LEVEL === 'none') {
  console.warn('⚠️ Obfuscation is disabled. This is not recommended for production!');
}
if (OBF_LEVEL === 'hard' && !MINIFY_HARD) {
  console.warn('⚠️ Hard obfuscation is enabled, but minification is set to false. This may lead to larger file sizes and reduced performance.');
}

const output = ghost_names
  ? {
      entryFileNames: 'assets/[hash].js',
      chunkFileNames: 'assets/[hash].js',
      assetFileNames: 'assets/[hash].[ext]',
    }
  : {
      entryFileNames: 'assets/[name].js',
      chunkFileNames: 'assets/[name].js',
      assetFileNames: 'assets/[name].[ext]',
    };

export default defineConfig(async () => {
  const files = await fg(['**/*.html', '!node_modules/**', '!dist/**']);
  const input = Object.fromEntries(
    files.map((f) => [
      f.replace(/\.html$/, '').replace(/[\\/]/g, '_'),
      resolve(__dirname, f),
    ]),
  );

  return {
    root: '.',
    publicDir: 'public/',
    plugins: [
      tailwindcss(),
      ...(OBF_LEVEL !== 'none'
        ? [obfuscator({ apply: 'build', options: obfPresets[OBF_LEVEL] })]
        : []),
    ],
    build: {
      outDir: 'dist',
      minify: MINIFY_HARD ? 'terser' : true,
      sourcemap: false,
      rollupOptions: {
        input,
        output
      }
    },
  };
});
