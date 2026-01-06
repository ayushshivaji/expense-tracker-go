package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getMessageBody(payload *gmail.MessagePart) string {
	var body string

	// If body data exists directly
	if payload.Body != nil && payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	// Recursively check parts
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" {
			if part.Body != nil && part.Body.Data != "" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					body = string(data)
				}
			}
		} else if part.MimeType == "text/html" && body == "" {
			// Fallback to HTML if no plain text
			if part.Body != nil && part.Body.Data != "" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					body = string(data)
				}
			}
		} else if part.MimeType == "multipart/alternative" || part.MimeType == "multipart/mixed" || part.MimeType == "multipart/related" {
			// Recurse into nested parts
			body = getMessageBody(part)
		}
	}

	return body
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	dec := json.NewDecoder(f)
	err = dec.Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("./desktop-app.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}
	if len(r.Labels) == 0 {
		fmt.Println("No labels found.")
		return
	}
	// fmt.Println("Labels:")
	// for _, l := range r.Labels {
	// 	fmt.Printf("- %s\n %s\n", l.Id, l.Name)
	// }

	messages, err := srv.Users.Messages.List(user).MaxResults(1).LabelIds("Label_1057535226825725461").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}
	if len(messages.Messages) == 0 {
		fmt.Println("No messages found.")
		return
	}
	// fmt.Println(len(messages.Messages))
	// for _, m := range messages.Messages {
	// 	msg, err := srv.Users.Messages.Get(user, m.Id).Format("full").Do()
	// 	if err != nil {
	// 		log.Fatalf("Unable to retrieve message: %v", err)
	// 	}
	// 	fmt.Println("Message snippet: %s", msg.Payload.Body)
	// }
	// fmt.Println("Messages:", messages.Messages)
	for _, m := range messages.Messages {
		msg, err := srv.Users.Messages.Get(user, m.Id).Format("full").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve message: %v", err)
		}
		fmt.Println(getMessageBody(msg.Payload))
	}
}
