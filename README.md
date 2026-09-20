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

`client.Sms` is the SMS channel: **beta, behind the launch gate**. It is wired in its GA shape (Send, Get, List, Brands, Numbers, contact SMS consent); while your workspace's gate is off every call returns `404`.

Docs: https://www.unitpost.com/docs
