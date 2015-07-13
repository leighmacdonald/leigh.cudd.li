#!/usr/bin/env bash
echo "> Generating new keys..."

openssl req -x509 -nodes -days 365 -newkey rsa:1024 -keyout key_priv.pem -out key_ca.pem