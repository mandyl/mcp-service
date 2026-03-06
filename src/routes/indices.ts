import { FastifyInstance } from 'fastify';
import { getEsClient } from '../es/client';

export async function indicesRoutes(fastify: FastifyInstance): Promise<void> {
  fastify.get('/api/v1/indices', async (_request, reply) => {
    try {
      const client = getEsClient();
      const response = await client.cat.indices({ format: 'json', h: 'index' });

      const indices = (response as Array<{ index?: string }>)
        .map((row) => row.index ?? '')
        .filter((name) => name && !name.startsWith('.'));  // skip system indices

      return reply.status(200).send({
        code: 0,
        message: 'success',
        data: { indices },
      });
    } catch (err) {
      fastify.log.error({ msg: 'failed to list indices', err });
      return reply.status(200).send({
        code: 50001,
        message: 'elasticsearch connection failed',
        data: null,
      });
    }
  });
}
