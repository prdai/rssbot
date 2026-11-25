import type { FeedItem } from "./models";
import { FEEDS } from "./data";
import { sendItemEmail } from "./email";
import { fetchFeed } from "./rss";
import { getSeenIds, setSeenIds } from "./store";

async function syncFeed(env: Env, feedUrl: string): Promise<void> {
  const feed = await fetchFeed(feedUrl);
  const seen = await getSeenIds(env.FEED_STATE, feedUrl);
  if (seen === null) {
    await setSeenIds(env.FEED_STATE, feedUrl, feed.items.map((item) => item.id));
    return;
  }

  const seenSet = new Set(seen);
  const newItems: FeedItem[] = feed.items.filter((item) => !seenSet.has(item.id));
  if (newItems.length === 0) return;

  const sent: string[] = [];
  for (const item of [...newItems].reverse()) {
    try {
      await sendItemEmail(env.EMAIL, item, env.FROM_EMAIL, env.TO_EMAIL);
      sent.push(item.id);
    } catch (error) {
      console.error(`failed to forward ${item.link ?? item.title}`, error);
    }
  }

  if (sent.length > 0) {
    await setSeenIds(env.FEED_STATE, feedUrl, [...sent.reverse(), ...seen]);
  }
}

export async function syncFeeds(env: Env): Promise<void> {
  const results = await Promise.allSettled(FEEDS.map((feed) => syncFeed(env, feed)));
  for (const result of results) {
    if (result.status === "rejected") {
      console.error(result.reason);
    }
  }
}
