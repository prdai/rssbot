import type { FeedItem } from "./models";

const escapeHtml = (value: string): string =>
  value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");

const stripHtml = (value: string): string => value.replaceAll(/<[^>]*>/g, "");

const metaLine = (item: FeedItem): string => {
  const parts = [item.feedTitle, item.author, item.publishedAt].filter(
    (part): part is string => part !== null && part.length > 0,
  );
  return parts.map(escapeHtml).join(" · ");
};

export async function sendItemEmail(
  binding: SendEmail,
  item: FeedItem,
  from: string,
  to: string,
): Promise<void> {
  const content = item.contentHtml ?? "";
  const html = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"></head>
<body style="margin:0;padding:24px;font-family:system-ui,-apple-system,sans-serif;line-height:1.6;color:#222;">
<article style="max-width:640px;margin:0 auto;">
<h1 style="font-size:24px;line-height:1.3;margin:0 0 8px;">${escapeHtml(item.title)}</h1>
<p style="margin:0 0 24px;font-size:13px;color:#666;">${metaLine(item)}</p>
${content}
${item.link === null ? "" : `<p style="margin-top:24px;"><a href="${escapeHtml(item.link)}">Read original</a></p>`}
</article>
</body>
</html>`;
  const text = [item.title, metaLine(item), stripHtml(content), item.link ?? ""]
    .filter((part) => part.length > 0)
    .join("\n\n");

  await binding.send({ from, to, subject: item.title, html, text });
}
