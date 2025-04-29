# ProdTerm

A terminal-based productivity assistant powered by Claude AI for handling daily tasks directly from the command line.

## Features

- Chat with Claude AI directly from your terminal
- Summarize webpages
- Get email summaries from your Gmail account
- Get summaries of Slack channel conversations
- Pretty terminal UI with command history and auto-completion
- Model Context Protocol integration for extensibility

## Installation

### Prerequisites

- Go 1.19 or newer
- Anthropic API key
- For Gmail integration: Google Cloud project with Gmail API enabled
- For Slack integration: Slack API token with appropriate permissions

### Building from source

```bash
# Clone the repository
git clone https://github.com/yourusername/terminal-claude
cd terminal-claude

# Install dependencies
go mod tidy

# Build the application
go build -o terminal-claude
```

## Usage

1. Set your Anthropic API key as an environment variable:
   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```

2. For Gmail integration, follow the setup instructions in [docs/gmail_setup.md](docs/gmail_setup.md)

3. For Slack integration, follow the setup instructions in [docs/slack_setup.md](docs/slack_setup.md)

4. Run the application:
   ```bash
   ./prodterm
   ```

5. Enter commands at the prompt:
   ```
   > what's on this webpage? bbc.co.uk
   > summarise my unread e-mails
   > list slack channels
   > summarise slack channel #general
   > tell me about golang
   ```

6. Press Ctrl+C or type `exit` to quit

## Configuration

You can configure the Claude model by setting the `CLAUDE_MODEL` environment variable:
```bash
export CLAUDE_MODEL="claude-3-opus-20240229"
```

The default model is `claude-3-sonnet-20240229` if not specified.

Available models:
- claude-3-opus-20240229
- claude-3-sonnet-20240229
- claude-3-haiku-20240307

## Model Context Protocol Providers

ProdTerm uses the Model Context Protocol to integrate with various services:

- **Gmail**: Access and summarize your emails
- **Slack**: Access and summarize your Slack channel discussions
- More providers coming soon!

## Development

To add a new provider, implement the `mcp.Provider` interface and register it in `main.go`.
