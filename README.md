# Product Recall Notification System

A system that monitors, scrapes, and notifies subscribers about product recalls in Romania from ANSVSA (National Sanitary Veterinary and Food Safety Authority).

## Features

- Web scraping of official product recalls
- Email subscription system with confirmation
- Rate-limited email notifications
- Resend integration for reliable email delivery
- Batched processing for large subscriber lists
- Unsubscribe functionality

## Prerequisites

- Go 1.23 or higher
- MySQL 8.0
- Docker

## Deployment

The production stack is defined in `docker-compose.yml` and runs Caddy, the
web service, MySQL, and a scraper/notifier job every two hours.

1. Copy `.env.example` to `.env` and set `DB_USER`, `DB_PASSWORD`,
   `DB_ROOT_PASSWORD`, `DB_NAME`, and `RESEND_API_KEY`.
2. Point `produseretrase.eu` and `www.produseretrase.eu` at the server's
   public IP address.
3. Run `docker compose up -d --build`.

Caddy obtains and renews HTTPS certificates automatically. MySQL is reachable
only from the Compose network; do not publish its port on the server.
