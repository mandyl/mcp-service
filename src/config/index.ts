export const config = {
  port: parseInt(process.env.PORT ?? '8080', 10),
  host: process.env.HOST ?? '0.0.0.0',
  logLevel: process.env.LOG_LEVEL ?? 'info',
  es: {
    host: process.env.ES_HOST ?? 'localhost',
    port: parseInt(process.env.ES_PORT ?? '9200', 10),
    username: process.env.ES_USERNAME,
    password: process.env.ES_PASSWORD,
  },
  search: {
    defaultSize: 10,
    maxSize: 100,
  },
};
