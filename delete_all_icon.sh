#!/usr/bin/env bash


set -eu

find /home/isucon/webapp/public/img/ -type f -name "*.jpeg" -exec rm -fv {} \;
