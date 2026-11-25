# rssbot

Forwards new RSS items to email. A Cloudflare Worker runs on a cron, fetches the
feeds listed in `src/data.ts`, and sends one email per new item from
`FROM_EMAIL` to `TO_EMAIL`.

The email subject is the item title. The body is the item content, with the feed
name, author, and published date above it. No digest and no LLM generation.

## How it works

1. Cron (`0 * * * *`) invokes the `scheduled` handler in `src/index.ts`.
2. For each feed, the Worker fetches and parses RSS, Atom, or RDF
   (`src/rss.ts`).
3. Seen item ids are stored per feed in the `FEED_STATE` KV namespace. On the
   first run for a feed the current ids are recorded without sending email.
4. Every unseen item is sent as its own email through the `EMAIL` binding. An
   item is marked seen only after its email is sent.

## Configuration

Non-secret config lives in `wrangler.jsonc` under `vars`.

| Var | Description |
| --- | --- |
| `FROM_EMAIL` | Sender address (must be on a domain onboarded to Email Sending) |
| `TO_EMAIL` | Recipient |

Bindings:

| Binding | Description |
| --- | --- |
| `FEED_STATE` | KV namespace that stores seen item ids per feed |
| `EMAIL` | Email Sending binding, restricted to `TO_EMAIL` |

## One-time setup

1. Create the KV namespace and put its id in `wrangler.jsonc`:

   ```sh
   bunx wrangler kv namespace create rssbot
   ```

2. Onboard the sending domain under **Compute > Email Service > Email Sending**.
   This adds the `cf-bounce` MX/SPF/DKIM records and a DMARC record.

3. Deploy.

## Local development

```sh
bun install
bun run cf-typegen   # regenerate worker-configuration.d.ts
bun run dev
```

`wrangler dev` simulates the email binding and logs messages locally.

## Deploy

```sh
bunx wrangler deploy
```

## Checks

```sh
./node_modules/.bin/tsc --noEmit
./node_modules/.bin/eslint src
```
