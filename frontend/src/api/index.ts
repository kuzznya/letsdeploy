import {Configuration, ManagedServiceApi, ProjectApi, ServiceApi} from "@/api/generated";
import {KeycloakInstance} from "@dsb-norge/vue-keycloak-js/dist/types";

let keycloak: KeycloakInstance | null = null

const config = new Configuration({
  basePath: import.meta.env.VITE_API_PATH,
  accessToken: () => {
    if (keycloak == null)
      throw new Error("Keycloak not provided to API client")
    if (!keycloak.authenticated || keycloak.token == null)
      throw new Error("User is not authenticated")
    return keycloak.token
  }
})

export default {
  ProjectApi: new ProjectApi(config),
  ServiceApi: new ServiceApi(config),
  ManagedServiceApi: new ManagedServiceApi(config),
  registerKeycloak(instance: KeycloakInstance) {
    keycloak = instance
  }
}
