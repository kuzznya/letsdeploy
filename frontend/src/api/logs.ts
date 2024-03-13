import { Configuration } from "@/api/generated";

export class ServiceLogsApi {
  private baseURL: string | undefined;

  constructor(config: Configuration) {
    this.baseURL = config.basePath;
  }

  connectToLogStream(service: number, token: string) {
    const url = this.baseURL?.substring("http".length);
    return new WebSocket(
      `ws${url}/api/v1/services/${service}/logs?token=${token}`,
    );
  }
}
