// @vitest-environment nuxt
import { describe, it, expect, vi } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { ref, computed } from 'vue'
import AdminListView from '../../layers/admin/app/components/admin/AdminListView.vue'
import type { UsePaginationReturn } from '../../layers/admin/app/composables/usePagination'

function paginationStub(
  overrides: Partial<{
    range: { start: number; end: number; total: number } | null
    isFirstPage: boolean
    isLastPage: boolean
  }> = {},
) {
  const nextPage = vi.fn()
  const previousPage = vi.fn()
  const setPageSize = vi.fn()

  const stub = {
    variables: ref({}),
    pageInfo: ref(null),
    totalCount: ref(overrides.range?.total ?? 0),
    pageSize: ref(15),
    range: computed(() =>
      overrides.range === undefined
        ? { start: 1, end: 15, total: 100 }
        : overrides.range,
    ),
    isFirstPage: computed(() => overrides.isFirstPage ?? true),
    isLastPage: computed(() => overrides.isLastPage ?? false),
    nextPage,
    previousPage,
    firstPage: vi.fn(),
    updateConnection: vi.fn(),
    reset: vi.fn(),
    setPageSize,
  } as unknown as UsePaginationReturn

  return { stub, nextPage, previousPage, setPageSize }
}

describe('AdminListView', () => {
  it('renders the table passed in its default slot', async () => {
    const { stub } = paginationStub()
    const wrapper = await mountSuspended(AdminListView, {
      props: { pagination: stub },
      slots: { default: () => 'TABLE HERE' },
    })

    expect(wrapper.text()).toContain('TABLE HERE')
  })

  // totalCount was fetched on every query and used only as a boolean gate.
  it('shows the position in the result set', async () => {
    const { stub } = paginationStub({
      range: { start: 16, end: 30, total: 13477 },
    })
    const wrapper = await mountSuspended(AdminListView, {
      props: { pagination: stub, itemLabel: 'brukere' },
    })

    const text = wrapper.text().replace(/\s+/g, ' ')
    expect(text).toContain('16')
    expect(text).toContain('30')
    expect(text).toContain('brukere')
  })

  it('renders no footer when there is nothing to page through', async () => {
    const { stub } = paginationStub({ range: null })
    const wrapper = await mountSuspended(AdminListView, {
      props: { pagination: stub },
    })

    expect(wrapper.text()).not.toContain('Viser')
  })

  it('disables the controls at the ends of the list', async () => {
    const { stub } = paginationStub({ isFirstPage: true, isLastPage: false })
    const wrapper = await mountSuspended(AdminListView, {
      props: { pagination: stub },
    })

    const buttons = wrapper.findAll('button')
    const previous = buttons.find((b) => b.text().includes('Forrige'))
    const next = buttons.find((b) => b.text().includes('Neste'))

    expect(previous?.attributes('disabled')).toBeDefined()
    expect(next?.attributes('disabled')).toBeUndefined()
  })

  it('steps pages through the pagination it was given', async () => {
    const { stub, nextPage } = paginationStub({
      isFirstPage: false,
      isLastPage: false,
    })
    const wrapper = await mountSuspended(AdminListView, {
      props: { pagination: stub },
    })

    const next = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Neste'))
    await next!.trigger('click')

    expect(nextPage).toHaveBeenCalledOnce()
  })

  it('renders a chip per active filter and can clear one', async () => {
    const { stub } = paginationStub()
    const wrapper = await mountSuspended(AdminListView, {
      props: {
        pagination: stub,
        activeFilters: [{ key: 'churchId', value: 'CH1', label: 'Østfold' }],
      },
    })

    expect(wrapper.text()).toContain('Østfold')

    const clear = wrapper
      .findAll('button')
      .find((b) => b.attributes('aria-label')?.includes('Fjern filter'))
    await clear!.trigger('click')

    expect(wrapper.emitted('clearFilter')?.[0]).toEqual(['churchId'])
  })

  it('offers a reset only while something is filtered', async () => {
    const { stub } = paginationStub()

    const unfiltered = await mountSuspended(AdminListView, {
      props: { pagination: stub },
    })
    expect(unfiltered.text()).not.toContain('Nullstill')

    const filtered = await mountSuspended(AdminListView, {
      props: {
        pagination: stub,
        activeFilters: [{ key: 'churchId', value: 'CH1' }],
      },
    })
    expect(filtered.text()).toContain('Nullstill')
  })

  it('changes page size through the pagination it was given', async () => {
    const { stub, setPageSize } = paginationStub()
    const wrapper = await mountSuspended(AdminListView, {
      props: { pagination: stub },
    })

    const select = wrapper.findComponent({ name: 'USelect' })
    await select.setValue(50)

    expect(setPageSize).toHaveBeenCalledWith(50)
  })

  it('binds the search box two-way', async () => {
    const { stub } = paginationStub()
    const wrapper = await mountSuspended(AdminListView, {
      props: { pagination: stub, search: '' },
    })

    await wrapper.find('input').setValue('sigve')

    expect(wrapper.emitted('update:search')?.at(-1)).toEqual(['sigve'])
  })
})
