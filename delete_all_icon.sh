#!/usr/bin/env bash


set -eu

mkdir -p /home/isucon/webapp/public/img/
find /home/isucon/webapp/public/img/ -type f -name "*.jpeg" -exec rm -fv {} \;
