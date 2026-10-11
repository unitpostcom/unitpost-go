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

Docs: https://www.unitpost.com/docs
