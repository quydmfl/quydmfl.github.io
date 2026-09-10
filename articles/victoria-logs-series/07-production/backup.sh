#!/usr/bin/env bash
# Backup VictoriaLogs bằng snapshot API: tạo snapshot (instant, không khoá
# ghi), copy snapshot ra ngoài container, rồi xoá snapshot tạm.
#
# Yêu cầu: docker, curl, jq.
# Lưu ý: field trong JSON response của API snapshot có thể khác nhau tuỳ
# version VictoriaLogs — kiểm tra lại với `curl -s .../snapshot/create | jq`
# trước khi dùng script này trong production.

set -euo pipefail

VLOGS_URL="${VLOGS_URL:-http://localhost:9428}"
CONTAINER_NAME="${CONTAINER_NAME:-victorialogs}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"

echo "==> Tạo snapshot mới trên VictoriaLogs (${VLOGS_URL})..."
RESPONSE="$(curl -sf -X POST "${VLOGS_URL}/internal/partition/snapshot/create")"
echo "Response: ${RESPONSE}"

SNAPSHOT_PATH="$(echo "${RESPONSE}" | jq -r '.snapshot // empty')"
if [[ -z "${SNAPSHOT_PATH}" ]]; then
  echo "Không lấy được đường dẫn snapshot từ response — kiểm tra lại API/version." >&2
  exit 1
fi

DEST="${BACKUP_DIR}/${TIMESTAMP}"
mkdir -p "${DEST}"

echo "==> Copy snapshot '${SNAPSHOT_PATH}' ra khỏi container ${CONTAINER_NAME}..."
docker cp "${CONTAINER_NAME}:${SNAPSHOT_PATH}" "${DEST}"

echo "==> Xoá snapshot tạm trên VictoriaLogs (đã copy xong, không cần giữ)..."
curl -sf -X POST "${VLOGS_URL}/internal/partition/snapshot/delete?path=${SNAPSHOT_PATH}" >/dev/null

echo "==> Backup xong tại: ${DEST}"
echo "Gợi ý: đẩy thư mục này lên object storage để an toàn khỏi mất cả host, ví dụ:"
echo "  rclone sync ${DEST} remote:my-bucket/victorialogs-backups/${TIMESTAMP}"
