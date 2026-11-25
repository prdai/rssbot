import type { AtomFeed, JsonFeed, RdfFeed, RssFeed } from "feedsmith";
import type { FeedItem, ParsedFeed } from "./models";
import { parseFeed } from "feedsmith";

interface ItemFields {
  id?: string;
  title?: string;
  link?: string;
  content?: string;
  author?: string;
  publishedAt?: string;
}

const clean = (value?: string): string | null => value?.trim() || null;

const makeItem = (
  feedUrl: string,
  feedTitle: string | undefined,
  fields: ItemFields,
): FeedItem => {
  const title = clean(fields.title) ?? "Untitled entry";
  const link = clean(fields.link);
  const publishedAt = clean(fields.publishedAt);
  const id = clean(fields.id) ?? link ?? `${title}:${publishedAt ?? ""}`;
  return {
    id,
    title,
    link,
    contentHtml: clean(fields.content),
    author: clean(fields.author),
    publishedAt,
    feedTitle: clean(feedTitle) ?? feedUrl,
    feedUrl,
  };
};

const atomLink = (links?: AtomFeed.Link<string>[]): string | undefined => {
  const alternate = links?.find((link) => !link.rel || link.rel === "alternate");
  return (alternate ?? links?.[0])?.href;
};

const rssItems = (feed: RssFeed.Feed<string>, feedUrl: string): FeedItem[] =>
  (feed.items ?? []).map((item) =>
    makeItem(feedUrl, feed.title, {
      id: item.guid?.value,
      title: item.title,
      link: item.link,
      content: item.content?.encoded ?? item.description,
      author: item.authors?.[0]?.name ?? item.dc?.creators?.[0],
      publishedAt: item.pubDate,
    }),
  );

const atomItems = (feed: AtomFeed.Feed<string>, feedUrl: string): FeedItem[] =>
  (feed.entries ?? []).map((entry) =>
    makeItem(feedUrl, feed.title?.value, {
      id: entry.id,
      title: entry.title?.value,
      link: atomLink(entry.links),
      content: entry.content?.value ?? entry.summary?.value,
      author: entry.authors?.[0]?.name ?? entry.dc?.creators?.[0],
      publishedAt: entry.published ?? entry.updated,
    }),
  );

const rdfItems = (feed: RdfFeed.Feed<string>, feedUrl: string): FeedItem[] =>
  (feed.items ?? []).map((item) =>
    makeItem(feedUrl, feed.title, {
      id: item.rdf?.about ?? item.link,
      title: item.title,
      link: item.link,
      content: item.content?.encoded ?? item.description,
      author: item.dc?.creators?.[0],
      publishedAt: item.dc?.dates?.[0],
    }),
  );

const jsonItems = (feed: JsonFeed.Feed<string>, feedUrl: string): FeedItem[] =>
  (feed.items ?? []).map((item) =>
    makeItem(feedUrl, feed.title, {
      id: item.id,
      title: item.title,
      link: item.url ?? item.external_url,
      content: item.content_html ?? item.content_text ?? item.summary,
      author: item.authors?.[0]?.name,
      publishedAt: item.date_published,
    }),
  );

export async function fetchFeed(feedUrl: string): Promise<ParsedFeed> {
  const response = await fetch(feedUrl, {
    headers: { "User-Agent": "rssbot/1.0" },
  });
  if (!response.ok) {
    throw new Error(`feed fetch failed (${response.status}): ${feedUrl}`);
  }

  const parsed = parseFeed(await response.text());
  switch (parsed.format) {
    case "rss":
      return { title: clean(parsed.feed.title) ?? feedUrl, items: rssItems(parsed.feed, feedUrl) };
    case "atom":
      return { title: parsed.feed.title?.value ?? feedUrl, items: atomItems(parsed.feed, feedUrl) };
    case "rdf":
      return { title: clean(parsed.feed.title) ?? feedUrl, items: rdfItems(parsed.feed, feedUrl) };
    case "json":
      return { title: clean(parsed.feed.title) ?? feedUrl, items: jsonItems(parsed.feed, feedUrl) };
  }
}
