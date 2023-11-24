#!/usr/bin/env bash

set -eux
cd `dirname $0`

################################################################################
echo "# Analyze"
################################################################################

# read env
# 計測用自作env
. /tmp/prepared_env

# isucon serviceで使うenv
. ./env.sh

result_dir=$HOME/result
mkdir -p ${result_dir}

# journal log
sudo journalctl -xe -ocat -u isucon.go.service --since="${prepared_time}" > "${app_journal_log}"
sudo journalctl -xe -ocat -u openresty.service --since="${prepared_time}" > "${nginx_journal_log}"

# alp
# ALPM="/int/\d+,/uuid/[A-Za-z0-9_]+,/6digits/[a-z0-9]{6}"
#ALPM="/@.+,/posts/\d+,/image/\d+.(jpg|png|gif),/posts?max_created_at.*$"
#ALPM="/api/courses/[a-zA-Z0-9]+$,/api/courses/[a-zA-Z0-9]+/status,/api/courses/[a-zA-Z0-9]+/classes,/api/courses/[a-zA-Z0-9]+/classes/[a-zA-Z0-9]+/assignments,/api/courses/[a-zA-Z0-9]+/classes/[a-zA-Z0-9]+/assignments/scores,/api/courses/[a-zA-Z0-9]+/classes/[a-zA-Z0-9]+/assignments/export,/api/announcements/[a-zA-Z0-9]+$"
ALPM="/initialize,/api/admin/clarifications,/api/admin/clarifications/\d,/api/session,/api/audience/teams,/api/audience/dashboard,/api/registration/session,/api/registration/team,/api/registration/contestant,/api/registration,/api/registration,/api/contestant/benchmark_jobs,/api/contestant/benchmark_jobs/\d,/api/contestant/clarifications,/api/contestant/clarifications,/api/contestant/dashboard,/api/contestant/notifications,/api/contestant/push_subscriptions,/api/contestant/push_subscriptions,/api/signup,/api/login,/api/logout"

OUTFORMT=count,1xx,2xx,3xx,4xx,5xx,method,uri,min,max,sum,avg,p95,min_body,max_body,avg_body
touch ${result_dir}/alp.md
cp ${result_dir}/alp.md ${result_dir}/alp.md.prev
alp json --file=${nginx_access_log} \
  --nosave-pos \
  --sort sum \
  --reverse \
  --output ${OUTFORMT} \
  --format markdown \
  --matching-groups ${ALPM}  \
  > ${result_dir}/alp.md

# OUTFORMT=count,uri_method_status,min,max,sum,avg,p95,trace_id_sample
# touch ${result_dir}/alp_trace.txt
# cp ${result_dir}/alp_trace.txt ${result_dir}/alp_trace.txt.prev
# alp-trace json --file=${nginx_access_log} \
#   --nosave-pos \
#   --sort sum \
#   --reverse \
#   --output ${OUTFORMT} \
#   --format pretty \
#   --matching-groups ${ALPM}  \
#   --trace \
#   > ${result_dir}/alp_trace.txt


# mysqlowquery
# sudo mysqldumpslow -s t ${mysql_slow_log} > ${result_dir}/mysqld-slow.txt

# touch ${result_dir}/pt-query-digest.txt
# cp ${result_dir}/pt-query-digest.txt ${result_dir}/pt-query-digest.txt.prev
# sudo chmod 755 `dirname ${mysql_slow_log}`
# sudo chmod 644 ${mysql_slow_log}
# pt-query-digest  --progress percentage,5 --explain "h=${DB_HOST},u=${DB_USER},p=${DB_PASS},D=${DB_DATABASE}" ${mysql_slow_log} > ${result_dir}/pt-query-digest.txt
# pt-query-digest ${mysql_slow_log} > ${result_dir}/pt-query-digest.txt


# 500エラー周辺のログをまとめる
touch log/app/5xx_journal.log
cp log/app/5xx_journal.log log/app/5xx_journal.log.prev
cat log/app/journal.log | grep -B5 -A5 -P '"status":5\d\d' > log/app/5xx_journal.log || true

touch log/nginx/5xx_access.log
cp log/nginx/5xx_access.log log/nginx/5xx_access.log.prev
cat log/nginx/access.log | grep -B5 -A5 -P '"status":"5\d\d"' > log/nginx/5xx_access.log || true


# access logから集計

# Varnishのキャッシュヒット率の集計
touch ${result_dir}/varnish_cache_hit.txt
cat log/nginx/access.log | jq -r '.cache' | sort | uniq -c | sort -nr > ${result_dir}/cache_hit_varnish.txt

# バックエンド(Goのプロセスキャッシュ・Redisなど)のキャッシュヒット率の集計
touch ${result_dir}/varnish_cache_hit.txt
cat log/nginx/access.log | jq -r '.upstream_http_cache_status' | sort | uniq -c | sort -nr > ${result_dir}/cache_hit_backend.txt

# Accept-Encodingの集計
touch ${result_dir}/accept_encoding.txt
cat log/nginx/access.log | jq -r .accept_encoding | sort | uniq -c | sort -nr > ${result_dir}/accept_encoding.txt

# User-Agentの集計
touch ${result_dir}/ua.txt
cat log/nginx/access.log | jq -r .ua | sort | uniq -c | sort -nr > ${result_dir}/ua.txt

# status code の集計
touch ${result_dir}/status_code.txt
cat log/nginx/access.log | jq -r .status | sort | uniq -c | sort -nr > ${result_dir}/status_code.txt
