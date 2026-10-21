package marketplace

import "testing"

func TestHandoffReady(t *testing.T) {
	tests := []struct {
		name string
		in   Handoff
		want bool
	}{
		{
			name: "seller can hand off a ready asset after buyer update",
			in: Handoff{OrderID: "ord-42", SellerID: "seller-7", AssetID: "asset-9",
				AssetReady: true, BuyerUpdated: true, SellerOnline: true},
			want: true,
		},
		{
			name: "offline seller keeps order pending",
			in: Handoff{OrderID: "ord-42", SellerID: "seller-7", AssetID: "asset-9",
				AssetReady: true, BuyerUpdated: true, SellerOnline: false},
			want: false,
		},
		{
			name: "buyer has not received the update",
			in: Handoff{OrderID: "ord-42", SellerID: "seller-7", AssetID: "asset-9",
				AssetReady: true, BuyerUpdated: false, SellerOnline: true},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Ready(); got != tt.want {
				t.Fatalf("Ready() = %v, want %v", got, tt.want)
			}
		})
	}
}
