import { describe, expect, it } from 'vitest'
import { headTags, pagePath, robotsTxt, sitemapXml, siteOrigin } from '../seo'

describe('siteOrigin', () => {
  it('drops the trailing slash and falls back to the production site', () => {
    expect(siteOrigin({ VITE_SITE_URL: ' https://masjid.example/ ' })).toBe(
      'https://masjid.example',
    )
    expect(siteOrigin({ VITE_SITE_URL: '' })).toBe('https://masjidyuk.com')
    expect(siteOrigin({})).toBe('https://masjidyuk.com')
  })
})

describe('pagePath', () => {
  it('maps the built HTML files to their public paths', () => {
    expect(pagePath('/index.html')).toBe('/')
    expect(pagePath('/docs/index.html')).toBe('/docs/')
  })
})

describe('robotsTxt', () => {
  it('keeps crawlers off the API and points at the sitemap', () => {
    expect(robotsTxt('https://masjid.example')).toBe(
      'User-agent: *\nAllow: /\nDisallow: /api/\n\nSitemap: https://masjid.example/sitemap.xml\n',
    )
  })
})

describe('sitemapXml', () => {
  it('lists every page with absolute URLs', () => {
    const xml = sitemapXml('https://masjid.example', '2026-11-27')
    expect(xml).toContain('<loc>https://masjid.example/</loc><lastmod>2026-11-27</lastmod>')
    expect(xml).toContain('<loc>https://masjid.example/docs/</loc>')
  })
})

describe('headTags', () => {
  it('adds canonical, og:url and an absolute share image for the page', () => {
    const tags = headTags({ VITE_SITE_URL: 'https://masjid.example' }, '/docs/')
    expect(tags).toContainEqual(
      expect.objectContaining({
        tag: 'link',
        attrs: { rel: 'canonical', href: 'https://masjid.example/docs/' },
      }),
    )
    expect(tags).toContainEqual(
      expect.objectContaining({
        attrs: { property: 'og:image', content: 'https://masjid.example/screenshots/tv.png' },
      }),
    )
  })

  it('uses the production site by default and adds the Search Console tag only when set', () => {
    expect(headTags({}, '/')).toContainEqual(
      expect.objectContaining({ attrs: { rel: 'canonical', href: 'https://masjidyuk.com/' } }),
    )
    expect(headTags({}, '/')).not.toContainEqual(
      expect.objectContaining({
        attrs: expect.objectContaining({ name: 'google-site-verification' }),
      }),
    )
    expect(headTags({ VITE_GOOGLE_SITE_VERIFICATION: 'abc' }, '/')[0]).toEqual({
      tag: 'meta',
      attrs: { name: 'google-site-verification', content: 'abc' },
      injectTo: 'head',
    })
  })
})
