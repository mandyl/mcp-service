import { FastifyInstance } from 'fastify';
import { checkEsConnection } from '../es/client';

export async function healthRoutes(fastify: FastifyInstance): Promise<void> {
  fastify.get('/health', async (_request, reply) => {
    const esConnected = await checkEsConnection();
    return reply.status(200).send({
      status: 'ok',
      es_connected: esConnected,
      timestamp: new Date().toISOString(),
    });
  });
}
