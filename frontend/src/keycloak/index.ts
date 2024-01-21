import { inject, InjectionKey } from "vue";
import Keycloak from "keycloak-js";

export const keycloakKey: InjectionKey<Keycloak> = Symbol("keycloak");

export function useKeycloak(): Keycloak {
  const instance = inject(keycloakKey);
  if (instance == undefined) throw new Error("Keycloak is not registered");
  return instance;
}
