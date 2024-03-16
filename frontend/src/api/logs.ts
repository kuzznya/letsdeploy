import { Configuration } from "@/api/generated";

export class ServiceLogsApi {
  private baseURL: string | undefined;

  constructor(config: Configuration) {
    this.baseURL = config.basePath;
  }

  connectToLogStream(service: number, token: string, replica?: number) {
    const baseUrl = this.baseURL?.substring("http".length);
    const replicaParam = replica ? `&replica=${replica}` : "";
    const url =
      `ws${baseUrl}/api/v1/services/${service}/logs?token=${token}` +
      replicaParam;
    return new WebSocket(url);
  }
}
