/**
 * Tabs: the hero's TV modes. Buttons switch the visible panel and arrow keys move between them.
 *
 * Markup: `[data-tabs]` holding `[role=tab][aria-controls]` buttons and `[role=tabpanel]`s.
 * `data-tabs-autoplay="5000"` also moves to the next tab every 5 seconds, like the TV changing
 * mode by itself; see autoplay.
 */
export function setupTabs(root: ParentNode) {
  for (const widget of root.querySelectorAll<HTMLElement>('[data-tabs]')) {
    const tabs = [...widget.querySelectorAll<HTMLButtonElement>('[role="tab"]')]
    const panelOf = (tab: HTMLButtonElement) =>
      widget.querySelector<HTMLElement>(`#${tab.getAttribute('aria-controls')}`)
    let current = 0
    const select = (index: number) => {
      current = index
      tabs.forEach((tab, i) => {
        const selected = i === index
        tab.setAttribute('aria-selected', String(selected))
        tab.tabIndex = selected ? 0 : -1
        const panel = panelOf(tab)
        if (panel) panel.hidden = !selected
      })
    }
    const step: Record<string, (i: number) => number> = {
      ArrowRight: (i) => (i + 1) % tabs.length,
      ArrowDown: (i) => (i + 1) % tabs.length,
      ArrowLeft: (i) => (i - 1 + tabs.length) % tabs.length,
      ArrowUp: (i) => (i - 1 + tabs.length) % tabs.length,
      Home: () => 0,
      End: () => tabs.length - 1,
    }

    const interval = Number(widget.dataset.tabsAutoplay)
    const stopAutoplay =
      interval > 0
        ? autoplay(widget, interval, () => {
            const next = (current + 1) % tabs.length
            const tab = tabs[next]
            const panel = tab && panelOf(tab)
            return (panel ? preloadImages(panel) : Promise.resolve()).then(() => select(next))
          })
        : () => {}

    tabs.forEach((tab, i) => {
      tab.addEventListener('click', () => {
        stopAutoplay()
        select(i)
      })
      tab.addEventListener('keydown', (e) => {
        const next = step[e.key]?.(i)
        if (next === undefined) return
        e.preventDefault()
        stopAutoplay()
        select(next)
        tabs[next]?.focus()
      })
    })
    select(
      Math.max(
        0,
        tabs.findIndex((tab) => tab.getAttribute('aria-selected') === 'true'),
      ),
    )
  }
}

/**
 * Calls advance every interval ms while the widget can be watched: not while the pointer is over
 * it or focus is in it, the widget is off screen, or the page is in a background tab. Never with
 * prefers-reduced-motion. Returns a stop function: a visitor picking a tab ends it for good, so
 * the page never takes their choice away.
 */
function autoplay(widget: HTMLElement, interval: number, advance: () => Promise<void>) {
  if (window.matchMedia?.('(prefers-reduced-motion: reduce)').matches) return () => {}

  let stopped = false
  let hovered = false
  let focused = false
  let onScreen = true
  let timer: ReturnType<typeof setTimeout> | undefined

  const paused = () => stopped || hovered || focused || !onScreen || document.hidden
  const schedule = () => {
    clearTimeout(timer)
    timer = undefined
    if (paused()) return
    timer = setTimeout(() => {
      if (paused()) return
      void advance().finally(schedule)
    }, interval)
  }

  widget.addEventListener('pointerenter', () => {
    hovered = true
    schedule()
  })
  widget.addEventListener('pointerleave', () => {
    hovered = false
    schedule()
  })
  widget.addEventListener('focusin', () => {
    focused = true
    schedule()
  })
  widget.addEventListener('focusout', (e) => {
    focused = widget.contains(e.relatedTarget as Node | null)
    schedule()
  })
  document.addEventListener('visibilitychange', schedule)
  if ('IntersectionObserver' in window) {
    new IntersectionObserver(([entry]) => {
      onScreen = entry?.isIntersecting ?? true
      schedule()
    }).observe(widget)
  }

  schedule()
  return () => {
    stopped = true
    clearTimeout(timer)
  }
}

/** Waits for a hidden panel's images, so autoplay never shows an empty TV. Errors don't block. */
function preloadImages(panel: HTMLElement): Promise<void> {
  const loads = [...panel.querySelectorAll('img')].map((img) => {
    img.loading = 'eager'
    return img.complete ? Promise.resolve() : img.decode().catch(() => {})
  })
  return Promise.all(loads).then(() => {})
}
