import log from "loglevel";

export const ScheduledCron = async (
        controller: ScheduledController,
        _: Env,
        __: ExecutionContext,
) => {
        log.info(
                `Triggered RSS Feed Sync from ${controller.cron} at ${controller.scheduledTime}`,
        );
        const options = { method: "POST" };
        await fetch("/", options);
};
