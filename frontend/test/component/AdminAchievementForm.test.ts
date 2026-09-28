// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminAchievementForm from '../../layers/admin/app/components/admin/achievement/AdminAchievementForm.vue'

const initialData = {
  name: 'Første steg',
  descriptionPending: 'Gjør oppgaven',
  descriptionCompleted: 'Godt jobbet',
  notificationText: 'Du fikk en utmerkelse',
  imagePending: '',
  imageCompleted: '',
  points: 25,
  hidden: false,
  awardableFrom: '2026-02-09T14:29',
}

const mount = (props: Record<string, unknown> = {}) =>
  mountSuspended(AdminAchievementForm, {
    props: {
      projectId: 'PR1',
      submitLabel: 'Lagre endringer',
      ...props,
    },
  })

const sectionTitles = (wrapper: Awaited<ReturnType<typeof mount>>) =>
  wrapper
    .findAllComponents({ name: 'AdminSection' })
    .map((section) => section.props('title'))

describe('AdminAchievementForm', () => {
  it('groups the fields into sections', async () => {
    const wrapper = await mount({
      initialData,
      isEditMode: true,
      achievementType: 'SIMPLE',
    })

    expect(sectionTitles(wrapper)).toEqual([
      'Innhold',
      'Varsling',
      'Poeng og synlighet',
    ])
  })

  // The type is a badge in the page header in edit mode; offering a disabled
  // selector invited clicking something that cannot change.
  it('offers the type selector only when creating', async () => {
    const creating = await mount()
    expect(sectionTitles(creating)).toContain('Utmerkelsestype')

    const editing = await mount({
      initialData,
      isEditMode: true,
      achievementType: 'SIMPLE',
    })
    expect(sectionTitles(editing)).not.toContain('Utmerkelsestype')
  })

  // Every schema field needs a control; the notification text lost its field in
  // an earlier draft of this restructure.
  it('keeps a control for every field the schema requires', async () => {
    const wrapper = await mount({
      initialData,
      isEditMode: true,
      achievementType: 'SIMPLE',
    })

    const text = wrapper.text()
    for (const label of [
      'Navn',
      'Beskrivelse (ikke oppnådd)',
      'Beskrivelse (oppnådd)',
      'Varslingstekst',
      'Bilde (ikke oppnådd)',
      'Bilde (oppnådd)',
      'Poeng for utmerkelsen',
      'Skjult',
      'Tidligste tildelings-tidspunkt',
    ]) {
      expect(text, label).toContain(label)
    }
  })

  it('uses the shared datetime field, not a raw datetime-local input', async () => {
    const wrapper = await mount({
      initialData,
      isEditMode: true,
      achievementType: 'SIMPLE',
    })

    expect(
      wrapper.findComponent({ name: 'AdminDateTimeField' }).props('modelValue'),
    ).toBe('2026-02-09T14:29')
    // Not `input[type=datetime-local]`: reka-ui's DateField renders one of
    // those itself, hidden, for form integration.
    expect(wrapper.findComponent({ name: 'UInputDate' }).exists()).toBe(true)
  })

  // Saving is the page's primary action.
  it('leaves only the submit button solid', async () => {
    const wrapper = await mount({
      initialData,
      isEditMode: true,
      achievementType: 'SIMPLE',
    })

    const solid = wrapper
      .findAllComponents({ name: 'UButton' })
      .filter((button) => (button.props('variant') ?? 'solid') === 'solid')
    expect(solid.map((button) => button.text())).toEqual(['Lagre endringer'])
  })
})
