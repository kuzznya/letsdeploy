import {inject, InjectionKey} from "vue";
import {KeycloakInstance} from "@dsb-norge/vue-keycloak-js/dist/types";

export const keycloakKey: InjectionKey<KeycloakInstance> = Symbol('keycloak')

export function useKeycloak(): KeycloakInstance {
  const instance = inject(keycloakKey)
  if (instance == undefined)
    throw new Error('Keycloak is not registered')
  return instance
}
