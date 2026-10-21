#!/bin/sh
set -eu

curl --fail-with-body --request POST \
  --header 'Content-Type: application/json' \
  --data '{"order_id":"ord-42","seller_id":"seller-7","asset_id":"asset-9","asset_ready":true,"buyer_updated":true}' \
  http://localhost:8080/workspaces/demo/handoffs
