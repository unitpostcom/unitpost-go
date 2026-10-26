package unitpost

import (
	"context"
	"errors"
	"fmt"
)

// Pager walks every row of a cursor-paginated list, following has_more with the
// last row's id as the next `after` cursor. Filters in the original params ride
// along on every page; only the cursor moves. Use it like bufio.Scanner:
//
//	p := client.Contacts.ListAll(map[string]string{"limit": "100"})
//	for p.Next(ctx) {
//		row := p.Current() // map[string]any
//	}
//	if err := p.Err(); err != nil { ... }
type Pager struct {
	http    *httpClient
	path    string
	params  map[string]string
	after   string
	buf     []map[string]any
	current map[string]any
	done    bool
	err     error
}

// ErrForwardOnly is Pager.Err() when ListAll was given a `before` cursor:
// iterators page forward only (use List to page backward).
var ErrForwardOnly = errors.New("unitpost: ListAll pages forward only: `before` is not supported (use List to page backward)")

func newPager(h *httpClient, path string, params map[string]string) *Pager {
	cp := map[string]string{}
	for k, v := range params {
		cp[k] = v
	}
	p := &Pager{http: h, path: path, params: cp, after: cp["after"]}
	// Refused up front: a `before` used to ride along with the moving `after`
	// and 422 on the second page.
	if cp["before"] != "" {
		p.err = ErrForwardOnly
	}
	return p
}

// Next advances to the next row, fetching the next page when the current one is
// exhausted. It returns false when the list is done or a request failed (check Err).
func (p *Pager) Next(ctx context.Context) bool {
	for len(p.buf) == 0 {
		if p.done || p.err != nil {
			return false
		}
		p.fetch(ctx)
	}
	p.current, p.buf = p.buf[0], p.buf[1:]
	return true
}

// Current is the row Next just advanced to.
func (p *Pager) Current() map[string]any { return p.current }

// Err is the first error a page request returned, if any.
func (p *Pager) Err() error { return p.err }

func (p *Pager) fetch(ctx context.Context) {
	params := map[string]string{}
	for k, v := range p.params {
		params[k] = v
	}
	params["after"] = p.after
	raw, err := p.http.request(ctx, "GET", p.path, q(params), nil, "")
	if err != nil {
		p.err = err
		return
	}
	page, ok := raw.(map[string]any)
	if !ok {
		p.err = fmt.Errorf("unitpost: %s did not return a list envelope", p.path)
		return
	}
	rows, _ := page["data"].([]any)
	for _, r := range rows {
		if m, ok := r.(map[string]any); ok {
			p.buf = append(p.buf, m)
		}
	}
	hasMore, _ := page["has_more"].(bool)
	if !hasMore || len(rows) == 0 {
		p.done = true
		return
	}
	last, _ := rows[len(rows)-1].(map[string]any)
	id, _ := last["id"].(string)
	if id == "" {
		p.done = true
		return
	}
	p.after = id
}
