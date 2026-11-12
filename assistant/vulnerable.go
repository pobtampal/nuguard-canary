// Package assistant is a toy support assistant.
//
// INTENTIONALLY VULNERABLE TEST CODE (see README.md). Each function in this
// file contains exactly one planted AI vulnerability, marked CANARY, and is
// otherwise written safely, so each finding has one cause.
package assistant

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pobtampal/nuguard-canary/llm"
)

const model = "gpt-4o-mini"

// SummarizeTicket summarises a customer's support ticket.
//
// CANARY 1 (nuguard-go-llm-prompt-injection-sprintf, LLM01): the customer's
// text is formatted straight into the prompt, so a ticket saying "ignore
// your instructions and ..." becomes part of the instructions.
func SummarizeTicket(ctx context.Context, client *llm.Client, ticketText string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	prompt := fmt.Sprintf("You are a support agent. Summarise this ticket in one sentence:\n%s", ticketText)
	return client.CreateChatCompletion(ctx, llm.ChatRequest{
		Model:    model,
		Messages: []llm.Message{{Role: "user", Content: prompt}},
	})
}

// TranslateReply translates an agent's reply for the customer.
//
// CANARY 2 (nuguard-go-llm-prompt-injection-sprintf, hand-written provider
// call): the same injection, through a helper that posts to the provider
// directly instead of an SDK client.
func TranslateReply(ctx context.Context, replyText, language string) (string, error) {
	prompt := fmt.Sprintf("Translate the following reply into %s:\n%s", language, replyText)
	return callOpenAI(ctx, prompt)
}

// NewHardcodedClient builds a client.
//
// CANARY 3 (nuguard-go-hardcoded-api-key): the API key is written in the
// source. The value is a FAKE placeholder; it is not, and never was, a key.
func NewHardcodedClient() *llm.Client {
	cfg := llm.DefaultConfig("")
	cfg.APIKey = "sk-test-NOT-A-REAL-KEY-nuguard-canary"
	return llm.NewClientWithConfig(cfg)
}

// ClassifyNoTimeout labels the canned health-check message.
//
// CANARY 4 (nuguard-go-llm-missing-context-timeout, LLM04): the call uses a
// root context with no deadline, so a slow or hung provider blocks forever.
func ClassifyNoTimeout(client *llm.Client) (string, error) {
	req := llm.ChatRequest{
		Model:    model,
		Messages: []llm.Message{{Role: "user", Content: "Reply with the single word: ok"}},
	}
	return client.CreateChatCompletion(context.Background(), req)
}

// AskOverInsecureTLS sends the canned health-check message through a custom
// HTTP client.
//
// CANARY 5 (nuguard-go-llm-insecure-tls): certificate verification is
// switched off for the connection that carries prompts, replies and the API
// key, so anyone on the path can read or change them.
func AskOverInsecureTLS(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cfg := llm.DefaultConfig(os.Getenv("LLM_API_KEY"))
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	httpClient := &http.Client{
		Transport: transport,
	}
	cfg.HTTPClient = httpClient
	client := llm.NewClientWithConfig(cfg)
	req := llm.ChatRequest{
		Model:    model,
		Messages: []llm.Message{{Role: "user", Content: "Reply with the single word: ok"}},
	}
	return client.CreateChatCompletion(ctx, req)
}

// callOpenAI posts a prompt to the provider without an SDK. It is a plain
// helper; the flaw is in what CANARY 2 passes to it.
func callOpenAI(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	body, err := json.Marshal(llm.ChatRequest{Model: model, Messages: []llm.Message{{Role: "user", Content: prompt}}})
	if err != nil {
		return "", err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://llm.invalid/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	r.Header.Set("Authorization", "Bearer "+os.Getenv("LLM_API_KEY"))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return resp.Status, nil
}
