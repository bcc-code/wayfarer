// @vitest-environment nuxt
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { flushPromises } from '@vue/test-utils'
import AdminChallengeForm from '../../layers/admin/app/components/admin/challenge/AdminChallengeForm.vue'
import { ChallengeType } from '../../app/api/generated'

const confirm = vi.fn(() => Promise.resolve(true))
mockNuxtImport('useConfirm', () => () => ({ confirm }))

const leaveGuards: (() => unknown)[] = []
mockNuxtImport('onBeforeRouteLeave', () => (guard: () => unknown) => {
  leaveGuards.push(guard)
})

const base = {
  name: 'Dagens oppgave',
  buttonText: 'Start',
  submitLabel: 'Lagre',
}

const simple = {
  type: ChallengeType.Simple,
  name: base.name,
  buttonText: base.buttonText,
}

async function submit(wrapper: {
  find: (s: string) => { trigger: (e: string) => Promise<void> }
}) {
  await wrapper.find('form').trigger('submit.prevent')
  await flushPromises()
}

beforeEach(() => {
  confirm.mockClear()
  leaveGuards.length = 0
  vi.useFakeTimers()
  vi.setSystemTime(new Date('2026-06-15T12:00:00'))
})

afterEach(() => {
  vi.useRealTimers()
})

describe('AdminChallengeForm', () => {
  // Quiz is what organisers come here to make; the other types are rarer and
  // read as jargon.
  it('starts a new challenge as a quiz', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: { submitLabel: base.submitLabel },
    })

    expect(wrapper.emitted('update:type')?.at(-1)).toEqual([ChallengeType.Quiz])
  })

  it('keeps the existing type when editing', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: { submitLabel: base.submitLabel, initialData: simple },
    })

    expect(wrapper.emitted('update:type')?.at(-1)).toEqual([
      ChallengeType.Simple,
    ])
  })

  // A quiz is shown to whoever has quiz session access, so the visibility
  // control would be a lie.
  it('offers no visibility choice for a quiz', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: { submitLabel: base.submitLabel },
    })

    expect(wrapper.text()).toContain('quiz-sesjon')
    expect(wrapper.text()).not.toContain('Hvem ser utfordringen')
  })

  it('offers the visibility choice for every other type', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: { submitLabel: base.submitLabel, initialData: simple },
    })

    expect(wrapper.text()).toContain('Hvem ser utfordringen')
    expect(wrapper.text()).not.toContain('quiz-sesjon')
  })

  // The old form left `visibleAt` empty by default, which silently meant
  // "nobody but enrolled users ever sees this".
  it('makes a new challenge visible to everyone', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: { submitLabel: base.submitLabel },
    })

    // As an organiser would: pick a type, fill the two required fields, save.
    const selects = wrapper.findAllComponents({ name: 'USelect' })
    await selects[0]?.vm.$emit('update:modelValue', ChallengeType.Simple)
    const inputs = wrapper.findAllComponents({ name: 'UInput' })
    await inputs[0]?.vm.$emit('update:modelValue', base.name)
    await inputs[1]?.vm.$emit('update:modelValue', base.buttonText)

    await submit(wrapper)

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      visibleAt: '2026-06-15T12:00',
    })
  })

  // An existing challenge is read as it is stored: an empty timestamp there is
  // a deliberate QR-only challenge, not a missing default.
  it('keeps an enrolled-only challenge enrolled-only', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: {
        submitLabel: base.submitLabel,
        initialData: { ...simple, visibleAt: undefined },
      },
    })

    await submit(wrapper)

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      visibleAt: undefined,
    })
  })

  it('does not move the date an already visible challenge went live', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: {
        submitLabel: base.submitLabel,
        initialData: { ...simple, visibleAt: '2026-01-02T08:30' },
      },
    })

    await submit(wrapper)

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      visibleAt: '2026-01-02T08:30',
    })
  })

  it('leaves a quiz timestamp untouched', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: {
        submitLabel: base.submitLabel,
        initialData: {
          type: ChallengeType.Quiz,
          name: base.name,
          buttonText: base.buttonText,
          visibleAt: undefined,
        },
      },
    })

    await submit(wrapper)

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      visibleAt: undefined,
    })
  })

  // The field stores HTML, so it should not be a raw HTML textarea.
  it('edits the description as rich text', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: { submitLabel: base.submitLabel, initialData: simple },
    })

    expect(wrapper.findComponent({ name: 'UEditor' }).exists()).toBe(true)
    expect(wrapper.text()).not.toContain('Støtter HTML-formatering')
  })

  // Clearing the editor leaves an empty paragraph, which would render as a gap.
  it('saves a cleared description as empty, not as an empty paragraph', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: {
        submitLabel: base.submitLabel,
        initialData: { ...simple, description: '<p></p>' },
      },
    })

    await submit(wrapper)

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      description: '',
    })
  })

  it('keeps a description that has content', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: {
        submitLabel: base.submitLabel,
        initialData: { ...simple, description: '<p>Les <b>Matteus 5</b></p>' },
      },
    })

    await submit(wrapper)

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      description: '<p>Les <b>Matteus 5</b></p>',
    })
  })

  // Half an hour of writing a challenge, thrown away by a misclick in the nav.
  it('asks before leaving with unsaved changes', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: { submitLabel: base.submitLabel, initialData: simple },
    })

    const inputs = wrapper.findAllComponents({ name: 'UInput' })
    await inputs[0]?.vm.$emit('update:modelValue', 'Noe helt annet')

    await leaveGuards[0]?.()
    expect(confirm).toHaveBeenCalled()
  })

  // The edit page fills the form in when its query resolves. That is not an
  // edit, and warning about it would train people to click through the dialog.
  it('does not ask when the form only loaded its data', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: { submitLabel: base.submitLabel },
    })

    await wrapper.setProps({ initialData: simple })
    await flushPromises()

    await expect(leaveGuards[0]?.()).resolves.toBe(true)
    expect(confirm).not.toHaveBeenCalled()
  })

  // Publishing time is always "now" in practice; it was only ever noise.
  it('no longer asks for a publishing time', async () => {
    const wrapper = await mountSuspended(AdminChallengeForm, {
      props: { submitLabel: base.submitLabel, initialData: simple },
    })

    expect(wrapper.text()).not.toContain('Publiseringstidspunkt')
  })
})
