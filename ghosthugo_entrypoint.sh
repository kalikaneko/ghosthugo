#!/bin/sh
set -e
cd /site
echo "[+] npm install"
npm install || echo "npm install failed"
echo "[+] importing ghost content..."
/usr/local/bin/ghosthugo
ls -latr content/posts
echo "[+] running hugo"
hugo "$@"
