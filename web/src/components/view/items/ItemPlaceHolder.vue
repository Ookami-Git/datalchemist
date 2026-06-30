<template>
  <div class="skeleton-wrapper" aria-hidden="true">
    <div class="skeleton-title modern-skeleton"></div>
    <div class="skeleton-line modern-skeleton w-75"></div>
    <div class="skeleton-line modern-skeleton w-50"></div>
    <div class="skeleton-line modern-skeleton w-85"></div>
    <div class="skeleton-button modern-skeleton"></div>
  </div>
</template>

<style scoped>
.skeleton-wrapper {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.25rem 0;
}

.skeleton-title {
  height: 1.25rem;
  width: 40%;
  margin-bottom: 0.5rem;
}

.skeleton-line {
  height: 0.75rem;
  border-radius: 4px;
}

.skeleton-button {
  height: 2.25rem;
  width: 30%;
  margin-top: 0.5rem;
  border-radius: 8px;
}

.modern-skeleton {
  background: linear-gradient(
    90deg,
    var(--bs-tertiary-bg) 25%,
    var(--bs-border-color) 50%,
    var(--bs-tertiary-bg) 75%
  );
  background-size: 200% 100%;
  animation: shimmer 1.6s infinite linear;
  border-radius: 6px;
}

@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}
</style>