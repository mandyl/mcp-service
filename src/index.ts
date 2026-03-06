import Fastify from 'fastify';
import { config } from './config';
import { healthRoutes } from './routes/health';
import { searchRoutes } from './routes/search';
import { indicesRoutes } from './routes/indices';

async function bootstrap(): Promise<void> {
  const fastify = Fastify({
    logger: {
      level: config.logLevel,
      transport:
        process.env.NODE_ENV !== 'production'
          ? { target: 'pino-pretty', options: { colorize: true } }
          : undefined,
    },
  });

  // Register all route plugins
  await fastify.register(healthRoutes);
  await fastify.register(searchRoutes);
  await fastify.register(indicesRoutes);

  // Graceful shutdown
  const signals: NodeJS.Signals[] = ['SIGTERM', 'SIGINT'];
  for (const signal of signals) {
    process.on(signal, async () => {
      fastify.log.info(`Received ${signal}, shutting down gracefully...`);
      await fastify.close();
      process.exit(0);
    });
  }

  try {
    await fastify.listen({ port: config.port, host: config.host });
    fastify.log.info(
      `mcp-service listening on ${config.host}:${config.port}`,
    );
    fastify.log.info(
      `Elasticsearch: http://${config.es.host}:${config.es.port}`,
    );
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
}

bootstrap();
