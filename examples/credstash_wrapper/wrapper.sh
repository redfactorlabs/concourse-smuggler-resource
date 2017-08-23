#!/bin/sh

set -e
set -u

SCRIPT_PATH="$(cd "$(dirname $0)"; pwd)"

if [ "${SMUGGLER_credstash_aws_iam_profile:-}" != "true" ]; then
	export AWS_ACCESS_KEY_ID="${SMUGGLER_credstash_aws_access_key_id}"
	export AWS_SECRET_ACCESS_KEY="${SMUGGLER_credstash_aws_secret_access_key}"
	export AWS_SESSION_TOKEN="${SMUGGLER_credstash_aws_session_token:-}"
fi
"${SCRIPT_PATH}"/unicreds \
	-r ${SMUGGLER_credstash_region:-eu-west-1} \
	${SMUGGLER_credstash_table:+-t ${SMUGGLER_credstash_table}} \
	exec "${SCRIPT_PATH}"/spruce merge | \
		"${SCRIPT_PATH}"/spruce json | \
		/opt/resource/"${SMUGGLER_COMMAND}.wrapped" $@
