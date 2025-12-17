<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import { z } from 'zod'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

const schema = z.object({
  key: z
    .string()
    .min(1, 'Key is required')
    .regex(
      /^[a-z0-9_-]+$/,
      'Key must be lowercase alphanumeric with underscores or hyphens',
    ),
  title: z.string().min(1, 'Title is required'),
  shortText: z.string().optional(),
  body: z.string().min(1, 'Body is required'),
  url: z.url('Must be a valid URL').optional().or(z.literal('')),
  publishNow: z.boolean().default(false),
  isRemote: z.boolean().default(false),
  managedBy: z.string().optional(),
})

type Schema = z.infer<typeof schema>

const state = reactive<Schema>({
  key: '',
  title: '',
  shortText: '',
  body: '',
  url: '',
  publishNow: false,
  isRemote: false,
  managedBy: '',
})

const { executeMutation: createConsent } = useCreateConsentMutation()
const toast = useToast()

async function handleSubmit(event: FormSubmitEvent<Schema>) {
  if (!event.data) return

  const result = await createConsent({
    key: event.data.key,
    title: event.data.title,
    shortText: event.data.shortText || undefined,
    body: event.data.body,
    url: event.data.url || undefined,
    publishedAt: event.data.publishNow ? new Date().toISOString() : undefined,
    isRemote: event.data.isRemote || undefined,
    managedBy: event.data.managedBy || undefined,
  })

  if (result.error) {
    toast.add({
      title: 'Failed to create consent',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Consent created',
    color: 'success',
  })

  const consentId = result.data?.createConsent.id
  if (consentId) {
    navigateTo({
      name: 'admin-consents-consentId',
      params: { consentId },
    })
  } else {
    navigateTo({ name: 'admin-consents' })
  }
}
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            { label: 'Consents', to: { name: 'admin-consents' } },
            { label: 'New Consent' },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <div class="mx-auto max-w-2xl">
        <h1 class="mb-6 text-3xl font-bold">Create New Consent</h1>

        <UForm :state :schema @submit="handleSubmit">
          <div class="space-y-6">
            <UFormField name="key" label="Key" required>
              <UInput
                v-model="state.key"
                class="w-full"
                placeholder="privacy_policy"
                description="Unique identifier for this consent. Use lowercase with underscores."
              />
            </UFormField>

            <UFormField name="title" label="Title" required>
              <UInput
                v-model="state.title"
                class="w-full"
                placeholder="Privacy Policy"
              />
            </UFormField>

            <UFormField name="shortText" label="Short Text">
              <UTextarea
                v-model="state.shortText"
                class="w-full"
                autoresize
                placeholder="A brief description shown to users before they read the full consent"
              />
            </UFormField>

            <UFormField name="body" label="Body (Markdown)" required>
              <UTextarea
                v-model="state.body"
                class="w-full font-mono"
                :rows="10"
                autoresize
                placeholder="# Privacy Policy&#10;&#10;Your markdown content here..."
              />
            </UFormField>

            <UFormField name="url" label="External URL">
              <UInput
                v-model="state.url"
                class="w-full"
                type="url"
                placeholder="https://example.com/privacy-policy"
              />
            </UFormField>

            <UDivider />

            <div class="space-y-4">
              <UFormField name="isRemote">
                <UCheckbox v-model="state.isRemote" label="Remote Management" />
                <template #description>
                  Enable if this consent is managed by an external system
                </template>
              </UFormField>

              <UFormField
                v-if="state.isRemote"
                name="managedBy"
                label="Managed By"
              >
                <UInput
                  v-model="state.managedBy"
                  class="w-full"
                  placeholder="External system name"
                />
              </UFormField>

              <UFormField name="publishNow">
                <UCheckbox
                  v-model="state.publishNow"
                  label="Publish immediately"
                />
                <template #description>
                  If unchecked, the consent will be saved as a draft
                </template>
              </UFormField>
            </div>

            <div class="flex justify-end gap-3 pt-4">
              <UButton variant="ghost" :to="{ name: 'admin-consents' }">
                Cancel
              </UButton>
              <UButton type="submit"> Create Consent </UButton>
            </div>
          </div>
        </UForm>
      </div>
    </UContainer>
  </div>
</template>
