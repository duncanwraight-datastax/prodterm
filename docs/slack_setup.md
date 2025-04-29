# Setting up Slack Integration

This document explains how to set up the Slack integration for ProdTerm.

## Getting a Slack API Token

1. Visit the [Slack API Apps page](https://api.slack.com/apps)
2. Click "Create New App" and select "From scratch"
3. Give your app a name (e.g., "ProdTerm") and select the workspace you want to use it in
4. Click "Create App"

### Required Scopes

After creating your app, you need to add the following OAuth scopes:

1. Go to "OAuth & Permissions" in the sidebar
2. Under "Bot Token Scopes", add these scopes:
   - `channels:history` - View messages in public channels
   - `channels:read` - View basic information about public channels
   - `groups:history` - View messages in private channels
   - `groups:read` - View basic information about private channels
   - `users:read` - View basic information about users

### Installing the App to Your Workspace

1. Go to "OAuth & Permissions" in the sidebar
2. Click "Install to Workspace"
3. Review the permissions and click "Allow"
4. Copy the "Bot User OAuth Token" that starts with `xoxb-`

## Setting Up ProdTerm to Use Slack

You have two options for providing the Slack token to ProdTerm:

### Option 1: Environment Variable

Set the `SLACK_TOKEN` environment variable to your Slack API token:

```bash
export SLACK_TOKEN=xoxb-your-token-here
```

### Option 2: Configuration File

Create a file at `~/.config/terminal-claude/slack_token.txt` containing just your Slack API token:

```bash
mkdir -p ~/.config/terminal-claude
echo "xoxb-your-token-here" > ~/.config/terminal-claude/slack_token.txt
```

## Using Slack in ProdTerm

Once configured, you can use the following commands:

- `list slack channels` - Shows available Slack channels
- `summarize slack channel #general` - Provides a summary of recent messages in a channel

You can refer to channels by their name (e.g., `#general`) or their ID (e.g., `C12345678`).