import { Hono } from "hono";
import { syncFeeds } from "./sync";

const app = new Hono<{ Bindings: Env }>();

app.get("/", (c) => c.text("rssbot"));
app.post("/", async (c) => {
  await syncFeeds(c.env);
  return c.text("ok");
});

export default {
  fetch: app.fetch,
  scheduled: async (
    controller: ScheduledController,
    env: Env,
    _ctx: ExecutionContext,
  ) => {
    console.warn(`RSS sync triggered by ${controller.cron}`);
    await syncFeeds(env);
  },
};
