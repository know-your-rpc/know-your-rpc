import express, { Request, Response } from "express";
import fetch from "node-fetch";

export class Proxy {
  #rpcMapping: Record<string, string>;

  constructor(rpcMapping: Record<string, string>) {
    this.#rpcMapping = rpcMapping;
  }

  public async start(hostname: string, port = 3000): Promise<void> {
    const app = express();

    app.use(express.raw({ type: "*/*" }));

    app.post("/*", async (req: Request, res: Response) => {
      try {
        const fullUrl = `https://${hostname}${req.url}`;
        const mappedPath = this.#rpcMapping[fullUrl];

        if (!mappedPath) {
          throw new Error(`No mapping found for path: ${fullUrl}`);
        }

        const fetchResponse = await fetch(mappedPath, {
          method: req.method,
          headers: {
            "content-type": req.headers["content-type"] as string,
          },
          body: req.body,
        });

        res.status(fetchResponse.status);

        fetchResponse.headers.forEach((value, key) => {
          res.setHeader(key, value);
        });

        if (fetchResponse.body) {
          fetchResponse.body.pipe(res);
        } else {
          res.end();
        }
      } catch (err: unknown) {
        const error = err as Error;
        console.error(error);
        res.status(500).json({ error: error.message });
      }
    });

    app.listen(port, "0.0.0.0", () => {
      console.log(`Express server listening on port ${port}`);
    });
  }

  public updateMapping(rpcMapping: Record<string, string>): void {
    this.#rpcMapping = rpcMapping;
  }

  public getMapping(): Record<string, string> {
    return this.#rpcMapping;
  }
}
