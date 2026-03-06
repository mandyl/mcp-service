import { Client } from '@elastic/elasticsearch';
import { config } from '../config';

let esClient: Client | null = null;

export function getEsClient(): Client {
  if (esClient) return esClient;

  const node = `http://${config.es.host}:${config.es.port}`;

  const clientOptions: ConstructorParameters<typeof Client>[0] = { node };

  if (config.es.username && config.es.password) {
    clientOptions.auth = {
      username: config.es.username,
      password: config.es.password,
    };
  }

  esClient = new Client(clientOptions);
  return esClient;
}

export async function checkEsConnection(): Promise<boolean> {
  try {
    const client = getEsClient();
    await client.ping();
    return true;
  } catch {
    return false;
  }
}
