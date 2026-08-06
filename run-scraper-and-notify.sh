#!/bin/bash
/app/scraper -dbuser "$DB_USER" -dbpass "$DB_PASSWORD" -dbhost "$DB_HOST" -dbport "$DB_PORT" -dbname "$DB_NAME"
/app/notify -dbuser "$DB_USER" -dbpass "$DB_PASSWORD" -dbhost "$DB_HOST" -dbport "$DB_PORT" -dbname "$DB_NAME"
