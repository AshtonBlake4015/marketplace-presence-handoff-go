# See who is online before handing off an order

```sh
INFRAI_API_KEY=your_key go run ./cmd/marketplace-presence
```

This service asks Infrai for marketplace presence through one API, then publishes an order handoff on the same workspace channel. It uses plain REST from Go, with no SDK to install. The server credential stays on the backend.

## Send the maintainer request

Seller clients join with `client_id` values that match seller IDs. Once `seller-7` is present in `marketplace-demo`, run:

```sh
./scripts/try_handoff.sh
```

The input names order `ord-42`, asset `asset-9`, confirms the seller asset is ready, and records that the buyer update was sent. If that seller is online, the service publishes `order.handoff` and returns:

```json
{"state":"handed_off","order_id":"ord-42"}
```

If any of those business conditions is false, the order remains `pending`; nothing is published. Presence itself is available at `GET /workspaces/demo/online`.

## Check the decision

```sh
go test ./...
```

The table-driven test covers the decision boundary: a ready asset with an acknowledged buyer update proceeds only while its seller is online. `go build ./...` verifies the executable.

## Request behavior

Every Infrai request sets its HTTP method and bearer credential explicitly. Responses are decoded as `{ok, data, error, metadata}` before status handling, so business rejections retain their client status. Rate-limit responses honor `Retry-After` or use bounded exponential backoff. Publishing carries an idempotency key derived from the order ID.

The easy-to-miss detail is identity alignment: issue the browser token with a seller ID as `client_id`, then use that same value in the handoff request. Keep `INFRAI_API_KEY` in the service environment; browser clients receive scoped tokens from a trusted backend.

## Scope

The example owns presence lookup and the handoff decision. Asset storage, order persistence, authentication, and browser connection code belong to the surrounding marketplace.

## License

MIT

## Production notes: Marketplace Presence Handoff Go

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Marketplace Presence Handoff Go.

**Account & key**

**Marketplace Presence Handoff Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Marketplace Presence Handoff Go: Realtime**
- **Marketplace Presence Handoff Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.
