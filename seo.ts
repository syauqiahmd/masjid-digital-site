import type { HtmlTagDescriptor, Plugin } from 'vite'

export type SeoEnv = { VITE_SITE_URL?: string; VITE_GOOGLE_SITE_VERIFICATION?: string }

// the production site; VITE_SITE_URL overrides it, e.g. for staging
export const DEFAULT_SITE_URL = 'https://masjidyuk.com'

// every page of the site, as served by Caddy
export const PAGES = ['/', '/docs/']

// the image shared links show (Google, WhatsApp, Facebook, X)
const SHARE_IMAGE = { path: '/screenshots/tv.png', width: '1920', height: '1080' }

/** The site's origin without a trailing slash: VITE_SITE_URL, or the production site when empty. */
export function siteOrigin(env: SeoEnv): string {
  return (env.VITE_SITE_URL?.trim() || DEFAULT_SITE_URL).replace(/\/+$/, '')
}

/** The page's public path for the HTML file Vite builds: /index.html is /, docs/index.html is /docs/. */
export function pagePath(htmlPath: string): string {
  return `/${htmlPath.replace(/^\/+/, '').replace(/index\.html$/, '')}`
}

export function robotsTxt(origin: string): string {
  return [
    'User-agent: *',
    'Allow: /',
    'Disallow: /api/',
    '',
    `Sitemap: ${origin}/sitemap.xml`,
    '',
  ].join('\n')
}

export function sitemapXml(origin: string, lastmod: string): string {
  const urls = PAGES.map(
    (path) => `  <url><loc>${origin}${path}</loc><lastmod>${lastmod}</lastmod></url>`,
  )
  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
    ...urls,
    '</urlset>',
    '',
  ].join('\n')
}

/** The head tags that need the site's absolute URL: canonical, og:url and the share image. */
export function headTags(env: SeoEnv, path: string): HtmlTagDescriptor[] {
  const meta = (attrs: Record<string, string>): HtmlTagDescriptor => ({
    tag: 'meta',
    attrs,
    injectTo: 'head',
  })
  const tags: HtmlTagDescriptor[] = []
  const verification = env.VITE_GOOGLE_SITE_VERIFICATION?.trim()
  if (verification) tags.push(meta({ name: 'google-site-verification', content: verification }))

  const origin = siteOrigin(env)
  const image = `${origin}${SHARE_IMAGE.path}`
  tags.push(
    { tag: 'link', attrs: { rel: 'canonical', href: `${origin}${path}` }, injectTo: 'head' },
    meta({ property: 'og:url', content: `${origin}${path}` }),
    meta({ property: 'og:image', content: image }),
    meta({ property: 'og:image:width', content: SHARE_IMAGE.width }),
    meta({ property: 'og:image:height', content: SHARE_IMAGE.height }),
    meta({ name: 'twitter:image', content: image }),
  )
  return tags
}

/** Search engine files for the build: the absolute-URL head tags on each page, robots.txt, sitemap.xml. */
export function seo(env: SeoEnv): Plugin {
  const origin = siteOrigin(env)
  return {
    name: 'masjid-seo',
    transformIndexHtml(_html, ctx) {
      return headTags(env, pagePath(ctx.path))
    },
    generateBundle() {
      const lastmod = new Date().toISOString().slice(0, 10)
      this.emitFile({ type: 'asset', fileName: 'robots.txt', source: robotsTxt(origin) })
      this.emitFile({ type: 'asset', fileName: 'sitemap.xml', source: sitemapXml(origin, lastmod) })
    },
  }
}
