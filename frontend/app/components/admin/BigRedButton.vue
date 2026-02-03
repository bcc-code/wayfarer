<script setup lang="ts">
defineProps<{
  label?: string
  loading?: boolean
  onClick?: () => void
}>()
</script>

<template>
  <div class="big-red-button-wrapper">
    <div class="button-base">
      <button
        class="button-dome"
        :class="{ 'is-loading': loading }"
        :disabled="loading"
        @click="onClick"
      >
        <span v-if="loading" class="spinner" />
        <span v-else class="label">{{ label }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.big-red-button-wrapper {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  padding: 24px;
}

.button-base {
  width: 200px;
  height: 200px;
  border-radius: 50%;
  background: linear-gradient(180deg, #94a3b8 0%, #64748b 40%, #475569 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow:
    0 4px 12px rgba(0, 0, 0, 0.3),
    inset 0 2px 4px rgba(255, 255, 255, 0.15),
    inset 0 -2px 6px rgba(0, 0, 0, 0.25);
  position: relative;
}

.button-base::before {
  content: '';
  position: absolute;
  inset: 8px;
  border-radius: 50%;
  background: linear-gradient(180deg, #475569 0%, #334155 100%);
  box-shadow: inset 0 2px 8px rgba(0, 0, 0, 0.4);
}

.button-dome {
  position: relative;
  z-index: 1;
  width: 150px;
  height: 150px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  background: radial-gradient(
    circle at 38% 32%,
    #f87171 0%,
    #ef4444 25%,
    #dc2626 50%,
    #b91c1c 75%,
    #991b1b 100%
  );
  box-shadow:
    0 8px 0 0 #7f1d1d,
    0 10px 20px rgba(0, 0, 0, 0.4),
    inset 0 -4px 8px rgba(0, 0, 0, 0.2);
  color: white;
  font-weight: 800;
  font-size: 16px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  display: flex;
  align-items: center;
  justify-content: center;
  transition:
    transform 0.1s ease,
    box-shadow 0.1s ease,
    background 0.15s ease;
  -webkit-user-select: none;
  user-select: none;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.4);
  transform: translateY(-6px);
}

/* Sheen highlight */
.button-dome::before {
  content: '';
  position: absolute;
  top: 12%;
  left: 20%;
  width: 40%;
  height: 30%;
  border-radius: 50%;
  background: radial-gradient(
    ellipse at center,
    rgba(255, 255, 255, 0.35) 0%,
    rgba(255, 255, 255, 0) 100%
  );
  pointer-events: none;
  transition: opacity 0.15s ease;
}

.button-dome:hover:not(:disabled) {
  background: radial-gradient(
    circle at 38% 32%,
    #fca5a5 0%,
    #f87171 25%,
    #ef4444 50%,
    #dc2626 75%,
    #b91c1c 100%
  );
}

.button-dome:active:not(:disabled) {
  transform: translateY(0);
  box-shadow:
    0 2px 0 0 #7f1d1d,
    0 4px 10px rgba(0, 0, 0, 0.3),
    inset 0 -2px 6px rgba(0, 0, 0, 0.25);
  background: radial-gradient(
    circle at 38% 32%,
    #ef4444 0%,
    #dc2626 25%,
    #b91c1c 50%,
    #991b1b 75%,
    #7f1d1d 100%
  );
}

.button-dome:active:not(:disabled)::before {
  opacity: 0.5;
}

/* Loading state */
.button-dome.is-loading {
  cursor: wait;
  background: radial-gradient(
    circle at 38% 32%,
    #f87171 0%,
    #ef4444 25%,
    #dc2626 50%,
    #b91c1c 75%,
    #991b1b 100%
  );
  animation: pulse-dome 1.5s ease-in-out infinite;
}

@keyframes pulse-dome {
  0%,
  100% {
    box-shadow:
      0 8px 0 0 #7f1d1d,
      0 10px 20px rgba(0, 0, 0, 0.4),
      inset 0 -4px 8px rgba(0, 0, 0, 0.2);
  }
  50% {
    box-shadow:
      0 8px 0 0 #7f1d1d,
      0 10px 30px rgba(239, 68, 68, 0.5),
      inset 0 -4px 8px rgba(0, 0, 0, 0.2);
  }
}

.label {
  position: relative;
  z-index: 1;
  padding: 0 12px;
  text-align: center;
  line-height: 1.2;
}

.spinner {
  position: relative;
  z-index: 1;
  width: 28px;
  height: 28px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
