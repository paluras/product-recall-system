#!/bin/sh
set -eu

# Cron does not inherit the container environment. Keep only the variables
# needed by the scheduled job in a root-readable runtime file.
env | grep -E '^(DB_USER|DB_PASSWORD|DB_HOST|DB_NAME|DB_PORT|RESEND_API_KEY)=' > /etc/product-recall.env
chmod 600 /etc/product-recall.env

echo '0 */2 * * * . /etc/product-recall.env; /app/run-scraper-and-notify.sh >> /proc/1/fd/1 2>&1' | crontab -
exec cron -f
