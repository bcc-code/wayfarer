<template>
  <!--
    Admin counterpart to the root `LoadingState`. Same shape; kept separate so
    the shared one can move to the user layer with the design system it belongs
    to. See AdminErrorState for the full reasoning.
  -->
  <div
    class="text-muted flex size-full grow flex-col items-center justify-center gap-4 p-8 text-2xl"
  >
    <UIcon name="svg-spinners:bars-rotate-fade" />
    <slot />
  </div>
</template>
