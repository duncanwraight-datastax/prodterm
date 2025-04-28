# Terminal Claude Assistant

A terminal-based assistant powered by Claude AI for handling daily tasks directly from the command line.

## Features

- Chat with Claude AI directly from your terminal
- Summarize webpages
- Get email summaries from your Gmail account
- Pretty terminal UI with command history and auto-completion
- Model Context Protocol integration for extensibility

## Installation

### Prerequisites

- Go 1.19 or newer
- Anthropic API key
- For Gmail integration: Google Cloud project with Gmail API enabled

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

3. Run the application:
   ```bash
   ./terminal-claude
   ```

4. Enter commands at the prompt:
   ```
   > what's on this webpage? bbc.co.uk
   > summarise my unread e-mails
   > tell me about golang
   ```

5. Press Ctrl+C or type `exit` to quit

## Configuration

You can configure the Claude model by setting the `CLAUDE_MODEL` environment variable:
```bash
export CLAUDE_MODEL="claude-3-opus-20240229"
```

The default model is `claude-3-sonnet-20240229` if not specified.

## Model Context Protocol Providers

Terminal Claude uses the Model Context Protocol to integrate with various services:

- **Gmail**: Access and summarize your emails
- More providers coming soon!

## Development

To add a new provider, implement the `mcp.Provider` interface and register it in `main.go`.
