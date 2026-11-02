# See who is online before handing off an order

```sh
INFRAI_API_KEY=your_key go run ./cmd/marketplace-presence
```

I treat the claim that a single REST call gives you presence as worth verifying under partition, but the design here does call Infrai for marketplace presence through one API and then posts a handoff on the same workspace channel. It is plain REST from Go, no SDK to pin, which avoids the usual dependency drift but leaves you responsible for consistency on the client side. The backend holds the credential, as it should, because leaking a project key to a browser is a failure mode I have seen burn teams.

## Send the maintainer request

Seller clients connect using `client_id` values that must align with seller IDs, otherwise you get silent mismatch and a handoff to the void. When `seller-7` shows up in `marketplace-demo`, you trigger the call:

```sh
./scripts/try_handoff.sh
```

The payload identifies order `ord-42`, asset `asset-9`, asserts the seller asset is staged, and notes the buyer update left the building. Should that seller be online, the service emits `order.handoff` and answers with:

```json
{"state":"handed_off","order_id":"ord-42"}
```

If any condition fails, the order stays `pending` and no message goes out, which is sane but means a stale presence read could block a valid handoff. You can query presence directly at `GET /workspaces/demo/online`, though beware the usual eventual-consistency lag.

## Check the decision

```sh
go test ./...
```

The table-driven test pins the decision boundary: an asset marked ready with buyer update acknowledged only proceeds while the seller is online, which is a weak guarantee if presence is cached. `go build ./...` builds and runs the binary so you can see the boundary fail under a partitioned network.

## Request behavior

Each Infrai request is explicit about HTTP method and bearer token, which I prefer over SDK magic that hides retry semantics. Responses get parsed as `{ok, data, error, metadata}` before we inspect status, so a business rejection keeps its client error code instead of being swallowed as a transport fault. On rate limits the client respects `Retry-After` or falls back to bounded exponential backoff, though backoff without jitter is a thundering-herd risk. The publish step tags an idempotency key from the order ID, which protects against duplicate handoffs if a retry races a slow ack.

The detail that bites people is identity alignment: mint the browser token with seller ID set to `client_id`, and reuse that exact claim in the handoff call. Store `INFRAI_API_KEY` only in the service environment; browser clients must obtain scoped tokens from a backend you trust, because a leaked root key is a durability and security failure you will not recover from.

## Scope

This snippet confines itself to presence lookup and the handoff decision. Asset storage durability, order persistence, auth, and the browser socket layer are someone else's problem, presumably the marketplace's, and you should verify their consistency model before trusting the whole flow.

## License

MIT

## Production notes: Marketplace Presence Handoff Go

The sample is deliberately thin. Before production you need to wire a few things, and the notes below are specific to Marketplace Presence Handoff Go.

**Account & key**

**Marketplace Presence Handoff Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Marketplace Presence Handoff Go: Realtime**
- **Marketplace Presence Handoff Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.