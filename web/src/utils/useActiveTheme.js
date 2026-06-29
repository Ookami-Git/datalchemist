import { ref, watch, onBeforeUnmount, getCurrentInstance } from 'vue';

/**
 * useActiveTheme – retourne le thème effectif à appliquer.
 *
 * Règle : si la préférence utilisateur vaut "default" (ou est absente),
 * on se rabat sur le thème du système (prefers-color-scheme).
 * Sinon, on utilise directement la préférence utilisateur ("light" | "dark").
 *
 * @param {import('vue').Ref} parametersRef – la ref `parameters` injectée
 * @returns {{ getActiveTheme: () => 'light' | 'dark', activeTheme: import('vue').Ref<'light' | 'dark'> }}
 */
export function useActiveTheme(parametersRef) {
  const getSystemTheme = () =>
    window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';

  const getActiveTheme = () => {
    const t = parametersRef?.value?.theme;
    return (!t || t === 'default') ? getSystemTheme() : t;
  };

  const activeTheme = ref(getActiveTheme());

  // Watcher sur les paramètres utilisateur
  watch(
    () => parametersRef?.value?.theme,
    () => {
      activeTheme.value = getActiveTheme();
    }
  );

  // Listener pour le thème système
  const systemThemeMQ = window.matchMedia('(prefers-color-scheme: dark)');
  const onSystemThemeChange = () => {
    activeTheme.value = getActiveTheme();
  };

  systemThemeMQ.addEventListener('change', onSystemThemeChange);

  // Nettoyage si on est dans le contexte d'un composant Vue
  if (getCurrentInstance()) {
    onBeforeUnmount(() => {
      systemThemeMQ.removeEventListener('change', onSystemThemeChange);
    });
  }

  return { getActiveTheme, activeTheme };
}

