import { Container, getContainer } from "@cloudflare/containers";
import { Hono } from "hono";
import log from "loglevel";
import { loadFile } from "./utils";

export class WorkerContainer extends Container<Env> {
  defaultPort = 8080;
  leepAfter = "9m";
  envVars = {};
  override onStart() {
    log.info("Container successfully started");
  }

  override onStop() {
    log.info("Container successfully shut down");
  }

  override onError(error: unknown) {
    log.info("Container error:", error);
  }
}

const app = new Hono<{
  Bindings: Env;
}>();

app.get("/", async (c) => {
  const rssfeeds = loadFile("./data/rssfeeds.json");
  const container = getContainer(c.env.CONTAINER);
  return await container.fetch(c.req.raw, { body: rssfeeds });
});

export default app;
