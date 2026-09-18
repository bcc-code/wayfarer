// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { defineComponent, h } from 'vue'

/**
 * `test/component/setup.ts` stubs posthog-js. Without the stub the
 * `posthog-client` plugin calls the real `posthog.init()` during app init,
 * which throws asynchronously in this environment ("t.addEventListener is not
 * a function"). Every test still passes and the run still exits non-zero, so
 * the only symptom is a red CI job with a green test list.
 *
 * That makes the stub invisible when it works and invisible when it silently
 * stops being applied — hence this test. It asserts the plugin handed the app
 * our stub, not the real client.
 */
describe('posthog is stubbed in component tests', () => {
  it('provides the stub as $posthog', async () => {
    let injected: unknown
    const Probe = defineComponent({
      setup() {
        injected = useNuxtApp().$posthog()
        return () => h('div')
      },
    })
    await mountSuspended(Probe)

    expect(injected).toBeDefined()
    expect((injected as { __isTestStub?: boolean }).__isTestStub).toBe(true)
  })

  it('never loaded the real client', async () => {
    const posthog = (await import('posthog-js')).default as {
      __isTestStub?: boolean
      __loaded?: boolean
    }
    expect(posthog.__isTestStub).toBe(true)
    // The plugin returns early when `__loaded` is truthy and then never
    // provides `$posthog`, which `useAnalytics` calls.
    expect(posthog.__loaded).toBe(false)
  })
})
