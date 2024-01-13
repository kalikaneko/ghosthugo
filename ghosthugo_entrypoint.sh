#!/bin/sh
set -e
# Check if content dir is defined, fallback to a default if not
CONTENT=${HUGO_CONTENT:-content/posts}
cd /site
echo "[+] npm install"
npm install || echo "npm install failed"
echo "[+] importing ghost content..."
/usr/local/bin/ghosthugo
ls -latr ${CONTENT}
echo "[+] running hugo"
hugo "$@"
