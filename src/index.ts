import process from "node:process";
import { Container, getContainer } from "@cloudflare/containers";
import { Hono } from "hono";
import log from "loglevel";
import { RSSFEEDS } from "./data";

export class WorkerContainer extends Container<Env> {
  defaultPort = 8080;
  leepAfter = "9m";
  envVars = {
    GOOGLE_API_KEY: process.env.GOOGLE_API_KEY ?? "",
    UNTRACKED_FEED_MAX_ITEMS: process.env.UNTRACKED_FEED_MAX_ITEMS ?? "",
    CLOUDFLARE_API_TOKEN: process.env.CLOUDFLARE_API_TOKEN ?? "",
    CLOUDFLARE_ACCOUNT_ID: process.env.CLOUDFLARE_ACCOUNT_ID ?? "",
    KV_NAMESPACE_ID: process.env.KV_NAMESPACE_ID ?? "",
    FROM_EMAIL: process.env.FROM_EMAIL ?? "",
    TO_EMAIL: process.env.TO_EMAIL ?? "",
  };

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

app.post("/", async (c) => {
  const rssfeeds = JSON.stringify(RSSFEEDS);
  const container = getContainer(c.env.CONTAINER);
  return await container.fetch(c.req.raw, { body: rssfeeds });
});

export default {
  fetch: app.fetch,
  scheduled: async (
    controller: ScheduledController,
    _: Env,
    __: ExecutionContext,
  ) => {
    log.info(
      `Triggered RSS Feed Sync from ${controller.cron} at ${controller.scheduledTime}`,
    );
    const options = { method: "POST" };
    await fetch("/", options);
  },
};
