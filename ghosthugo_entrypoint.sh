#!/bin/sh
set -e
cd /site
echo "[+] importing ghost content..."
/usr/local/bin/ghosthugo
ls -latr content/posts
echo "[+] running hugo"
hugo "$@"
