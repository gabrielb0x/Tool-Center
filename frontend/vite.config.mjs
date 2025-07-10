import { defineConfig } from 'vite';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import fg from 'fast-glob';
import tailwindcss from '@tailwindcss/vite';
import obfuscator from 'vite-plugin-javascript-obfuscator';
import compression from 'vite-plugin-compression';
import { visualizer } from 'rollup-plugin-visualizer';

const __dirname = dirname(fileURLToPath(import.meta.url));

// ---------------------------------------------------------------------------
// ⚙️  Build switches (toggle as you need)
// ---------------------------------------------------------------------------
const OBF_LEVEL     = 'hard';       // 'none' | 'simple' | 'medium' | 'hard'
const MINIFY_HARD   = true;         // true = terser hard, false = esbuild
const ghost_names   = true;         // true = hashed names, false = readable names

// NEW FLAGS ↓↓↓ -------------------------------------------------------------
const ANALYZE_BUNDLE = false;            // false | true
const COMPRESS       = 'none';          // 'none' | 'gzip' | 'brotli' | 'both'
const CSS_MINIFIER   = 'lightningcss';  // 'esbuild' | 'lightningcss'
const SOURCE_MAPS    = 'none';          // 'dev' | 'prod' | 'none'
// ---------------------------------------------------------------------------

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

  const extraPlugins = [];

  if (OBF_LEVEL !== 'none') {
    extraPlugins.push(obfuscator({ apply: 'build', options: obfPresets[OBF_LEVEL] }));
  }

  if (ANALYZE_BUNDLE) {
    extraPlugins.push(
      visualizer({
        filename: 'dist/stats.html',
        template: 'treemap',
        open: false,
        gzipSize: true,
        brotliSize: true,
      }),
    );
  }

  if (COMPRESS === 'gzip' || COMPRESS === 'both') {
    extraPlugins.push(
      compression({ algorithm: 'gzip', ext: '.gz' }),
    );
  }
  if (COMPRESS === 'brotli' || COMPRESS === 'both') {
    extraPlugins.push(
      compression({ algorithm: 'brotliCompress', ext: '.br' }),
    );
  }

  return {
    root: '.',
    publicDir: 'public/',
    plugins: [
      tailwindcss(),
      ...extraPlugins,
    ],

    build: {
      outDir: 'dist',
      minify: MINIFY_HARD ? 'terser' : true,
      cssMinify: CSS_MINIFIER,                // esbuild or lightningcss
      sourcemap: SOURCE_MAPS === 'prod',      // only generate in prod if flag is 'prod'
      rollupOptions: {
        input,
        output,
      },
    },
  };
});
