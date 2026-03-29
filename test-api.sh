#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:3000"
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
SERVER_PID=""

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log() {
  echo -e "$1"
}

cleanup() {
  if [[ -n "${SERVER_PID}" ]]; then
    kill "${SERVER_PID}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

require_cmd() {
  local cmd="$1"
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    log "${RED}Missing required command: ${cmd}${NC}"
    exit 1
  fi
}

wait_for_health() {
  for _ in {1..30}; do
    if curl -sS "${BASE_URL}/api/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  return 1
}

start_server_if_needed() {
  if curl -sS "${BASE_URL}/api/health" >/dev/null 2>&1; then
    log "${GREEN}API server already running${NC}"
    return
  fi

  log "${YELLOW}API server not running; starting with go run...${NC}"
  (cd "${PROJECT_ROOT}" && go run ./cmd/api/main.go >/tmp/pos_wms_api.log 2>&1) &
  SERVER_PID="$!"

  if ! wait_for_health; then
    log "${RED}Server did not become healthy in time${NC}"
    log "${YELLOW}Server log:${NC}"
    tail -50 /tmp/pos_wms_api.log || true
    exit 1
  fi

  log "${GREEN}API server started${NC}"
}

request() {
  local method="$1"
  local path="$2"
  local expected_status="$3"
  local body="${4:-}"

  local response
  local status
  local payload

  if [[ -n "${body}" ]]; then
    response=$(curl -sS -X "${method}" "${BASE_URL}${path}" \
      -H "Content-Type: application/json" \
      -d "${body}" \
      -w "\n%{http_code}")
  else
    response=$(curl -sS -X "${method}" "${BASE_URL}${path}" -w "\n%{http_code}")
  fi

  status=$(echo "${response}" | tail -n1)
  payload=$(echo "${response}" | sed '$d')

  if [[ "${status}" != "${expected_status}" ]]; then
    log "${RED}${method} ${path} expected ${expected_status}, got ${status}${NC}"
    echo "${payload}"
    exit 1
  fi

  echo "${payload}"
}

extract_id() {
  local json="$1"
  local field="$2"
  if [[ "${field}" != ".data.id" ]]; then
    log "${RED}Unsupported field selector: ${field}${NC}"
    exit 1
  fi
  echo "${json}" | sed -n 's/.*"data":{"id":\([0-9][0-9]*\).*/\1/p'
}

main() {
  require_cmd curl
  require_cmd go

  log "${BLUE}POS & WMS MVP - API Test Suite${NC}"
  log "${BLUE}=================================${NC}"

  start_server_if_needed

  local run_id
  run_id=$(date +%s)

  log "${BLUE}1. Health Check${NC}"
  request GET /api/health 200

  log "${BLUE}2. Create Branches${NC}"
  local branch_1_json
  local branch_2_json
  local branch_1_id
  local branch_2_id

  branch_1_json=$(request POST /api/branches 201 "{
    \"name\": \"Bangkok Store ${run_id}\",
    \"address\": \"123 Silom Rd, Bangkok\",
    \"phone\": \"02-123-4567\"
  }")
  branch_2_json=$(request POST /api/branches 201 "{
    \"name\": \"Chiang Mai Store ${run_id}\",
    \"address\": \"456 Nimman Rd, Chiang Mai\",
    \"phone\": \"053-234-5678\"
  }")

  branch_1_id=$(extract_id "${branch_1_json}" '.data.id')
  branch_2_id=$(extract_id "${branch_2_json}" '.data.id')

  echo "Branch 1 ID: ${branch_1_id}"
  echo "Branch 2 ID: ${branch_2_id}"
  request GET /api/branches 200 >/dev/null

  log "${BLUE}3. Create Products${NC}"
  local product_1_json
  local product_2_json
  local product_3_json
  local product_1_id
  local product_2_id
  local product_3_id

  product_1_json=$(request POST /api/products 201 "{
    \"sku\": \"SKU-001-${run_id}\",
    \"name\": \"iPhone 15 Pro\",
    \"description\": \"Latest Apple smartphone\",
    \"price\": 35999.00,
    \"cost\": 28000.00
  }")
  product_2_json=$(request POST /api/products 201 "{
    \"sku\": \"SKU-002-${run_id}\",
    \"name\": \"Samsung Galaxy S24\",
    \"description\": \"Premium Android Phone\",
    \"price\": 32999.00,
    \"cost\": 25000.00
  }")
  product_3_json=$(request POST /api/products 201 "{
    \"sku\": \"SKU-003-${run_id}\",
    \"name\": \"USB-C Cable\",
    \"description\": \"2m white cable\",
    \"price\": 299.00,
    \"cost\": 80.00
  }")

  product_1_id=$(extract_id "${product_1_json}" '.data.id')
  product_2_id=$(extract_id "${product_2_json}" '.data.id')
  product_3_id=$(extract_id "${product_3_json}" '.data.id')

  request GET "/api/products/${product_1_id}" 200 >/dev/null
  request GET "/api/products?limit=10&offset=0" 200 >/dev/null

  log "${BLUE}4. Seed Inventory via API${NC}"
  request POST /api/inventory 201 "{
    \"product_id\": ${product_1_id},
    \"branch_id\": ${branch_1_id},
    \"quantity\": 15,
    \"minimum_qty\": 5
  }" >/dev/null

  request POST /api/inventory 201 "{
    \"product_id\": ${product_1_id},
    \"branch_id\": ${branch_2_id},
    \"quantity\": 8,
    \"minimum_qty\": 5
  }" >/dev/null

  request POST /api/inventory 201 "{
    \"product_id\": ${product_2_id},
    \"branch_id\": ${branch_1_id},
    \"quantity\": 10,
    \"minimum_qty\": 5
  }" >/dev/null

  request POST /api/inventory 201 "{
    \"product_id\": ${product_3_id},
    \"branch_id\": ${branch_1_id},
    \"quantity\": 50,
    \"minimum_qty\": 20
  }" >/dev/null

  request POST /api/inventory 201 "{
    \"product_id\": ${product_3_id},
    \"branch_id\": ${branch_2_id},
    \"quantity\": 30,
    \"minimum_qty\": 20
  }" >/dev/null

  request GET "/api/inventory/product/${product_1_id}/branch/${branch_1_id}" 200 >/dev/null
  request GET "/api/inventory/branch/${branch_1_id}?limit=50&offset=0" 200 >/dev/null

  log "${BLUE}5. Process Sale (ACID Transaction)${NC}"
  local sale_json
  local order_id

  sale_json=$(request POST /api/sales 201 "{
    \"branch_id\": ${branch_1_id},
    \"customer_name\": \"John Doe\",
    \"items\": [
      {
        \"product_id\": ${product_1_id},
        \"quantity\": 2,
        \"discount\": 0
      },
      {
        \"product_id\": ${product_3_id},
        \"quantity\": 5,
        \"discount\": 50
      }
    ]
  }")
  order_id=$(extract_id "${sale_json}" '.data.id')

  echo "Order created ID: ${order_id}"

  log "${BLUE}6. Verify Order + Inventory After Sale${NC}"
  request GET "/api/orders/${order_id}" 200 >/dev/null
  request GET "/api/orders/branch/${branch_1_id}?limit=50&offset=0" 200 >/dev/null
  request GET "/api/inventory/product/${product_1_id}/branch/${branch_1_id}" 200 >/dev/null
  request GET "/api/inventory/low-stock/${branch_1_id}" 200 >/dev/null

  log "${GREEN}All API tests passed successfully${NC}"
}

main "$@"
