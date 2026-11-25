const KEY_PREFIX = "feed:";
const MAX_SEEN_IDS = 5000;

export async function getSeenIds(kv: KVNamespace, feedUrl: string): Promise<string[] | null> {
  const value = await kv.get(`${KEY_PREFIX}${feedUrl}`, "json");
  return Array.isArray(value) ? value.filter((id): id is string => typeof id === "string") : null;
}

export async function setSeenIds(kv: KVNamespace, feedUrl: string, ids: string[]): Promise<void> {
  await kv.put(`${KEY_PREFIX}${feedUrl}`, JSON.stringify(ids.slice(0, MAX_SEEN_IDS)));
}
