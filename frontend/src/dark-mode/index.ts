import { App, computed, inject, InjectionKey, ref } from "vue";

function loadDarkMode(): boolean {
  const savedMode = localStorage.getItem("dark-mode");
  const mode =
    savedMode != null
      ? savedMode === "true"
      : window.matchMedia("(prefers-color-scheme: dark)").matches;
  document.documentElement.className = mode ? "dark" : "light";
  return mode;
}

class DarkMode {
  private enabled = ref(loadDarkMode());

  isEnabled() {
    return this.enabled;
  }

  switch() {
    this.enabled.value = !this.enabled.value;
    localStorage.setItem("dark-mode", this.enabled.value.toString() ?? "false");
    document.documentElement.className = this.enabled.value ? "dark" : "light";
  }

  asComputed = () => computed(() => this.isEnabled().value);
}

const manager = new DarkMode();

const darkModeKey: InjectionKey<DarkMode> = Symbol("dark-mode");

export function useDarkMode(): DarkMode {
  return inject(darkModeKey) ?? manager;
}

export default function install(app: App) {
  app.provide(darkModeKey, manager);
}
