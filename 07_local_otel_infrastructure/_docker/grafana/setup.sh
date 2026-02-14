#!/bin/sh
# Grafana の初期設定スクリプト

# Grafana が起動するまで待機
echo "Waiting for Grafana to start..."
sleep 15

# 組織名を設定
echo "Updating organization name to 'OTEL Sample'..."
curl -X PUT "http://grafana:3000/api/org" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"name":"OTEL Sample"}' \
  && echo "✓ Organization name updated successfully!" \
  || echo "✗ Failed to update organization name (may already be set)"
