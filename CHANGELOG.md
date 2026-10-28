# Changelog

Notable changes to the Unitpost Go SDK. Changes to the API itself (new
operations, new fields, behavior changes) are in the API changelog:
https://www.unitpost.com/docs/changelog.

To release, move the `Unreleased` items under a `## X.Y.Z` heading in the same
change that bumps the version (docs/sdk/RELEASING.md).

## Unreleased

- The SMS list takes `sender_id`, a number's `snum_…` id: that number's
  texts, both directions. A value it can't apply is ignored and named in the
  `Unitpost-Warning` response header; `Unitpost-Strict: true` refuses it.
- `Sms.Quotes` (new): `Create(ctx, body)`, `Get(ctx, id)` and `Confirm(ctx,
  id, body, opts...)` buy SMS carrier resources (`squote_` ids): register a
  10DLC brand, vet it, register a campaign, order a number. `Confirm` mints an
  Idempotency-Key unless you pass `WithIdempotencyKey`, so a retry can't buy
  twice. `Create` prices the order and binds what it buys (the brand's and
  program's content, the price) and returns `due_now_cents`, `monthly_cents`,
  `line_items` and a single-use `confirm_token` (`sqct_`), only in that
  response; a quote lasts 10 minutes and charges nothing. The confirm places
  the order at that price, or answers `409` when the price or content changed.
  It needs `sms:purchase`, which no key preset includes, and the workspace
  owner's monthly cap for SMS purchases by API (no cap set refuses it). Show
  the person the price and get their yes before confirming.
- `Logs.List(ctx, params)` (new `Client.Logs`): the workspace's API request
  log, newest first and cursor-paginated (`rlog_` ids). Each row is one call:
  `method`, `path`, `status`, `duration_ms`, `api_key_id` and its
  `request_id` (the `x-request-id`), never the caller's IP or a body; rows are
  kept 30 days. Filters: `"status"` (a code or `"2xx"`…`"5xx"`),
  `"api_key_id"`, `"method"`, `"path_prefix"`, `"request_id"`,
  `"created_after"`, `"created_before"`. It needs `apikeys:read`, which no key
  preset includes.
- Opt many contacts out of a topic at once: `Email.Topics.OptOut(ctx, id,
  body, WithIdempotencyKey(key))` and `Sms.Topics.OptOut(...)` (exactly one
  of `contact_ids` or `emails`, up to 1,000, or `segment_id`). They answer
  the job (202); follow it with `Jobs.Get`. The SDK mints an idempotency key
  when you pass none, so a retry answers the same job. There is no bulk
  opt-in. Needs `topics:write` (a segment also `segments:read`).
- `Sms.Registrations.List`, `Sms.TestDestinations.List`,
  `Sms.ConsentSetups.List`, `Sms.Billing.ListResources` and
  `Sms.MessagingPrograms.List` take an optional params map with `"limit"`,
  `"after"` and `"before"` (`List(ctx)` still compiles): the five lists are
  cursor-paginated, newest first.
- New response fields are optional in the models (the API always sends
  them): pointers with `omitempty`, nil when absent. Keyed struct literals
  written against 0.4.0 still compile.
- The activity log, an API key's usage, several topic preferences at once and
  list search. `Workspace.ListActivity(ctx, params)` and
  `ListAllActivity(params)`: the workspace's activity log, newest first
  (cursor-paginated on `act_` ids): what changed, when, who (`actor.name`)
  and, for a change over the API, MCP or Ask Unit, how (`actor.via`, with the
  API key). Narrow it with `"subject"` (and `"subject_id"`), `"contact_id"`,
  `"campaign_id"`, `"q"` and `"channel": "email"` (an unknown key is a `422`).
  It needs `activity:read`. `ApiKeys.Usage(ctx, id)`: a key's requests in the
  last 24 hours (`requests_24h`, `client_errors_24h`, `server_errors_24h`,
  `rate_limited_24h`) and `last_used_at`. `ApiKeys.List` now also works with
  an API key or an MCP sign-in (it was dashboard-session only); with one,
  `List` and `Usage` need `apikeys:read`. Creating and revoking keys stay in
  the dashboard. No key preset includes `activity:read` or `apikeys:read`.
  `Email.Topics.UpdateTopics(ctx, contactID, {"topics": [{"topic_id": ...,
  "subscribed": ...}]})` (`PATCH /contacts/{id}/topics`, up to 50) changes
  several of a contact's preferences at once, leaves the others as they are,
  and answers the list (`"object": "list"`, `"data"`); `SetTopic` is
  unchanged. `Email.List` / `ListAll` take `"q"` (an
  exact recipient, or part of the subject or the sender), `"source"` (`api`,
  `smtp`, `test`, `batch`, `campaign` or `reply`) and `"outcome"`
  (`delivered`, `opened`, `clicked`, `unsubscribed`, `bounced`, `complained`,
  `failed` or `pending`; with `"campaign_id"`, who opened or clicked it).
  `Sms.List` / `ListAll` take `"q"` (the message text, or 3 or more digits of
  either number) and `"type"` (`transactional` or `marketing`). A `q`,
  `source`, `outcome` or `type` value the API can't apply is ignored and named
  in the `Unitpost-Warning` response header; `Unitpost-Strict: true` refuses
  it.
- List filters and a stats period. `Suppressions.List` / `ListAll` take `"q"`
  (part of the address), `"reason"`, `"created_after"` and `"created_before"`
  (`"scope"` stays, deprecated). `Webhooks.ListDeliveries` /
  `ListAllDeliveries` take `"status"` (comma-separated, e.g.
  `"failed,exhausted"`) and `"event_type"`. `Email.Stats` keeps `"days"` and
  takes a period instead: `"from"` / `"to"`, with `"grain"` (`5m`, `1h` or
  `1d`) and `"tz"` for `series`, the send volume per bucket, and `"domain_id"`
  for one sending domain. These lists ignored the filters before, so a value
  the API can't apply is ignored and named in the `Unitpost-Warning` response
  header; `Unitpost-Strict: true` refuses it. Responses: `Webhooks.Get` adds
  `backlog` (`queued`, `failed`: what a flush would send); `Email.Stats` adds
  the permanent and transient bounces, `suppressed`, `unsubscribed`,
  `rates.unsubscribe`, `domain_id`, `from`, `to` and `series`; topics add
  `subscriber_count` and `unsubscribed_count`; `Usage.Get` adds the plan's
  `limits` and `features`.
- Who is calling, an email's events, a received email's original and
  template rendering. `Workspace.Whoami(ctx)`: the workspace, the credential
  (an API key or an OAuth sign-in, never its secret), its `capabilities` and
  the workspace's rate limit; any valid credential may call it, so call it
  first to see what a key may do. `Email.ListEvents(ctx, id, params)`
  (`"limit"`, `"after"`, `"before"`): one email's delivery events, oldest
  first, with each event's detail under the webhook's names
  (cursor-paginated on `emev_` ids; no `ListAll`).
  `Email.ReceivedRaw(ctx, id)`: a signed URL (about 5 minutes) to a received
  message's original `.eml`, the way to read a body over 256 KB; a copy of
  your own send is a `403`. `Email.Templates.Render(ctx, id, body)`
  (`{"variables": ..., "subject": ..., "contact_id": ...}`, or `nil`): the
  subject, HTML and text a send would carry, and its `missing_variables`; it
  sends and stores nothing, so it takes no idempotency key. `ListEvents` and
  `ReceivedRaw` need `emails:read`; `Render` needs `templates:read`.
- SMS brands and registrations. On `Sms`:
  `CreateBrand(ctx, body, opts...)`, `GetBrand(ctx, id)` and
  `UpdateBrand(ctx, id, body)` (send `{"archived": ...}` alone to archive or
  restore). A new resource on `Sms`: `Sms.Registrations.List(ctx)`, which is
  cursor-paginated (no `ListAll`). Creating a brand is free and makes no
  carrier call; its tax id is write-only (read back as
  `tax_id_on_file`), and a locked brand refuses carrier-field edits with
  `409 brand_locked`. The reads need `sms:read`; the brand writes need
  `sms:manage`. Registering a brand with carriers (a paid step), ordering
  numbers and a brand's documents stay in the dashboard.
- SMS setup. On `Sms.Templates`: `Usage(ctx, id)` and
  `TestSend(ctx, id, body, opts...)`. Four new resources on `Sms`:
  `Sms.MessagingPrograms.List(ctx[, params])` (`{"archived": "true"}` includes
  archived ones), `Create(ctx, body, opts...)`, `Get(ctx, id)`,
  `Update(ctx, id, body)`, `Delete(ctx, id)` (it archives) and
  `Duplicate(ctx, id, opts...)`; `Sms.TestDestinations.List(ctx)` and
  `Delete(ctx, id)`; `Sms.ConsentSetups.List(ctx)`; and
  `Sms.Billing.ListResources(ctx)`. These four lists are cursor-paginated (no
  `ListAll`). A test text goes only to the simulator or a verified test
  number, and is billed like any text. The reads need `sms:read`
  (`Usage` needs `templates:read`); the test send needs `sms:send`; the
  program writes and removing a test number need `sms:manage` (FULL keys have
  it, SENDING keys don't). Ordering numbers, sender
  registrations and payments stay in the dashboard.
- Brand Kit writes and imports on `BrandKits`: `BrandKits.Create(ctx, body)`,
  `BrandKits.Update(ctx, id, body)`, `BrandKits.Delete(ctx, id)`,
  `BrandKits.Reset(ctx, id)`, `BrandKits.SetDefault(ctx, id)` and
  `BrandKits.AddGraphic(ctx, id, body)`; `BrandKits.Import(ctx, body)`,
  `BrandKits.GetImport(ctx, id)`,
  `BrandKits.AcceptImport(ctx, id, body)` (nil sends `{}`) and
  `BrandKits.CancelImport(ctx, id)`. An import builds a profile from a website
  (`url`) or the brand's own words (`text`) in the background (`202`; follow
  it with `GetImport`). By default it is applied when it lands; with
  `"review": true` it stops at a proposal for `AcceptImport`. Deleting the
  last profile is a `409` (reset it instead). `AddGraphic` brings an image
  from a public https URL into the image library and pins it. The writes, and
  starting, accepting or canceling an import, need the `branding:write`
  capability, which no key preset grants: tick it on the key. The reads stay
  `templates:read`.
- Three new resources. Image library: `LibraryImages.List(ctx, params)`,
  `LibraryImages.ListAll(params)`,
  `LibraryImages.Create(ctx, body, opts...)`, `LibraryImages.Get(ctx, id)`,
  `LibraryImages.Update(ctx, id, body)` and
  `LibraryImages.Delete(ctx, id)` / `ForceDelete(ctx, id)`. Branding: `Branding.Get(ctx)`,
  `Branding.Update(ctx, body)`, `Branding.SetLogo(ctx, body)` and
  `Branding.RemoveLogo(ctx)`.
  Workspace: `Workspace.Update(ctx, body)`, `Workspace.SetImage(ctx, body)`
  and `Workspace.RemoveImage(ctx)`. The library holds the images a template
  embeds: an image's `path` (`/img/…`) goes in a template's `<Image src>`.
  `Create` uploads from a public https `url` or base64 `data` (PNG, JPEG,
  GIF or WebP, up to 5 MB) and takes `WithIdempotencyKey`, minted when you
  don't pass one, so a retried upload stores one image. Deleting an image
  that a template, a campaign or sent mail uses is a `409` `resource_in_use`
  unless you pass `force` set to `true` (sent as `?force=true`). `Branding`
  is the look of marketing emails' footer and the hosted unsubscribe page,
  as Settings → Brand shows it (`null` restores a default). `Workspace`
  renames the workspace and sets its image. `RemoveLogo` and `RemoveImage`
  answer the updated object. The branding writes and every `Workspace` call
  need the `branding:write` capability, which no key preset grants: tick it
  on the key.
- Jobs: `Jobs.List(ctx, params)`, `Jobs.Cancel(ctx, id)`,
  `Jobs.Retry(ctx, id)` and `Jobs.Download(ctx, id)`. Contacts:
  `Contacts.Export(ctx, body)` (the filters of `Contacts.List`; nil exports
  every contact). Segments: `Segments.Preview(ctx, id)`,
  `Segments.PreviewFilter(ctx, body)`, `Segments.AllContacts(ctx)` and
  `Segments.UpdateMember(ctx, id, contact, body)`.
  `Export` starts a background job that writes a CSV of every contact, or
  of the ones the filters match (one export at a time: another while it runs
  is a `409`); follow it with `Jobs.Get`, then `Jobs.Download` gives a signed URL to
  the file that expires in 5 minutes. `Jobs.List` lists your background jobs;
  `Cancel` stops a queued or running one, and `Retry` resumes a `FAILED`
  one from where it stopped. `Preview` and `PreviewFilter`
  count who a saved segment or an unsaved filter reaches and how many of them
  a campaign would mail; nothing is saved. `AllContacts` returns the
  managed "All subscribed contacts" segment, made the first time you ask.
  `UpdateMember` opts a member out of a static segment's sends, or back
  in, without removing it. None of these takes an idempotency key.
- Webhooks: `Webhooks.ReplayDelivery(ctx, id, deliveryID)` and
  `Webhooks.Flush(ctx, id, body)`. Email:
  `Email.Retry(ctx, id, WithIdempotencyKey(key))` and
  `Email.ListBatches(ctx, params)`. Domains: sender addresses with
  `Email.Domains.ListSenders(ctx, params)`, `CreateSender(ctx, body)`,
  `GetSender(ctx, id)`, `UpdateSender(ctx, id, body)` and
  `DeleteSender(ctx, id)`.
  `ReplayDelivery` delivers one event again with the same event id, and `Flush`
  delivers the deliveries held while a webhook was suspended (with
  `include_failed`, also the ones that failed for good); both need the
  webhook enabled (`409` otherwise). `Retry` sends a failed email again on
  the same id (an email that can't be retried is a `409` saying why); it
  sends real email, so it takes an idempotency key and generates one when you
  don't pass it. `ListBatches` lists your batch sends with the same
  rollup as `GetBatch`. A
  sender is a saved From address on one of your domains; saving a mailbox
  that is already saved there is a `409`.
- Email templates:
  `Email.Templates.Duplicate(ctx, id, body, WithIdempotencyKey(key))`,
  `Email.Templates.FromLibrary(ctx, body, WithIdempotencyKey(key))`,
  `Email.Templates.Usage(ctx, id)`,
  `Email.Templates.ListVersions(ctx, id, params)` and
  `Email.Templates.RestoreVersion(ctx, id, version)`.
  `Duplicate` copies a template into a new draft, and `FromLibrary` adds a
  gallery email as a draft Design template (the ones marked for paid plans are
  `403 limit_reached` on the free plan); both take an idempotency key and
  generate one when you don't pass it. `Usage` says what uses a template and
  whether an edit would be accepted now. `ListVersions` pages by version
  number: pass the last row's `version` as `after`. `RestoreVersion` puts a
  version back and is refused with `409` while a live campaign renders the
  template.
- Inbox conversations: a new `Email.Threads` resource with
  `List(ctx, params)`, `Get(ctx, id)`, `Delete(ctx, id)` and
  `Reply(ctx, id, body, WithIdempotencyKey(key))`. A conversation is a
  received message plus the replies and forwards sent from it, as Emails →
  Received shows them. `List` pages with `next_cursor` (pass it as `after`; it
  is not a row id); `Delete` erases the received messages for good. `Reply` replies, replies all or forwards one
  received message and sends real email, so it takes an idempotency key and
  generates one when you don't pass it.
- Email campaigns: `Email.Campaigns.Stats(ctx, id)`,
  `Email.Campaigns.ListStats(ctx, ids)`,
  `Email.Campaigns.ListRecipients(ctx, id, params)`,
  `Email.Campaigns.ListAffected(ctx, id)`,
  `Email.Campaigns.RemoveAffected(ctx, id)` and
  `Email.Campaigns.RequestReview(ctx, id)`. SMS campaigns:
  `Sms.Campaigns.ListRecipients(ctx, id, params)`. `Stats` reads a campaign's
  results: delivered, unique opens and clicks, unsubscribes, bounces,
  complaints and their rates. `ListStats` reads them for up to 100 campaigns
  at once. On an email campaign, `ListRecipients` lists who a send now would
  reach and why any other member would be left out; pass the page's
  `next_cursor` as `after` for the next page. After an automatic quality
  pause, `ListAffected` lists the hard-bounced and complained addresses,
  `RemoveAffected` takes them out of the campaign's static segment, and
  `RequestReview` asks for the review a repeat pause needs before `Resume`.
  Both are safe to repeat: a second call removes no one and notifies no one.
  On an SMS campaign, `ListRecipients` lists how each recipient fared, with
  `skipped_by_reason` counts for the whole campaign.
- Contacts: `Contacts.Addresses(ctx, idOrEmail)`,
  `Contacts.Activity(ctx, idOrEmail, params)`,
  `Contacts.ListSegments(ctx, idOrEmail)` and
  `Contacts.RetryImport(ctx, id)`. SMS:
  `Sms.RequestSmsOptIn(ctx, contactID, body, opts...)`. `RequestSmsOptIn`
  texts the contact a request to reply YES to confirm
  marketing texts. It sends a real text, so it sends the idempotency key you
  pass, or one it makes up, and a retry returns the same text instead of
  sending another.
- Contacts carry `phone`, their SMS number in E.164, on create, update and
  read. `null` on update removes the number only with the `Unitpost-Strict: true` header; without it the number is kept and the response warns. Setting it records no SMS
  consent: record that with `Sms.RecordSmsConsent`.
- `Contacts.List` / `Contacts.ListAll` take `q`, `status` (`subscribed` |
  `unsubscribed`), `phone` (`has` | `none`), `segment_id` and
  `segment_subscription` (`subscribed` | `opted_out`) in `params`;
  `Email.List` / `Email.ListAll` take `to` and `contact_id`.
- Email and SMS campaigns: `Unschedule(ctx, id)` and `Duplicate(ctx, id, body, opts...)`. Unschedule moves a scheduled campaign
  back to a draft; duplicate copies a campaign into a new draft. Duplicate
  sends the idempotency key you pass, or one it makes up, so a retry returns
  the same copy instead of making a second one.
- The campaign update model takes `name`, `subject`, `from`, `preview_text`,
  `template_id` and `segment_id`: the API now edits these on a draft.
- The email campaign and segment lists take the optional `archived` query
  parameter (`archived=true` lists archived rows).
- The API now answers `422` for a misnamed send time (`send_at`,
  `scheduledAt` on SMS, …) on sends and campaign writes, instead of sending at
  once or dropping it. Use `scheduled_at`.
- Campaign create's `403` now says which refusal it is: `limit_reached` for
  the plan's contact limit, `domain_not_found` / `domain_blocked` /
  `domain_not_verified` for the sender (it was always `domain_not_verified`).
  Read `error.code` to tell them apart.
- Domains: the update model takes `capabilities` (`{ receiving: "enabled" |
  "disabled" }` turns receiving on or off), and domain reads carry
  `capabilities` and a `receiving` object (status, the MX record to publish,
  the last checks).
- A `429` whose `Retry-After` is an HTTP date now waits until that date
  (capped at 20 s); only the seconds form was read before, so a date fell back
  to the jittered backoff. `Retry-After: 0` now retries at once.

## 0.4.0

The release this changelog starts from. Earlier changes are in the git history
(tags `sdk-go-v*`).
