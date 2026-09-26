import process from "node:process";
import { Container, getContainer } from "@cloudflare/containers";
import { Hono } from "hono";
import log from "loglevel";
import { RSSFEEDS } from "./data";

export class WorkerContainer extends Container<Env> {
  defaultPort = 8080;
  sleepAfter = "9m";
  envVars = {
    OPENCODE_API_KEY: process.env.OPENCODE_API_KEY ?? "",
    OPENCODE_MODEL: process.env.OPENCODE_MODEL ?? "",
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

const runSync = async (env: Env) => {
  const container = getContainer(env.CONTAINER);
  return await container.fetch("http://container/", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(RSSFEEDS),
  });
};

app.post("/", async (c) => runSync(c.env));

export default {
  fetch: app.fetch,
  scheduled: async (
    controller: ScheduledController,
    env: Env,
    _ctx: ExecutionContext,
  ) => {
    log.info(
      `Triggered RSS Feed Sync from ${controller.cron} at ${controller.scheduledTime}`,
    );
    await runSync(env);
  },
};
