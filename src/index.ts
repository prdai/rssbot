import { syncFeeds } from "./sync";

export default {
  scheduled: async (
    controller: ScheduledController,
    env: Env,
    _ctx: ExecutionContext,
  ) => {
    console.warn(`RSS sync triggered by ${controller.cron}`);
    await syncFeeds(env);
  },
};
