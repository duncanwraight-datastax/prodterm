# Gmail Integration Setup

This guide explains how to set up the Gmail integration for Terminal Claude.

## Creating Google API Credentials

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Gmail API for your project:
   - In the sidebar, click on "APIs & Services" > "Library"
   - Search for "Gmail API"
   - Click on it and press "Enable"
4. Create credentials:
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth client ID"
   - Select "Desktop app" as Application type
   - Give it a name (e.g., "Terminal Claude")
   - Click "Create"
5. Download the credentials as JSON by clicking the download icon

## Setting Up Gmail Authentication

1. Create the config directory (if it doesn't exist):
   ```bash
   mkdir -p ~/.config/terminal-claude
   ```

2. Copy your downloaded credentials to the config directory:
   ```bash
   cp ~/Downloads/client_secret_xxxxxxxx.json ~/.config/terminal-claude/gmail_credentials.json
   ```

3. Alternatively, set the path via environment variable:
   ```bash
   export GMAIL_CREDENTIALS="/path/to/your/credentials.json"
   ```

## First Run Authentication

When you first run Terminal Claude and try to access Gmail features, the application will:

1. Open a web browser or provide a URL to visit
2. Prompt you to log in to your DataStax Google account
3. Ask for permission to access your Gmail data
4. Give you an authorization code
5. Prompt you to enter this code back in the terminal

This authentication happens only once. The app will save the token for future use.

## Using Gmail Features

After authentication, you can use commands like:

```
> summarise my unread e-mails
```

## Troubleshooting

If you encounter authentication issues:

1. Ensure your credentials file is correct
2. Remove the stored token to force re-authentication:
   ```bash
   rm ~/.config/terminal-claude/gmail_token.json
   ```
3. Check that you've enabled the Gmail API for your project
