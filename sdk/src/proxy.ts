import fastify from "fastify";
export class Proxy {
  #rpcMapping: Record<string, string>;
  #hostname: string;
  #baseUrl: string;

  constructor(baseUrl: string, rpcMapping: Record<string, string>) {
    this.#baseUrl = baseUrl;
    this.#hostname = new URL(baseUrl).hostname;
    this.#rpcMapping = rpcMapping;
  }

  async start(port = 3000) {
    const server = fastify({ logger: true });

    // disable content type parsing for all paths
    server.removeContentTypeParser("*");
    server.addContentTypeParser(
      "*",
      { parseAs: "buffer" },
      (req, body, done) => {
        done(null, body);
      }
    );

    // Handle POST requests for any path
    server.post("/*", async (request, reply) => {
      try {
        if (request.hostname !== this.#hostname) {
          throw new Error(
            `Request from not supported hostname: ${
              request.hostname
            }, supported: ${this.#hostname}`
          );
        }

        const originalPath = request.url;
        const mappedPath = this.#rpcMapping[originalPath];

        if (!mappedPath) {
          throw new Error(`No mapping found for path: ${originalPath}`);
        }

        const response = await fetch(mappedPath, {
          method: request.method,
          body: request.body as BodyInit,
          headers: request.headers as HeadersInit,
        });

        return response;
      } catch (error: any) {
        console.error(error.toString());
        reply.code(500).send({ error: error.message });
      }
    });

    try {
      await server.listen({ port, host: "0.0.0.0" });
    } catch (err: any) {
      console.error("Error starting server:", err);
      process.exit(1);
    }
  }

  updateMapping(rpcMapping: Record<string, string>) {
    this.#rpcMapping = rpcMapping;
  }

  getMapping(): Record<string, string> {
    return this.#rpcMapping;
  }
}
