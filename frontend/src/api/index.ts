import {Configuration, ManagedServiceApi, ProjectApi, ServiceApi} from "@/api/generated";
import {useKeycloak} from "@/keycloak";

const config = new Configuration({
  basePath: import.meta.env.VITE_API_PATH,
  accessToken: () => {
    const keycloak = useKeycloak()
    if (!keycloak.authenticated || keycloak.token == null)
      throw new Error("User is not authenticated")
    return keycloak.token
  }
})

export default {
  ProjectApi: new ProjectApi(config),
  ServiceApi: new ServiceApi(config),
  ManagedServiceApi: new ManagedServiceApi(config)
}
