package marketplace

// Handoff describes the state needed before an order can move to its buyer.
type Handoff struct {
	OrderID      string `json:"order_id"`
	SellerID     string `json:"seller_id"`
	AssetID      string `json:"asset_id"`
	AssetReady   bool   `json:"asset_ready"`
	BuyerUpdated bool   `json:"buyer_updated"`
	SellerOnline bool   `json:"seller_online"`
}

// Ready reports whether publishing the handoff is a valid state transition.
func (h Handoff) Ready() bool {
	return h.OrderID != "" && h.SellerID != "" && h.AssetID != "" &&
		h.AssetReady && h.BuyerUpdated && h.SellerOnline
}
