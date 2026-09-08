#!/bin/bash
cd /usr/share/nginx/html &&
rm -Rf manager/* &&
mv manager-* manager &&
cd manager &&
unzip manager-* &&
rm -f manager-*
echo 'manager deploy finished.'
