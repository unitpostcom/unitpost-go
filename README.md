# unitpost-go

Official Unitpost SDK for Go.

```bash
go get github.com/unitpostcom/unitpost-go
```

```go
client := unitpost.New() // reads UNITPOST_API_KEY
data, err := client.Email.Send(ctx, map[string]any{
    "from": "Acme <team@acme.com>",
    "to":   "user@example.com",
    "html": "<p>Hi</p>",
})
```

`client.Sms` is the SMS channel (Send, Get, List, Brands, Numbers, contact SMS consent). Available on paid plans: on the free plan `Send` returns a `403` with code `limit_reached`, while the reads keep working.

## Organization keys

A platform that gives each of its customers a workspace uses an organization key (`ok_live_…`) with the second client, `OrganizationClient`. It takes the key and the organization's id (`org_…`), and reads `UNITPOST_ORG_KEY` and `UNITPOST_ORGANIZATION_ID` when you leave them out (like `New`, it panics with a `*unitpost.Error`, code `missing_api_key` or `missing_organization_id`, when one is missing):

```go
org := unitpost.NewOrganizationWithOptions(unitpost.OrganizationOptions{
    APIKey:         os.Getenv("UNITPOST_ORG_KEY"),
    OrganizationID: os.Getenv("UNITPOST_ORGANIZATION_ID"),
})

ws, err := org.Workspaces.Create(ctx, map[string]any{"name": "Acme Bakery", "external_id": "cust_42"})
minted, err := org.WorkspaceKeys.Create(ctx, ws.(map[string]any)["id"].(string), map[string]any{"name": "Production", "scope": "sending"})

// Work inside that workspace with its own key.
acme := unitpost.NewWithOptions(unitpost.Options{APIKey: minted.(map[string]any)["key"].(string)})
```

| Field | Covers |
| --- | --- |
| `Workspaces` | `List` (`external_id` filter), `ListAll`, `Create` (takes `WithIdempotencyKey`), `Get`, `Update`, `Close`, `Restore` |
| `WorkspaceKeys` | `Create(ctx, workspaceID, body)`: mint a workspace key; the secret is in that answer only |
| `Webhooks` | The organization's endpoints (events from every workspace, each with `workspace_id` and `external_id`): `List`, `ListAll`, `Create`, `Get`, `Update`, `Delete`, `RotateSecret`, `Test`, `ListDeliveries` (needs `organizations:write`), `ListAllDeliveries`, `ReplayDelivery`, `Flush`, `Verify` |
| `Usage` | `List` (`"period"`), `ListAll`: usage per workspace for its billing period |
| `PlanQuotes` | Change a workspace's plan: `Create` prices it, `Get`, `Confirm` applies and charges it (takes `WithIdempotencyKey`) |
| `SharedPlan` | The organization's shared plan: `Get`, `ListWorkspaces`, `AddWorkspace`, `UpdateWorkspace` (its cap), `RemoveWorkspace` (adding and removing take `WithIdempotencyKey`) |

Each client takes one kind of key: the organization client only calls `/organizations/{id}/…` routes, and the API answers `401` to an organization key on any other route. An Owner or Admin makes the key in the dashboard under Organization → API.

## Idempotency

`Email.Send`, `Email.Batch` and `Sms.Send` always carry an `Idempotency-Key`: the SDK mints one per call so its own automatic retries can never send twice. Pass your own key to dedupe across separate calls (for example a job retried after your process crashed): the API replays the first result for the same key for 24 hours instead of sending again.

```go
client.Email.Send(ctx, body, unitpost.WithIdempotencyKey("order-1234-receipt"))
client.Email.Batch(ctx, emails, unitpost.WithIdempotencyKey("digest-2026-10-24"))
client.Sms.Send(ctx, sms, unitpost.WithIdempotencyKey("otp-5678"))
```

## Retries

Transient failures are retried automatically: `429`, `5xx` and transport errors (timeouts, network failures) are retried up to 2 times by default (3 attempts in all). When the error response carries a `Retry-After` header (a `429` always does), the client waits that long, read as seconds or an HTTP date and capped at 20 seconds. Otherwise it waits with exponential backoff and full jitter: a random time up to 500 ms × 2^attempt, also capped at 20 seconds. A cancelled `ctx` ends the wait. Client errors (`401`/`403`/`404`/`409`/`422`) are never retried.

Only requests that are safe to replay are retried: `GET`s and writes that carry an idempotency key (`Email.Send`, `Email.Batch` and `Sms.Send` always do, see above). Any other write without a key is sent once.

The retry count is the only setting; the delays are fixed. `MaxRetries` applies when `MaxRetriesSet` is true:

```go
// No retries at all.
client := unitpost.NewWithOptions(unitpost.Options{MaxRetries: 0, MaxRetriesSet: true})
```

## Pagination

List endpoints are cursor-paginated. `list` returns one page; the iterator below walks every page, following `has_more` with the last row's `id` as the next `after` cursor. Filters you pass ride along on every page.

```go
p := client.Contacts.ListAll(map[string]string{"limit": "100"})
for p.Next(ctx) {
    contact := p.Current() // map[string]any
    fmt.Println(contact["email"])
}
if err := p.Err(); err != nil { /* *unitpost.Error */ }
```

## Receive a text and answer it

US toll-free, 10DLC and Canadian numbers take replies (`two_way` is true on `client.Sms.Numbers`). Subscribe a webhook to `sms.received`, skip STOP and texts Unitpost already answered, read the conversation with `participant`, and answer without `from` so the reply leaves from the number they texted.

```go
// Your webhook endpoint, subscribed to sms.received.
func onWebhook(client *unitpost.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        raw, _ := io.ReadAll(r.Body) // the RAW body
        event, err := client.Webhooks.Verify(string(raw), os.Getenv("UNITPOST_WEBHOOK_SECRET"), r.Header, 0) // whsec_…; 0 = 300s tolerance
        if err != nil {
            http.Error(w, "bad signature", http.StatusBadRequest)
            return
        }
        e := event.(map[string]any)
        reply, _ := e["data"].(map[string]any)
        if e["type"] != "sms.received" || reply["keyword"] == "OPT_OUT" || reply["auto_replied"] == true {
            return // not a reply, STOP, or already answered
        }
        from, _ := reply["from"].(string)
        body, _ := reply["body"].(string) // empty when the reply carried no text

        // The whole conversation with this person, both directions, newest first.
        // Other filters: direction, status, from, to, contact_id, campaign_id,
        // created_after, created_before.
        history, _ := client.Sms.List(r.Context(), map[string]string{"participant": from, "limit": "20"})
        answer := yourAgent(history, body)

        // No "from": the answer goes out from the number they texted.
        client.Sms.Send(r.Context(), map[string]any{"to": from, "body": answer})
    }
}
```

Docs: https://www.unitpost.com/docs
