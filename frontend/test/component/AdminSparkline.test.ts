// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminSparkline from '../../layers/admin/app/components/admin/AdminSparkline.vue'

const series = [
  { date: '2026-09-19', value: 0 },
  { date: '2026-09-20', value: 120 },
  { date: '2026-09-21', value: 60 },
]

describe('AdminSparkline', () => {
  it('draws a bar per point plus a baseline', async () => {
    const wrapper = await mountSuspended(AdminSparkline, {
      props: { points: series, label: 'Poeng' },
    })

    // 3 bars + 1 baseline.
    expect(wrapper.findAll('svg rect')).toHaveLength(4)
  })

  // An all-zero window is normal for a project between bursts of activity. A
  // flat series still occupies the plot's full height, so it rendered as a tall
  // empty box with a stray baseline adrift in it — a sentence is the honest
  // answer, and it keeps the tile compact.
  it('shows a sentence instead of a chart when nothing happened', async () => {
    const wrapper = await mountSuspended(AdminSparkline, {
      props: {
        points: series.map((point) => ({ ...point, value: 0 })),
        label: 'Poeng',
        emptyLabel: 'Ingen aktivitet siste 14 dager',
      },
    })

    expect(wrapper.find('svg').exists()).toBe(false)
    expect(wrapper.text()).toContain('Ingen aktivitet siste 14 dager')
  })

  it('shows the chart as soon as any day has activity', async () => {
    const wrapper = await mountSuspended(AdminSparkline, {
      props: {
        points: [
          { date: '2026-09-20', value: 0 },
          { date: '2026-09-21', value: 1 },
        ],
        label: 'Poeng',
      },
    })

    expect(wrapper.find('svg').exists()).toBe(true)
  })

  // Tailwind's `sr-only` sets no `display`, so on a <table> the <caption>
  // escapes the clip rect and renders as visible text. The class must sit on a
  // wrapping div.
  it('keeps the table view clipped by wrapping it, not classing the table', async () => {
    const wrapper = await mountSuspended(AdminSparkline, {
      props: { points: series, label: 'Poeng' },
    })

    expect(wrapper.find('table.sr-only').exists()).toBe(false)
    expect(wrapper.find('.sr-only table').exists()).toBe(true)
  })

  it('lists every day in the table view', async () => {
    const wrapper = await mountSuspended(AdminSparkline, {
      props: { points: series, label: 'Poeng siste 14 dager' },
    })

    const table = wrapper.find('.sr-only table')
    expect(table.text()).toContain('Poeng siste 14 dager')
    expect(table.findAll('tbody tr')).toHaveLength(3)
    expect(table.text()).toContain('120')
  })

  it('applies the caller format to values', async () => {
    const wrapper = await mountSuspended(AdminSparkline, {
      props: {
        points: series,
        label: 'Poeng',
        format: (value: number) => `${value} p`,
      },
    })

    expect(wrapper.find('.sr-only table').text()).toContain('120 p')
  })

  // The plot is hidden from assistive tech because the table is the accessible
  // path; a role="img" summary next to a full table would say it twice.
  it('hides the plot from assistive technology', async () => {
    const wrapper = await mountSuspended(AdminSparkline, {
      props: { points: series, label: 'Poeng' },
    })

    expect(wrapper.find('svg').attributes('aria-hidden')).toBe('true')
  })

  it('treats an empty series as no activity', async () => {
    const wrapper = await mountSuspended(AdminSparkline, {
      props: { points: [], label: 'Poeng' },
    })

    expect(wrapper.find('svg').exists()).toBe(false)
    expect(wrapper.find('.sr-only').exists()).toBe(false)
  })
})
