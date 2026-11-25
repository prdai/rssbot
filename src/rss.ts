import type { FeedItem, ParsedFeed } from "./models";
import { XMLParser } from "fast-xml-parser";

const parser = new XMLParser({
  ignoreAttributes: false,
  attributeNamePrefix: "@_",
  parseTagValue: false,
  htmlEntities: true,
  trimValues: true,
});

const asArray = <T>(value: T | T[] | undefined): T[] => {
  if (value === undefined) return [];
  return Array.isArray(value) ? value : [value];
};

const asText = (value: unknown): string | null => {
  if (value === undefined || value === null) return null;
  if (typeof value === "string") return value.trim() || null;
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  if (Array.isArray(value)) {
    for (const entry of value) {
      const text = asText(entry);
      if (text !== null) return text;
    }
    return null;
  }
  if (typeof value === "object" && "#text" in value) {
    return asText((value as { "#text": unknown })["#text"]);
  }
  return null;
};

const atomLink = (value: unknown): string | null => {
  const links = asArray(value);
  const alternate = links.find((link) => {
    return (
      typeof link === "object" &&
      link !== null &&
      (link as Record<string, unknown>)["@_rel"] === "alternate"
    );
  });
  for (const link of [alternate, ...links]) {
    if (typeof link === "string") return link.trim() || null;
    if (typeof link === "object" && link !== null) {
      const href = (link as Record<string, unknown>)["@_href"];
      if (typeof href === "string" && href.trim()) return href.trim();
    }
  }
  return null;
};

const makeItem = (
  feedUrl: string,
  feedTitle: string,
  fields: {
    id: unknown;
    title: unknown;
    link: unknown;
    content: unknown;
    author: unknown;
    publishedAt: unknown;
  },
): FeedItem | null => {
  const title = asText(fields.title) ?? "Untitled entry";
  const link = asText(fields.link);
  const publishedAt = asText(fields.publishedAt);
  const id = asText(fields.id) ?? link ?? `${title}:${publishedAt ?? ""}`;
  if (!id) return null;
  return {
    id,
    title,
    link,
    contentHtml: asText(fields.content),
    author: asText(fields.author),
    publishedAt,
    feedTitle,
    feedUrl,
  };
};

const parseRss = (feedUrl: string, document: Record<string, unknown>): ParsedFeed | null => {
  const rss = document.rss as Record<string, unknown> | undefined;
  const channel = rss?.channel as Record<string, unknown> | undefined;
  if (channel === undefined) return null;
  const feedTitle = asText(channel.title) ?? feedUrl;
  const items = asArray(channel.item as Record<string, unknown>[] | undefined)
    .map((item) =>
      makeItem(feedUrl, feedTitle, {
        id: item.guid ?? item.link,
        title: item.title,
        link: item.link,
        content: item["content:encoded"] ?? item.description,
        author: item["dc:creator"] ?? item.author,
        publishedAt: item.pubDate ?? item["dc:date"],
      }),
    )
    .filter((item): item is FeedItem => item !== null);
  return { title: feedTitle, items };
};

const parseAtom = (feedUrl: string, document: Record<string, unknown>): ParsedFeed | null => {
  const feed = document.feed as Record<string, unknown> | undefined;
  if (feed === undefined) return null;
  const feedTitle = asText(feed.title) ?? feedUrl;
  const items = asArray(feed.entry as Record<string, unknown>[] | undefined)
    .map((entry) => {
      const author = entry.author as Record<string, unknown> | undefined;
      return makeItem(feedUrl, feedTitle, {
        id: entry.id,
        title: entry.title,
        link: atomLink(entry.link),
        content: entry.content ?? entry.summary,
        author: author?.name ?? entry["dc:creator"],
        publishedAt: entry.published ?? entry.updated,
      });
    })
    .filter((item): item is FeedItem => item !== null);
  return { title: feedTitle, items };
};

const parseRdf = (feedUrl: string, document: Record<string, unknown>): ParsedFeed | null => {
  const rdf = document["rdf:RDF"] as Record<string, unknown> | undefined;
  if (rdf === undefined) return null;
  const channel = rdf.channel as Record<string, unknown> | undefined;
  const feedTitle = asText(channel?.title) ?? feedUrl;
  const items = asArray(rdf.item as Record<string, unknown>[] | undefined)
    .map((item) =>
      makeItem(feedUrl, feedTitle, {
        id: item["rdf:about"] ?? item.link,
        title: item.title,
        link: item.link,
        content: item["content:encoded"] ?? item.description,
        author: item["dc:creator"],
        publishedAt: item["dc:date"],
      }),
    )
    .filter((item): item is FeedItem => item !== null);
  return { title: feedTitle, items };
};

export async function fetchFeed(feedUrl: string): Promise<ParsedFeed> {
  const response = await fetch(feedUrl, {
    headers: { "User-Agent": "rssbot/1.0" },
  });
  if (!response.ok) {
    throw new Error(`feed fetch failed (${response.status}): ${feedUrl}`);
  }
  const document = parser.parse(await response.text()) as Record<string, unknown>;
  const feed = parseRss(feedUrl, document) ?? parseAtom(feedUrl, document) ?? parseRdf(feedUrl, document);
  if (feed === null) {
    throw new Error(`unrecognised feed format: ${feedUrl}`);
  }
  return feed;
}
