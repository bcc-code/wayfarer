// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { ref } from 'vue'
import AdminQuizSelector from '../../layers/admin/app/components/admin/achievement/AdminQuizSelector.vue'

const quizzes = ref<
  { quizzes: { edges: { node: { id: string; name: string } }[] } } | undefined
>({
  quizzes: {
    edges: [
      { node: { id: 'QZ1', name: 'Quiz 1' } },
      { node: { id: 'QZ2', name: 'Quiz 2' } },
    ],
  },
})
const fetching = ref(false)

mockNuxtImport('useAdminProjectQuizzesQuery', () => () => ({
  data: quizzes,
  fetching,
  error: ref(undefined),
}))

mockNuxtImport('useAuthReady', () => () => ({ isAuthReady: ref(true) }))

const mount = (props: Record<string, unknown> = {}) =>
  mountSuspended(AdminQuizSelector, {
    props: { projectId: 'PR1', requireCompletion: false, ...props },
  })

const select = (wrapper: Awaited<ReturnType<typeof mount>>) =>
  wrapper.findComponent({ name: 'USelect' })

describe('AdminQuizSelector', () => {
  it('lists the project quizzes in Norwegian', async () => {
    const wrapper = await mount()

    expect(select(wrapper).props('placeholder')).toBe('Velg quiz...')
    expect(select(wrapper).props('items')).toEqual([
      { value: 'QZ1', label: 'Quiz 1' },
      { value: 'QZ2', label: 'Quiz 2' },
    ])
  })

  // `help` is a UFormField prop, so on a UCheckbox it rendered nothing.
  it('describes both requirements on the checkboxes themselves', async () => {
    const wrapper = await mount()

    const descriptions = wrapper
      .findAllComponents({ name: 'UCheckbox' })
      .map((box) => box.props('description'))
    expect(descriptions.filter(Boolean)).toHaveLength(2)
    expect(wrapper.text()).toContain('fullfører hele quizen')
  })

  // Neither requirement set means everyone who submits earns it.
  it('says when the achievement is evaluated', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('når deltakeren leverer quizen')
    expect(wrapper.text()).toContain('uansett resultat')
  })

  // A quiz belongs to a quiz challenge; with none there is nothing to pick.
  it('explains an empty project instead of offering an empty select', async () => {
    quizzes.value = { quizzes: { edges: [] } }
    const wrapper = await mount()

    expect(select(wrapper).props('placeholder')).toBe(
      'Ingen quizer i dette prosjektet',
    )
    expect(select(wrapper).props('disabled')).toBe(true)
    expect(wrapper.text()).toContain('Opprett en quiz-utfordring')

    quizzes.value = {
      quizzes: { edges: [{ node: { id: 'QZ1', name: 'Quiz 1' } }] },
    }
  })

  it('only shows the score field once the requirement is on', async () => {
    const off = await mount()
    expect(off.text()).not.toContain('Minste poengandel')

    const on = await mount({ minScorePercentage: 80 })
    expect(on.text()).toContain('Minste poengandel')
    expect(on.findComponent({ name: 'UInput' }).props('modelValue')).toBe(80)
  })
})
