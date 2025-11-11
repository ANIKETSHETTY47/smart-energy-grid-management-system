#!/usr/bin/env bash
set -euo pipefail
{
  echo "===== POSTDEPLOY DEBUG ====="
  echo "PWD: $(pwd)"
  echo "Listing /var/app/current:"
  ls -la /var/app/current || true
  echo "Listing /var/app/staging:"
  ls -la /var/app/staging || true
  echo "ENV:"
  env | sort
} >> /var/log/web.stdout.log 2>&1
