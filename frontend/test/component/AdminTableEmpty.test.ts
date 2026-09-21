// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminTableEmpty from '../../layers/admin/app/components/admin/AdminTableEmpty.vue'
import AdminTableLoading from '../../layers/admin/app/components/admin/AdminTableLoading.vue'

describe('AdminTableEmpty', () => {
  it('shows the empty title when nothing is filtered', async () => {
    const wrapper = await mountSuspended(AdminTableEmpty, {
      props: { title: 'Ingen brukere', description: 'Legg til en bruker.' },
    })

    expect(wrapper.text()).toContain('Ingen brukere')
    expect(wrapper.text()).toContain('Legg til en bruker.')
  })

  // Telling someone "no challenges yet" when they have filtered them all away
  // is actively misleading, and leaves them no way back.
  it('says something different when a filter hid everything', async () => {
    const wrapper = await mountSuspended(AdminTableEmpty, {
      props: {
        filtered: true,
        title: 'Ingen utfordringer ennå',
        filteredTitle: 'Ingen utfordringer av denne typen',
      },
    })

    expect(wrapper.text()).toContain('Ingen utfordringer av denne typen')
    expect(wrapper.text()).not.toContain('Ingen utfordringer ennå')
  })

  it('offers a way out only from a filtered miss', async () => {
    const unfiltered = await mountSuspended(AdminTableEmpty, {
      props: { title: 'Ingen brukere' },
    })
    expect(unfiltered.find('button').exists()).toBe(false)

    const filtered = await mountSuspended(AdminTableEmpty, {
      props: { filtered: true, title: 'Ingen brukere' },
    })
    await filtered.find('button').trigger('click')
    expect(filtered.emitted('clear')).toHaveLength(1)
  })

  it('falls back to a generic line when no filtered title is given', async () => {
    const wrapper = await mountSuspended(AdminTableEmpty, {
      props: { filtered: true, title: 'Ingen brukere' },
    })

    expect(wrapper.text()).toContain('Ingenting passer søket')
  })

  // The description describes what an empty list *is*; it does not apply to a
  // list that merely has a filter on it.
  it('hides the description in the filtered case', async () => {
    const wrapper = await mountSuspended(AdminTableEmpty, {
      props: {
        filtered: true,
        title: 'Ingen brukere',
        description: 'Legg til en bruker.',
      },
    })

    expect(wrapper.text()).not.toContain('Legg til en bruker.')
  })

  it('decorates only the genuinely-empty case with an icon', async () => {
    const empty = await mountSuspended(AdminTableEmpty, {
      props: { title: 'Ingen superlag', icon: 'lucide:users' },
    })
    expect(empty.find('.iconify').exists()).toBe(true)

    const filtered = await mountSuspended(AdminTableEmpty, {
      props: { filtered: true, title: 'Ingen superlag', icon: 'lucide:users' },
    })
    expect(filtered.find('.iconify').exists()).toBe(false)
  })
})

describe('AdminTableLoading', () => {
  // `:loading` alone draws a bar over an empty table, which looks the same as
  // an empty list. Skeleton rows say "something is coming".
  it('renders skeleton rows', async () => {
    const wrapper = await mountSuspended(AdminTableLoading, {
      props: { rows: 4 },
    })

    expect(wrapper.findAll('.animate-pulse')).toHaveLength(4)
  })

  it('defaults to a sensible number of rows', async () => {
    const wrapper = await mountSuspended(AdminTableLoading)

    expect(wrapper.findAll('.animate-pulse').length).toBeGreaterThan(0)
  })
})
