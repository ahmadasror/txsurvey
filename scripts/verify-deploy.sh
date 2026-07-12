#!/usr/bin/env bash
# Post-deploy smoke gate for txsurvey. Three layers, each red line naming the
# likely culprit: container up -> local app /health -> public edge.
#
# Run ONLY on the prod host, AFTER `docker compose -f docker-compose.prod.yml
# -p txsurvey up -d --build`. See docs/runbooks/deploy.md. Exits non-zero if any
# layer fails.
set -uo pipefail

APP_LOCAL="${APP_LOCAL:-http://172.17.0.1:8092}"
EDGES=("${EDGE_1:-https://brainzap.net/txsurvey/}" "${EDGE_2:-https://ct.tuxceria.biz.id/txsurvey/}")
CONTAINER="${CONTAINER:-txsurvey-app-1}"
rc=0

code() { curl -s -o /dev/null -w "%{http_code}" -L --max-time 10 "$1" 2>/dev/null || echo "000"; }

echo "== 1. container =="
if docker ps --filter "name=${CONTAINER}" --format '{{.Names}}' 2>/dev/null | grep -q "${CONTAINER}"; then
  echo "  ok   ${CONTAINER} running"
else
  echo "  FAIL ${CONTAINER} not running → check 'docker compose -f docker-compose.prod.yml -p txsurvey logs app'; the image may have failed to boot (bad migration / missing env in deploy/production.env)."
  rc=1
fi

echo "== 2. local app (${APP_LOCAL}) =="
h=$(code "${APP_LOCAL}/health")
if [ "$h" = "200" ]; then
  echo "  ok   /health 200 (app up, DB reachable)"
else
  echo "  FAIL /health ${h} → app not answering on the docker bridge. Container up but unhealthy? Check port publish (172.17.0.1:8092->8080) and DB connectivity."
  rc=1
fi

echo "== 3. public edge =="
for e in "${EDGES[@]}"; do
  c=$(code "$e")
  if [ "$c" = "200" ]; then
    echo "  ok   ${e} 200"
  else
    echo "  FAIL ${e} ${c} → app is local-healthy but the edge isn't serving. Usual culprits: nginx holding a stale upstream inode (reload nginx), Cloudflare tunnel down, or a missing iptables/docker-bridge rule after a host reboot."
    rc=1
  fi
done

echo
[ "$rc" = 0 ] && echo "verify-deploy: PASS" || echo "verify-deploy: FAIL"
exit "$rc"
