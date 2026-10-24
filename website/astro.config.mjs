// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
	site: 'https://vardius.github.io',
	base: '/worker-pool',
	trailingSlash: 'always',
	// Pages import the runnable programs in ../examples, so the dev server may read outside this folder.
	vite: { server: { fs: { allow: ['..'] } } },
	integrations: [
		starlight({
			title: 'worker-pool',
			description:
				'A worker pool for Go: one bounded queue of typed jobs, with backpressure, workers you add and remove at any time, and a graceful stop.',
			logo: { src: './src/assets/mark.svg', alt: '' },
			favicon: '/favicon.svg',
			components: { Hero: './src/components/Hero.astro' },
			// Long lines wrap instead of scrolling under the copy button on phones.
			expressiveCode: { defaultProps: { wrap: true } },
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/vardius/worker-pool' }],
			customCss: ['./src/styles/custom.css'],
			sidebar: [
				{
					label: 'Start here',
					items: ['getting-started'],
				},
				{
					label: 'The pool',
					items: ['queue', 'workers', 'stopping'],
				},
				{
					label: 'Reference',
					items: [
						'benchmarks',
						'migrating-from-v2',
						{ label: 'API reference', link: 'https://pkg.go.dev/github.com/vardius/worker-pool/v3' },
					],
				},
			],
		}),
	],
});
