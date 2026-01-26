<script setup lang="ts">
definePageMeta({
  layout: false,
})

// Reactive state for interactive components
const buttonLoading = ref(false)
const inputValue = ref('')
const textareaValue = ref('')
const switchValue = ref(false)
const switchLoading = ref(false)
const sliderValue = ref(50)
const activeTab = ref('tab1')
const secondaryTab = ref('tabA')
const drawerOpen = ref(false)

// Sample data for tabs
const primaryTabs = [
  { key: 'tab1', label: 'First', value: 'tab1' },
  { key: 'tab2', label: 'Second', value: 'tab2' },
  { key: 'tab3', label: 'Third', value: 'tab3' },
]

const secondaryTabs = [
  { key: 'tabA', label: 'Option A', value: 'tabA' },
  { key: 'tabB', label: 'Option B', value: 'tabB' },
]

// Sample image data
const sampleImage = {
  url: 'https://picsum.photos/400/300',
  width: 400,
  height: 300,
  blurhash: 'LEHV6nWB2yk8pyo0adR*.7kCMdnj',
}

// Button loading demo
function simulateLoading() {
  buttonLoading.value = true
  setTimeout(() => {
    buttonLoading.value = false
  }, 2000)
}
</script>

<template>
  <div class="min-h-screen bg-background-default text-text-default">
    <div class="max-w-2xl mx-auto p-6 space-y-8">
      <header class="space-y-2">
        <h1 class="text-heading-1">Design System</h1>
        <p class="text-text-muted">Component library showcase</p>
      </header>

      <!-- Form Controls -->
      <section class="space-y-6">
        <h2 class="text-heading-2 border-b border-border-default pb-2">
          Form Controls
        </h2>

        <!-- DesignButton -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignButton</h3>

          <div class="space-y-3">
            <p class="text-caption text-text-hint">Variants</p>
            <div class="flex flex-wrap gap-3">
              <DesignButton variant="primary">Primary</DesignButton>
              <DesignButton variant="secondary">Secondary</DesignButton>
              <DesignButton variant="tertiary">Tertiary</DesignButton>
            </div>
          </div>

          <div class="space-y-3">
            <p class="text-caption text-text-hint">Sizes</p>
            <div class="flex flex-wrap items-center gap-3">
              <DesignButton size="small">Small</DesignButton>
              <DesignButton size="medium">Medium</DesignButton>
              <DesignButton size="large">Large</DesignButton>
            </div>
          </div>

          <div class="space-y-3">
            <p class="text-caption text-text-hint">States</p>
            <div class="flex flex-wrap items-center gap-3">
              <DesignButton :loading="buttonLoading" @click="simulateLoading">
                {{ buttonLoading ? 'Loading...' : 'Click to load' }}
              </DesignButton>
              <DesignButton disabled>Disabled</DesignButton>
            </div>
          </div>
        </div>

        <!-- DesignIconButton -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignIconButton</h3>

          <div class="space-y-3">
            <p class="text-caption text-text-hint">Variants</p>
            <div class="flex flex-wrap gap-3">
              <DesignIconButton icon="IconClose" variant="primary" />
              <DesignIconButton icon="IconClose" variant="secondary" />
              <DesignIconButton icon="IconClose" variant="tertiary" />
            </div>
          </div>

          <div class="space-y-3">
            <p class="text-caption text-text-hint">Sizes</p>
            <div class="flex flex-wrap items-center gap-3">
              <DesignIconButton icon="IconChevronRight" size="small" />
              <DesignIconButton icon="IconChevronRight" size="medium" />
              <DesignIconButton icon="IconChevronRight" size="large" />
            </div>
          </div>
        </div>

        <!-- DesignInput -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignInput</h3>

          <div class="space-y-3">
            <DesignInput v-model="inputValue" label="With label" />
            <DesignInput v-model="inputValue" />
            <p class="text-caption text-text-hint">
              Value: {{ inputValue || '(empty)' }}
            </p>
          </div>
        </div>

        <!-- DesignTextarea -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignTextarea</h3>

          <div class="space-y-3">
            <DesignTextarea
              v-model="textareaValue"
              label="Description"
              placeholder="Enter your text here..."
              :maxlength="200"
            />
            <p class="text-caption text-text-hint">
              Value: {{ textareaValue || '(empty)' }}
            </p>
          </div>
        </div>

        <!-- DesignSwitch -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignSwitch</h3>

          <div class="space-y-3">
            <div class="flex items-center gap-4">
              <span class="text-caption text-text-hint">Default</span>
              <DesignSwitch v-model="switchValue" />
              <span class="text-tiny">{{ switchValue ? 'On' : 'Off' }}</span>
            </div>
            <div class="flex items-center gap-4">
              <span class="text-caption text-text-hint">Loading</span>
              <DesignSwitch v-model="switchLoading" loading />
            </div>
            <div class="flex items-center gap-4">
              <span class="text-caption text-text-hint">Disabled</span>
              <DesignSwitch :model-value="true" disabled />
            </div>
          </div>
        </div>

        <!-- DesignSlider -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignSlider</h3>

          <div class="space-y-3">
            <DesignSlider v-model="sliderValue" :min="0" :max="100" :step="1" />
            <p class="text-caption text-text-hint">Value: {{ sliderValue }}</p>
          </div>
        </div>
      </section>

      <!-- Layout/Container -->
      <section class="space-y-6">
        <h2 class="text-heading-2 border-b border-border-default pb-2">
          Layout / Container
        </h2>

        <!-- DesignCard -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignCard</h3>

          <DesignCard>
            <div class="p-4">
              <p class="text-label">Card Content</p>
              <p class="text-caption text-text-hint">
                This is sample content inside a DesignCard component.
              </p>
            </div>
          </DesignCard>
        </div>

        <!-- DesignPanel -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignPanel</h3>

          <DesignPanel>
            <div class="p-3 space-y-2">
              <p class="text-label">Panel Content</p>
              <p class="text-caption text-text-hint">
                This is sample content inside a DesignPanel component.
              </p>
            </div>
          </DesignPanel>
        </div>

        <!-- DesignDrawer -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignDrawer</h3>

          <DesignButton variant="secondary" @click="drawerOpen = true">
            Open Drawer
          </DesignButton>

          <DesignDrawer v-model:open="drawerOpen" title="Sample Drawer">
            <template #content>
              <DesignPanel>
                <div class="p-4 space-y-4">
                  <p class="text-label">Drawer Content</p>
                  <p class="text-caption text-text-hint">
                    This is sample content inside the drawer. You can put any
                    components here.
                  </p>
                  <DesignButton @click="drawerOpen = false">
                    Close Drawer
                  </DesignButton>
                </div>
              </DesignPanel>
            </template>
          </DesignDrawer>
        </div>
      </section>

      <!-- Navigation -->
      <section class="space-y-6">
        <h2 class="text-heading-2 border-b border-border-default pb-2">
          Navigation
        </h2>

        <!-- DesignTabs -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignTabs</h3>

          <div class="space-y-3">
            <p class="text-caption text-text-hint">Primary variant</p>
            <DesignTabs
              v-model="activeTab"
              :tabs="primaryTabs"
              variant="primary"
            />
            <p class="text-tiny text-text-hint">Selected: {{ activeTab }}</p>
          </div>

          <div class="space-y-3">
            <p class="text-caption text-text-hint">Secondary variant</p>
            <DesignTabs
              v-model="secondaryTab"
              :tabs="secondaryTabs"
              variant="secondary"
            />
            <p class="text-tiny text-text-hint">Selected: {{ secondaryTab }}</p>
          </div>
        </div>
      </section>

      <!-- Media -->
      <section class="space-y-6">
        <h2 class="text-heading-2 border-b border-border-default pb-2">
          Media
        </h2>

        <!-- DesignImage -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignImage</h3>

          <div class="space-y-3">
            <p class="text-caption text-text-hint">With blurhash placeholder</p>
            <div class="w-64 h-48 rounded-card overflow-hidden">
              <DesignImage :image="sampleImage" alt="Sample image" />
            </div>
          </div>

          <div class="space-y-3">
            <p class="text-caption text-text-hint">Without image (skeleton)</p>
            <div class="w-64 h-48 rounded-card overflow-hidden">
              <DesignImage alt="No image" />
            </div>
          </div>
        </div>
      </section>

      <!-- Loading -->
      <section class="space-y-6">
        <h2 class="text-heading-2 border-b border-border-default pb-2">
          Loading
        </h2>

        <!-- DesignSkeleton -->
        <div class="space-y-4">
          <h3 class="text-label text-text-muted">DesignSkeleton</h3>

          <div class="space-y-3">
            <DesignSkeleton class="h-4 w-3/4 bg-background-raised" />
            <DesignSkeleton class="h-4 w-1/2 bg-background-raised" />
            <DesignSkeleton class="h-20 w-full bg-background-raised" />
            <div class="flex gap-3">
              <DesignSkeleton
                class="h-10 w-10 rounded-full bg-background-raised"
              />
              <div class="flex-1 space-y-2">
                <DesignSkeleton class="h-4 w-1/3 bg-background-raised" />
                <DesignSkeleton class="h-3 w-2/3 bg-background-raised" />
              </div>
            </div>
          </div>
        </div>
      </section>

      <footer class="text-center text-caption text-text-hint py-8">
        Wayfarer Design System
      </footer>
    </div>
  </div>
</template>
