<script setup lang="ts"></script>

<template>
  <UModal
    :ui="{ content: 'bg-background-default' }"
    :transition="false"
    fullscreen
    modal
  >
    <slot />
    <template #content="{ close }">
      <PageLayout :title="$t('pages.getPoints')">
        <template #action>
          <DesignIconButton icon="lucide:x" @click="close" />
        </template>
      </PageLayout>
    </template>
  </UModal>
</template>
