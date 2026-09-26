# rssbot

Daily RSS digest bot. A Cloudflare Worker triggers a Go container on a cron
schedule; the container fetches the feeds, diffs them against stored state, asks
an LLM to write an HTML digest, and sends it by email.

## How it works

1. Cron (`0 0 * * *`) invokes the `scheduled` handler in `src/index.ts`.
2. The Worker forwards the feed list from `src/data.ts` to the `WorkerContainer`
   Durable Object, which proxies requests to the Go HTTP server on `:8080`.
3. The container (Go, `container_src/`):
   - fetches each feed concurrently (`gofeed`),
   - reads the last seen item hash per feed from Cloudflare KV and stores the new
     one (key is a hash of the feed URL, value is the last item hash),
   - renders the new items into the digest prompt (`container_src/prompts/index.j2`),
   - calls an OpenCode Zen chat model for a `{ title, body }` JSON digest.
4. The container sends the digest through Cloudflare Email Service.

Cloudflare REST APIs are called from Go with the official `cloudflare-go/v6` SDK.
`send_email` and KV bindings are Worker-only, which is why the logic lives in the
container but talks to the REST API.

## Configuration

Non-secret config lives in `wrangler.jsonc` under `vars` (it is deployed from the
repo, so a redeploy cannot wipe it). Credentials are Worker secrets.

Vars (`wrangler.jsonc`):

| Var | Description |
| --- | --- |
| `OPENCODE_MODEL` | OpenCode Zen model id (default `space-bunny-free`) |
| `UNTRACKED_FEED_MAX_ITEMS` | Items to backfill for a feed seen for the first time |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare account id |
| `KV_NAMESPACE_ID` | KV namespace that stores feed state |
| `FROM_EMAIL` | Sender address for the digest |
| `TO_EMAIL` | Recipient of the digest |

Secrets:

| Secret | Description |
| --- | --- |
| `OPENCODE_API_KEY` | OpenCode Zen API key (`oc_sk_...`) |
| `CLOUDFLARE_API_TOKEN` | API token with Workers KV Storage Edit and Email Sending Send |

Set secrets with:

```sh
CLOUDFLARE_ACCOUNT_ID=<account_id> bunx wrangler secret put OPENCODE_API_KEY
CLOUDFLARE_ACCOUNT_ID=<account_id> bunx wrangler secret put CLOUDFLARE_API_TOKEN
```

The account has more than one Cloudflare account, so `CLOUDFLARE_ACCOUNT_ID` must
be set (or `account_id` in `wrangler.jsonc`) for non-interactive Wrangler
commands.

### OpenCode Zen

Models are served from `https://opencode.ai/zen/v1` (OpenAI-compatible chat
completions). `space-bunny-free` is free and reachable with an API key. Other
free models are restricted to use inside the OpenCode app, and paid models
require Zen credits; pick one and set `OPENCODE_MODEL` accordingly.

## One-time setup

1. Create the KV namespace and put its id in `wrangler.jsonc` (`KV_NAMESPACE_ID`):

   ```sh
   bunx wrangler kv namespace create rssbot
   ```

2. Create a Cloudflare API token with `Account > Workers KV Storage > Edit` and
   `Account > Email Sending > Send`, then set it as `CLOUDFLARE_API_TOKEN`.

3. Onboard the sending domain under **Compute > Email Service > Email Sending**.
   This adds the `cf-bounce` MX/SPF/DKIM records and a DMARC record. Sending is a
   separate onboarding from Email Routing; a domain with Email Routing enabled is
   not necessarily able to send.

4. Set `OPENCODE_API_KEY` and deploy.

## Local development

```sh
bun install          # if this fails extracting wrangler, use: npm install --ignore-scripts
bun run cf-typegen   # regenerate worker-configuration.d.ts
bun run dev
```

## Deploy

```sh
CLOUDFLARE_ACCOUNT_ID=<account_id> bunx wrangler deploy
```

## Checks

```sh
cd container_src && go build ./... && go vet ./... && gofmt -l .
mise exec node@26.7.0 -- ./node_modules/.bin/tsc --noEmit
mise exec node@26.7.0 -- ./node_modules/.bin/eslint src
```
