// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import QuizQuestionHeading from '../../layers/user/app/components/challenges/quiz/QuizQuestionHeading.vue'

const image = {
  url: 'https://cdn.example.test/q.png',
  width: 1200,
  height: 800,
  blurhash: null,
}

describe('QuizQuestionHeading', () => {
  it('shows the question text', async () => {
    const wrapper = await mountSuspended(QuizQuestionHeading, {
      props: { questionText: 'Hvem skrev Romerbrevet?', index: 0, total: 3 },
    })

    expect(wrapper.text()).toContain('Hvem skrev Romerbrevet?')
  })

  it('renders no image when the question has none', async () => {
    const wrapper = await mountSuspended(QuizQuestionHeading, {
      props: { questionText: 'Hvem skrev Romerbrevet?', index: 0, total: 1 },
    })

    expect(wrapper.find('img').exists()).toBe(false)
  })

  // "Above the question" is the whole point of the field, so the order of the
  // two is worth asserting rather than just their presence.
  it('puts the image above the question text', async () => {
    const wrapper = await mountSuspended(QuizQuestionHeading, {
      props: {
        questionText: 'Hvem skrev Romerbrevet?',
        index: 0,
        total: 1,
        image,
      },
    })

    const img = wrapper.find('img')
    expect(img.attributes('src')).toBe(image.url)

    const html = wrapper.html()
    expect(html.indexOf('<img')).toBeLessThan(html.indexOf('<h1'))
  })

  // A box of the image's own shape: `DesignImage` crops to its container, and
  // a question image is often the thing being asked about.
  it('keeps the image at its own aspect ratio', async () => {
    const wrapper = await mountSuspended(QuizQuestionHeading, {
      props: {
        questionText: 'Hvem skrev Romerbrevet?',
        index: 0,
        total: 1,
        image,
      },
    })

    expect(wrapper.find('img').attributes('style')).toContain('aspect-ratio')
  })
})
