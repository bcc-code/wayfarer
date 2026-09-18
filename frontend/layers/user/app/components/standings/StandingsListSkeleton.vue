<template>
  <div>
    <!-- Title -->
    <div
      class="p-medium gap-medium mb-list-section-gap flex flex-col items-center"
    >
      <DesignSkeleton class="h-9 w-36 rounded bg-background-raised" />
    </div>

    <!-- Sub-tabs -->
    <div
      class="rounded-navigation border-border-default mb-list-section-gap flex border p-navigation-inset"
    >
      <div
        class="flex grow flex-col items-center gap-1 py-2 bg-background-raised rounded-navigation-inset gradient-border relative"
      >
        <DesignSkeleton class="size-6.5 rounded" />
        <DesignSkeleton class="h-3 w-10 rounded" />
      </div>
      <div class="flex grow flex-col items-center gap-1 py-2">
        <DesignSkeleton class="size-6.5 rounded bg-background-raised" />
        <DesignSkeleton class="h-3 w-10 rounded bg-background-raised" />
      </div>
    </div>
  </div>
</template>
