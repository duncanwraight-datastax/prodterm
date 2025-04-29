package gmail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"terminal-claude/mcp"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Provider implements the Model Context Protocol for Gmail
type Provider struct {
	service *gmail.Service
}

// New creates a new Gmail provider
func New() (*Provider, error) {
	ctx := context.Background()
	
	// Get credentials from environment variable or file
	credentialsPath := os.Getenv("GMAIL_CREDENTIALS")
	if credentialsPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to get home directory: %v", err)
		}
		credentialsPath = filepath.Join(homeDir, ".config", "terminal-claude", "gmail_credentials.json")
	}

	b, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	// Parse the credentials
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file: %v", err)
	}

	// Get token from file or generate new one
	token, err := getToken(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get token: %v", err)
	}

	// Create the Gmail service
	srv, err := gmail.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %v", err)
	}

	return &Provider{
		service: srv,
	}, nil
}

// Name returns the provider's name
func (p *Provider) Name() string {
	return "Gmail"
}

// GetCapabilities returns the provider's capabilities
func (p *Provider) GetCapabilities() []mcp.Capability {
	return []mcp.Capability{
		{
			Name:        "email",
			Description: "Access and manipulate email",
			Commands:    []string{"list_unread", "get_email", "summarize_unread"},
		},
	}
}

// Execute runs a command with the given parameters
func (p *Provider) Execute(command string, params map[string]interface{}) (interface{}, error) {
	switch command {
	case "list_unread":
		return p.listUnreadEmails()
	case "get_email":
		id, ok := params["id"].(string)
		if !ok {
			return nil, fmt.Errorf("email id parameter required")
		}
		return p.getEmail(id)
	case "summarize_unread":
		count := 10 // Default count
		if c, ok := params["count"].(float64); ok {
			count = int(c)
		}
		return p.summarizeUnreadEmails(count)
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// listUnreadEmails lists unread emails
func (p *Provider) listUnreadEmails() ([]map[string]interface{}, error) {
	user := "me"
	r, err := p.service.Users.Messages.List(user).Q("is:unread").MaxResults(10).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %v", err)
	}

	var messages []map[string]interface{}
	for _, m := range r.Messages {
		msg, err := p.service.Users.Messages.Get(user, m.Id).Format("metadata").Do()
		if err != nil {
			continue
		}

		email := map[string]interface{}{
			"id": msg.Id,
		}

		// Extract headers (From, To, Subject, Date)
		for _, header := range msg.Payload.Headers {
			switch header.Name {
			case "From", "To", "Subject", "Date":
				email[strings.ToLower(header.Name)] = header.Value
			}
		}

		messages = append(messages, email)
	}

	return messages, nil
}

// getEmail gets a specific email by ID
func (p *Provider) getEmail(id string) (map[string]interface{}, error) {
	user := "me"
	msg, err := p.service.Users.Messages.Get(user, id).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve message: %v", err)
	}

	email := map[string]interface{}{
		"id":       msg.Id,
		"snippet":  msg.Snippet,
		"threadId": msg.ThreadId,
	}

	// Extract headers
	for _, header := range msg.Payload.Headers {
		switch header.Name {
		case "From", "To", "Subject", "Date":
			email[strings.ToLower(header.Name)] = header.Value
		}
	}

	// Extract body
	if msg.Payload.Body != nil && msg.Payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
		if err == nil {
			email["body"] = string(data)
		}
	}

	// Try to extract body from parts if not found in payload
	if _, ok := email["body"]; !ok {
		var findBodyPart func(parts []*gmail.MessagePart) string
		findBodyPart = func(parts []*gmail.MessagePart) string {
			for _, part := range parts {
				if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
					data, err := base64.URLEncoding.DecodeString(part.Body.Data)
					if err == nil {
						return string(data)
					}
				}
				if len(part.Parts) > 0 {
					if body := findBodyPart(part.Parts); body != "" {
						return body
					}
				}
			}
			return ""
		}

		if msg.Payload.Parts != nil {
			if body := findBodyPart(msg.Payload.Parts); body != "" {
				email["body"] = body
			}
		}
	}

	return email, nil
}

// summarizeUnreadEmails gets a summary of unread emails
func (p *Provider) summarizeUnreadEmails(count int) (map[string]interface{}, error) {
	fmt.Println("DEBUG - Gmail provider: Fetching unread emails")
	
	user := "me"
	r, err := p.service.Users.Messages.List(user).Q("is:unread").MaxResults(int64(count)).Do()
	if err != nil {
		fmt.Printf("DEBUG - Gmail error: %v\n", err)
		return nil, fmt.Errorf("unable to retrieve messages: %v", err)
	}
	
	fmt.Printf("DEBUG - Found %d unread messages\n", len(r.Messages))

	var emails []map[string]interface{}
	for _, m := range r.Messages {
		msg, err := p.service.Users.Messages.Get(user, m.Id).Format("metadata").Do()
		if err != nil {
			continue
		}

		email := map[string]interface{}{
			"id": msg.Id,
		}

		// Extract headers
		for _, header := range msg.Payload.Headers {
			switch header.Name {
			case "From", "To", "Subject", "Date":
				email[strings.ToLower(header.Name)] = header.Value
			}
		}

		email["snippet"] = msg.Snippet
		emails = append(emails, email)
	}

	return map[string]interface{}{
		"count":  len(emails),
		"emails": emails,
	}, nil
}

// getTokenFromFile retrieves a token from a local file
func getTokenFromFile() (*oauth2.Token, error) {
	tokenPath := getTokenPath()
	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// saveToken saves a token to a file
func saveToken(token *oauth2.Token) error {
	tokenPath := getTokenPath()
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	
	f, err := os.OpenFile(tokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// getToken gets an OAuth token
func getToken(config *oauth2.Config) (*oauth2.Token, error) {
	// Try to read token from file
	token, err := getTokenFromFile()
	if err == nil {
		return token, nil
	}

	// If no token found, get one from user
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser: \n%v\n", authURL)
	fmt.Println("Enter the authorization code:")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	token, err = config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	
	if err := saveToken(token); err != nil {
		log.Printf("Warning: unable to save token: %v", err)
	}
	
	return token, nil
}

// getTokenPath returns the path to the token file
func getTokenPath() string {
	tokenPath := os.Getenv("GMAIL_TOKEN")
	if tokenPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Unable to get home directory: %v", err)
		}
		tokenPath = filepath.Join(homeDir, ".config", "terminal-claude", "gmail_token.json")
	}
	return tokenPath
}
