<template>
  <section class="dc-loader" aria-live="polite" aria-busy="true">
    <div class="dc-loader__core" aria-hidden="true">
      <span class="dc-loader__halo"></span>
      <span class="dc-loader__arc dc-loader__arc--primary"></span>
      <span class="dc-loader__arc dc-loader__arc--secondary"></span>
      <span class="dc-loader__pulse"></span>
    </div>
    <span class="visually-hidden">Loading</span>
  </section>
</template>

<style lang="scss" scoped>
.dc-loader {
  --loader-primary: #2f7cf4;
  --loader-secondary: #17b897;
  --loader-ink: #1f2b3d;
  --loader-glow: rgba(47, 124, 244, 0.3);
  --loader-size: 54px;

  width: 100%;
  min-width: 0;
  min-height: 5.5rem;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

:global([data-bs-theme='dark']) .dc-loader {
  --loader-primary: #7ab7ff;
  --loader-secondary: #71e0bf;
  --loader-ink: #eef4ff;
  --loader-glow: rgba(122, 183, 255, 0.48);
}

.dc-loader__core {
  position: relative;
  width: var(--loader-size);
  height: var(--loader-size);
  display: inline-grid;
  place-items: center;
}

.dc-loader__halo,
.dc-loader__arc,
.dc-loader__pulse {
  position: absolute;
  border-radius: 999px;
}

.dc-loader__halo {
  inset: -8px;
  background: radial-gradient(circle, var(--loader-glow) 0%, rgba(0, 0, 0, 0) 72%);
  filter: blur(1px);
  animation: loader-breathe 2.4s ease-in-out infinite;
}

.dc-loader__arc {
  inset: 0;
  border: 3px solid transparent;
}

.dc-loader__arc--primary {
  border-top-color: var(--loader-primary);
  border-right-color: var(--loader-primary);
  animation: loader-spin-cw 1.1s linear infinite;
}

.dc-loader__arc--secondary {
  inset: 10px;
  border-bottom-color: var(--loader-secondary);
  border-left-color: var(--loader-ink);
  animation: loader-spin-ccw 0.95s cubic-bezier(0.6, 0.1, 0.4, 0.9) infinite;
}

.dc-loader__pulse {
  width: 9px;
  height: 9px;
  background: var(--loader-primary);
  box-shadow:
    0 0 0 0 rgba(0, 0, 0, 0),
    0 0 14px var(--loader-glow);
  animation: loader-pulse 1.5s ease-in-out infinite;
}

@keyframes loader-spin-cw {
  0% {
    transform: rotate(0deg);
  }

  100% {
    transform: rotate(360deg);
  }
}

@keyframes loader-spin-ccw {
  0% {
    transform: rotate(360deg);
  }

  100% {
    transform: rotate(0deg);
  }
}

@keyframes loader-pulse {

  0%,
  100% {
    transform: scale(0.86);
    opacity: 0.74;
  }

  50% {
    transform: scale(1.12);
    opacity: 1;
  }
}

@keyframes loader-breathe {

  0%,
  100% {
    transform: scale(0.9);
    opacity: 0.45;
  }

  50% {
    transform: scale(1.08);
    opacity: 0.9;
  }
}
</style>