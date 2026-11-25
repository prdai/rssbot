export interface FeedItem {
  id: string;
  title: string;
  link: string | null;
  contentHtml: string | null;
  author: string | null;
  publishedAt: string | null;
  feedTitle: string;
  feedUrl: string;
}

export interface ParsedFeed {
  title: string;
  items: FeedItem[];
}
