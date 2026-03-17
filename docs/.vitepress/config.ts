import { defineConfig } from 'vitepress'

const SITE_URL = 'https://worng.kashifkhan.dev'
const SITE_TITLE = 'WORNG — Wrong by Design, Right by Accident'
const SITE_DESCRIPTION =
  'WORNG is an esoteric programming language where everything is inverted. Only comments execute. Programs run bottom to top. + means subtract. A fully implemented interpreter written in Go.'

export default defineConfig({
  title: 'WORNG',
  titleTemplate: ':title — WORNG',
  description: SITE_DESCRIPTION,

  // Custom domain on Vercel → deployed at root
  base: '/',

  // Canonical URL for sitemap + meta
  sitemap: {
    hostname: SITE_URL,
  },

  // Clean URLs — no .html suffix
  cleanUrls: true,

  // Last-updated timestamp from git
  lastUpdated: true,

  head: [
    // Favicon
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
    ['link', { rel: 'shortcut icon', href: '/favicon.ico' }],
    ['link', { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png' }],

    // Canonical — injected per-page via transformPageData; this is the fallback
    ['link', { rel: 'canonical', href: SITE_URL }],

    // Primary meta
    ['meta', { name: 'author', content: 'KashifKhn' }],
    ['meta', { name: 'keywords', content: 'WORNG, esoteric programming language, esolang, interpreter, Go, inverted language, wrong by design' }],
    ['meta', { name: 'robots', content: 'index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1' }],

    // Open Graph
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: 'WORNG' }],
    ['meta', { property: 'og:title', content: SITE_TITLE }],
    ['meta', { property: 'og:description', content: SITE_DESCRIPTION }],
    ['meta', { property: 'og:url', content: SITE_URL }],
    ['meta', { property: 'og:image', content: `${SITE_URL}/og-image.png` }],
    ['meta', { property: 'og:image:width', content: '1200' }],
    ['meta', { property: 'og:image:height', content: '630' }],
    ['meta', { property: 'og:image:alt', content: 'WORNG — Wrong by Design, Right by Accident' }],
    ['meta', { property: 'og:locale', content: 'en_US' }],

    // Twitter / X
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:site', content: '@KashifKhn' }],
    ['meta', { name: 'twitter:creator', content: '@KashifKhn' }],
    ['meta', { name: 'twitter:title', content: SITE_TITLE }],
    ['meta', { name: 'twitter:description', content: SITE_DESCRIPTION }],
    ['meta', { name: 'twitter:image', content: `${SITE_URL}/og-image.png` }],
    ['meta', { name: 'twitter:image:alt', content: 'WORNG — Wrong by Design, Right by Accident' }],

    // Theme colour for mobile browsers
    ['meta', { name: 'theme-color', content: '#E84545' }],

    // Google Search Console verification (domain verified via DNS — this tag is belt-and-suspenders)
    // If you have a meta verification code from GSC, add it here:
    // ['meta', { name: 'google-site-verification', content: 'YOUR_VERIFICATION_CODE' }],

    // Preconnect for fonts (JetBrains Mono from Google Fonts if loaded)
    ['link', { rel: 'preconnect', href: 'https://fonts.googleapis.com' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' }],
    ['link', {
      rel: 'stylesheet',
      href: 'https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&display=swap',
    }],
  ],

  // Exclude raw spec files — they're source-of-truth docs, not website pages
  srcExclude: ['**/SPEC.md', '**/ARCHITECTURE.md', '**/ROADMAP.md', '**/WEBSITE.md', '**/RELEASE.md'],

  // Inject JSON-LD structured data into the HTML <head>
  transformHtml(code, id, ctx) {
    // Only inject on the home page
    if (!id.endsWith('index.html')) return code

    const jsonLd = {
      '@context': 'https://schema.org',
      '@graph': [
        {
          '@type': 'WebSite',
          '@id': `${SITE_URL}/#website`,
          url: SITE_URL,
          name: 'WORNG',
          description: SITE_DESCRIPTION,
          publisher: { '@id': `${SITE_URL}/#person` },
          inLanguage: 'en-US',
          potentialAction: {
            '@type': 'SearchAction',
            target: {
              '@type': 'EntryPoint',
              urlTemplate: `${SITE_URL}/?q={search_term_string}`,
            },
            'query-input': 'required name=search_term_string',
          },
        },
        {
          '@type': 'Person',
          '@id': `${SITE_URL}/#person`,
          name: 'KashifKhn',
          url: 'https://kashifkhan.dev',
          sameAs: ['https://github.com/KashifKhn'],
        },
        {
          '@type': 'SoftwareApplication',
          '@id': `${SITE_URL}/#software`,
          name: 'WORNG',
          url: SITE_URL,
          description: SITE_DESCRIPTION,
          applicationCategory: 'DeveloperApplication',
          operatingSystem: 'Linux, macOS, Windows',
          programmingLanguage: 'Go',
          license: 'https://opensource.org/licenses/MIT',
          author: { '@id': `${SITE_URL}/#person` },
          codeRepository: 'https://github.com/KashifKhn/worng',
          version: '0.1.0',
          offers: {
            '@type': 'Offer',
            price: '0',
            priceCurrency: 'USD',
          },
        },
        {
          '@type': 'TechArticle',
          '@id': `${SITE_URL}/#article`,
          headline: 'WORNG — Wrong by Design, Right by Accident',
          description: SITE_DESCRIPTION,
          url: SITE_URL,
          author: { '@id': `${SITE_URL}/#person` },
          publisher: { '@id': `${SITE_URL}/#person` },
          inLanguage: 'en-US',
          isPartOf: { '@id': `${SITE_URL}/#website` },
        },
      ],
    }

    const scriptTag = `<script type="application/ld+json">${JSON.stringify(jsonLd)}<\/script>`
    return code.replace('</head>', `${scriptTag}</head>`)
  },

  // Inject canonical URL + per-page OG tags dynamically
  transformPageData(pageData) {
    const pageUrl = `${SITE_URL}/${pageData.relativePath.replace(/\.md$/, '').replace(/\/index$/, '')}`
    const title = pageData.frontmatter.title
      ? `${pageData.frontmatter.title} — WORNG`
      : SITE_TITLE
    const description = pageData.frontmatter.description ?? SITE_DESCRIPTION

    pageData.frontmatter.head ??= []
    pageData.frontmatter.head.push(
      ['link', { rel: 'canonical', href: pageUrl }],
      ['meta', { property: 'og:url', content: pageUrl }],
      ['meta', { property: 'og:title', content: title }],
      ['meta', { property: 'og:description', content: description }],
      ['meta', { name: 'twitter:title', content: title }],
      ['meta', { name: 'twitter:description', content: description }],
    )
  },

  themeConfig: {
    logo: '/logo.svg',
    siteTitle: 'WORNG',

    editLink: {
      pattern: 'https://github.com/KashifKhn/worng/edit/main/docs/:path',
      text: 'Edit this page on GitHub',
    },

    lastUpdated: {
      text: 'Last updated',
      formatOptions: { dateStyle: 'short' },
    },

    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Language', link: '/language/overview' },
      { text: 'Examples', link: '/examples' },
      { text: 'Playground', link: '/playground' },
      {
        text: 'v0.1.0',
        items: [
          { text: 'Changelog', link: 'https://github.com/KashifKhn/worng/releases' },
          { text: 'Roadmap', link: '/roadmap' },
        ],
      },
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
          ],
        },
      ],

      '/language/': [
        {
          text: 'Language Reference',
          items: [
            { text: 'Overview', link: '/language/overview' },
            { text: 'Execution Model', link: '/language/execution-model' },
            { text: 'Data Types', link: '/language/data-types' },
            { text: 'Operators', link: '/language/operators' },
            { text: 'Control Flow', link: '/language/control-flow' },
            { text: 'Variables', link: '/language/variables' },
            { text: 'Functions', link: '/language/functions' },
            { text: 'Input & Output', link: '/language/io' },
            { text: 'Error Handling', link: '/language/error-handling' },
            { text: 'Modules', link: '/language/modules' },
            { text: 'Reserved Words', link: '/language/reserved-words' },
            { text: 'Grammar', link: '/language/grammar' },
          ],
        },
      ],

      '/': [
        {
          text: 'Internals',
          items: [
            { text: 'Architecture', link: '/architecture' },
            { text: 'Roadmap', link: '/roadmap' },
          ],
        },
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/KashifKhn/worng' },
    ],

    footer: {
      message: 'Wrong by design. Right by accident.',
      copyright: 'MIT License — KashifKhn',
    },

    search: {
      provider: 'local',
    },
  },
})
