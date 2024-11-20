# OTPus Prime

A simple Go-based Telegram bot that integrates with Auth0 for handling OTP codes and managing initial configuration. The server component uses the Gin framework to process updates and requests, while a small React app provides a secure way for users to input and manage Auth0 credentials.

## Overview

This project is focused on building a basic, stateless server in Go that securely handles communication between Telegram and Auth0. It uses HMAC secrets to validate and authenticate requests, helping to keep interactions secure. While the main logic resides in the Go server, the React app plays a minor role, offering a straightforward interface for setting up Auth0.

## Environment Configuration

Set your environment variables in a `.env` file:

```plaintext
# Server Configuration
GIN_MODE=info
BOT_PORT=8080
BASE_URL=https://<BOT_URL>

# Secrets
DEFAULT_SECRET_TOKEN=a-very-long-default-secret-string  # Used to authenticate Telegram messages
HMAC_DEFAULT_SECRET=a-very-long-default-secret-string  # Used to authenticate Auth0 requests

# Required Configuration
TELEGRAM_BOT_TOKEN=your-telegram-bot-token  # Obtain from the BotFather on Telegram

# Telegram Messages Configuration
TELEGRAM_MESSAGE_EXPIRATION_TIME=5  # Time in minutes before messages are deleted. Set to 0 to disable auto-deletion
```

## Key Endpoints

- **/bot/updates**: Handles updates from the Telegram bot. It checks the `x-telegram-bot-api-secret-token` header and processes commands.
- **/auth0/OTPs**: Processes OTP messages sent by Auth0, using HMAC validation for security.
- **/bot/auth-form**: Serves the React app for securely entering credentials to set up Auth0.

## Bash Script for Project Management

The `project.sh` script can be customised to manage common tasks like building, running, testing, cleaning, formatting code, and handling Docker commands.

## Direct Dependencies

- **[github.com/gin-gonic/gin v1.9.1](https://github.com/gin-gonic/gin)**: Used for routing and HTTP server functionality.
- **[github.com/go-resty/resty/v2 v2.11.0](https://github.com/go-resty/resty)**: A simple HTTP client library for making API requests, with support for retries and timeouts.
- **[github.com/joho/godotenv v1.5.1](https://github.com/joho/godotenv)**: A utility for loading environment variables from a `.env` file, making local development easier.
- **[github.com/stretchr/testify v1.8.3](https://github.com/stretchr/testify)**: A toolkit for writing and structuring unit tests in Go.
- **[go.uber.org/zap v1.27.0](https://github.com/uber-go/zap)**: A structured logging library used to log information in a more organised way.

## Telegram Bot Configuration Guide

1. **Register Your Bot**: Use the [BotFather](https://core.telegram.org/bots#botfather) on Telegram to create your bot and get the `TELEGRAM_BOT_TOKEN`.
2. **Set the Webhook**: Use this URL to set up the webhook:
   ```
   https://api.telegram.org/<BOT TOKEN>/setWebhook?secret_token=<DEFAULT_SECRET_TOKEN>&url=<YOUR BOT HOSTNAME>/bot/updates&max_connections=3
   ```
   - Replace `<BOT TOKEN>` with your `TELEGRAM_BOT_TOKEN`.
   - Replace `<DEFAULT_SECRET_TOKEN>` with your `DEFAULT_SECRET_TOKEN` from the `.env` file.
   - Replace `<YOUR BOT HOSTNAME>` with your server's public URL.

3. **Connect to the Internet**: Ensure your bot is accessible online. You can use tools like:
   - **ngrok**: A simple way to expose your local server to the internet temporarily.
   - **Cloudflare Tunnels**: A secure way to connect your bot to the internet.
   - **Other Services**: Feel free to explore other tunneling options.
