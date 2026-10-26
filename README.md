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

## Idempotency

`Email.Send`, `Email.Batch` and `Sms.Send` always carry an `Idempotency-Key`: the SDK mints one per call so its own automatic retries can never send twice. Pass your own key to dedupe across separate calls (for example a job retried after your process crashed): the API replays the first result for the same key for 24 hours instead of sending again.

```go
client.Email.Send(ctx, body, unitpost.WithIdempotencyKey("order-1234-receipt"))
client.Email.Batch(ctx, emails, unitpost.WithIdempotencyKey("digest-2026-10-24"))
client.Sms.Send(ctx, sms, unitpost.WithIdempotencyKey("otp-5678"))
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
