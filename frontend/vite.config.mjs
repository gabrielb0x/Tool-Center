import { defineConfig } from 'vite';
import { resolve } from 'path';
import fg from 'fast-glob';
import obfuscator from 'vite-plugin-javascript-obfuscator';
import tailwindcss from '@tailwindcss/vite';

const OBF_LEVEL   = 'hard';     // 'none' | 'simple' | 'medium' | 'hard'
const MINIFY_HARD = true;       // true  = Terser vénère, false = esbuild par défaut

const obfPresets = {
  simple: {
    compact: true,
    identifierNamesGenerator: 'hexadecimal'
  },
  medium: {
    compact: true,
    controlFlowFlattening: true,
    identifierNamesGenerator: 'hexadecimal',
    stringArray: true,
    stringArrayEncoding: ['rc4'],
    stringArrayThreshold: 0.6
  },
  hard: {
    compact: true,
    controlFlowFlattening: true,
    deadCodeInjection: true,
    debugProtection: false, // Cette cause des problèmes de performance, jusqu'à 1S DE LATENCE !
    selfDefending: true,
    identifierNamesGenerator: 'hexadecimal',
    stringArray: true,
    stringArrayEncoding: ['rc4'],
    stringArrayThreshold: 1
  }
};

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
    plugins: [
      tailwindcss(),
      ...(OBF_LEVEL !== 'none'
        ? [obfuscator({ apply: 'build', options: obfPresets[OBF_LEVEL] })]
        : [])
    ],
    build: {
      outDir: 'dist',
      minify: MINIFY_HARD ? 'terser' : true,
      sourcemap: false,
      rollupOptions: { input }
    }
  };
});
