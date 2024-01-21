import {
  Configuration,
  ManagedServiceApi,
  ProjectApi,
  ServiceApi,
} from "@/api/generated";
import Keycloak from "keycloak-js";

let keycloak: Keycloak | null = null;

const config = new Configuration({
  basePath: import.meta.env.VITE_API_PATH,
  accessToken: () => {
    if (keycloak == null)
      throw new Error("Keycloak not provided to API client");
    if (!keycloak.authenticated || keycloak.token == null)
      throw new Error("User is not authenticated");
    return keycloak.token;
  },
});

export default {
  ProjectApi: new ProjectApi(config),
  ServiceApi: new ServiceApi(config),
  ManagedServiceApi: new ManagedServiceApi(config),
  registerKeycloak(instance: Keycloak) {
    keycloak = instance;
  },
};
