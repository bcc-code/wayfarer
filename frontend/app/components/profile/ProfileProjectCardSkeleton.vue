<template>
  <DesignCard>
    <div class="p-default gap-medium flex flex-col">
      <!-- Project name -->
      <div class="flex justify-center">
        <DesignSkeleton class="h-5 w-32 rounded" />
      </div>

      <!-- Score/rank stats -->
      <div class="divide-border-default grid grid-cols-2 divide-x py-2">
        <div class="flex flex-col items-center gap-1">
          <DesignSkeleton class="h-6 w-16 rounded" />
          <DesignSkeleton class="h-4 w-12 rounded" />
        </div>
        <div class="flex flex-col items-center gap-1">
          <DesignSkeleton class="h-6 w-10 rounded" />
          <DesignSkeleton class="h-4 w-12 rounded" />
        </div>
      </div>

      <!-- Buttons -->
      <div class="gap-medium grid grid-cols-2">
        <DesignSkeleton class="h-12 w-full rounded-button-medium!" />
        <DesignSkeleton class="h-12 w-full rounded-button-medium!" />
      </div>
    </div>

    <!-- Achievement badges -->
    <div class="p-medium gap-medium grid grid-cols-4">
      <DesignSkeleton class="aspect-square w-full rounded-full" />
      <DesignSkeleton class="aspect-square w-full rounded-full" />
      <DesignSkeleton class="aspect-square w-full rounded-full" />
      <DesignSkeleton class="aspect-square w-full rounded-full" />
      <DesignSkeleton class="aspect-square w-full rounded-full" />
      <DesignSkeleton class="aspect-square w-full rounded-full" />
      <DesignSkeleton class="aspect-square w-full rounded-full" />
      <DesignSkeleton class="aspect-square w-full rounded-full" />
      <DesignSkeleton class="aspect-square w-full rounded-full" />
      <DesignSkeleton class="aspect-square w-full rounded-full" />
    </div>
  </DesignCard>
</template>
