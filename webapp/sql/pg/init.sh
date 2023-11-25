#!/usr/bin/env bash

set -eux
cd $(dirname $0)

if test -f /home/isucon/env.sh; then
	. /home/isucon/env.sh
fi

ISUCON_DB_HOST=${ISUCON13_MYSQL_DIALCONFIG_ADDRESS:-127.0.0.1}
ISUCON_DB_PORT=${ISUCON13_MYSQL_DIALCONFIG_PORT:-3306}
ISUCON_DB_USER=${ISUCON13_MYSQL_DIALCONFIG_USER:-isucon}
ISUCON_DB_PASSWORD=${ISUCON13_MYSQL_DIALCONFIG_PASSWORD:-isucon}
ISUCON_DB_NAME=${ISUCON13_MYSQL_DIALCONFIG_DATABASE:-isupipe}

psql -U isucon -d isupipe -f "./init.sql";

psql -U isucon -d isupipe -c "\COPY isupipe.users FROM ./users.csv DELIMITER ',' CSV";
psql -U isucon -d isupipe -c "\COPY isupipe.livestreams FROM ./livestreams.csv DELIMITER ',' CSV";
psql -U isucon -d isupipe -c "\COPY isupipe.tags FROM ./tags.csv DELIMITER ',' CSV";
psql -U isucon -d isupipe -c "\COPY isupipe.livestream_tags FROM ./livestream_tags.csv DELIMITER ',' CSV";
psql -U isucon -d isupipe -c "\COPY isupipe.reservation_slots FROM ./reservation_slots.csv DELIMITER ',' CSV";
psql -U isucon -d isupipe -c "\COPY isupipe.reactions FROM ./reactions.csv DELIMITER ',' CSV";
psql -U isucon -d isupipe -c "\COPY isupipe.ng_words FROM ./ng_words.csv DELIMITER ',' CSV";
psql -U isucon -d isupipe -c "\COPY isupipe.livecomments FROM ./livecomments.csv DELIMITER ',' CSV";

bash ../../pdns/init_zone.sh


