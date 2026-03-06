import { FastifyInstance, FastifyRequest, FastifyReply } from 'fastify';
import { getEsClient } from '../es/client';
import { config } from '../config';

interface SearchRequestBody {
  index: string;
  query: string;
  from?: number;
  size?: number;
  filters?: {
    field: string;
    value: string;
  };
}

export async function searchRoutes(fastify: FastifyInstance): Promise<void> {
  fastify.post(
    '/api/v1/search',
    async (request: FastifyRequest<{ Body: SearchRequestBody }>, reply: FastifyReply) => {
      const body = request.body as SearchRequestBody;

      // --- Parameter validation ---
      if (!body || typeof body.index !== 'string' || !body.index.trim()) {
        return reply.status(200).send({
          code: 40002,
          message: 'missing required field: index',
          data: null,
        });
      }

      if (!body.query || typeof body.query !== 'string' || !body.query.trim()) {
        return reply.status(200).send({
          code: 40002,
          message: 'missing required field: query',
          data: null,
        });
      }

      const from = body.from ?? 0;
      const size = body.size ?? config.search.defaultSize;

      if (size > config.search.maxSize) {
        return reply.status(200).send({
          code: 40002,
          message: `size exceeds maximum allowed value of ${config.search.maxSize}`,
          data: null,
        });
      }

      // --- Build ES query ---
      const esQuery: Record<string, unknown> = {
        bool: {
          must: [
            {
              multi_match: {
                query: body.query,
                type: 'best_fields',
                fields: ['*'],
              },
            },
          ],
        },
      };

      if (body.filters?.field && body.filters?.value !== undefined) {
        (esQuery.bool as Record<string, unknown>).filter = [
          { term: { [body.filters.field]: body.filters.value } },
        ];
      }

      // --- Execute search ---
      try {
        const client = getEsClient();

        const response = await client.search({
          index: body.index,
          from,
          size,
          query: esQuery,
        });

        const hits = response.hits.hits.map((hit) => ({
          _id: hit._id,
          _score: hit._score,
          _source: hit._source,
        }));

        const total =
          typeof response.hits.total === 'number'
            ? response.hits.total
            : (response.hits.total?.value ?? 0);

        fastify.log.info({
          msg: 'search success',
          index: body.index,
          query: body.query,
          total,
          returned: hits.length,
          userId: request.headers['x-user-id'],
        });

        return reply.status(200).send({
          code: 0,
          message: 'success',
          data: { total, hits, from, size },
        });
      } catch (err: unknown) {
        // Distinguish "index_not_found" from other errors
        const errAny = err as Record<string, unknown>;
        const meta = errAny?.meta as Record<string, unknown> | undefined;
        const body = meta?.body as Record<string, unknown> | undefined;
        const errType = (body?.error as Record<string, unknown> | undefined)?.type;

        if (errType === 'index_not_found_exception') {
          fastify.log.warn({ msg: 'index not found', index: body?.index });
          return reply.status(200).send({
            code: 40001,
            message: `index not found: ${body?.index ?? 'unknown'}`,
            data: null,
          });
        }

        // Connection / general ES error
        const isConnectionError =
          errType === undefined &&
          (String(errAny?.message ?? '').includes('connect') ||
            String(errAny?.message ?? '').includes('ECONNREFUSED'));

        if (isConnectionError) {
          fastify.log.error({ msg: 'ES connection failed', err });
          return reply.status(200).send({
            code: 50001,
            message: 'elasticsearch connection failed',
            data: null,
          });
        }

        fastify.log.error({ msg: 'search error', err });
        return reply.status(200).send({
          code: 50002,
          message: 'internal server error',
          data: null,
        });
      }
    },
  );
}
