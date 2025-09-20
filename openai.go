package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientOption func(*LLMClient) error

type LLMClient struct {
	ctx        context.Context
	baseURL    string
	key        string
	httpClient doer // http.Client
	middleware []option.Middleware

	openaiC openai.Client
}

func LLM(ctx context.Context, opts ...ClientOption) (*LLMClient, error) {
	out := &LLMClient{
		ctx: ctx,
	}

	for _, opt := range opts {
		if err := opt(out); err != nil {
			return nil, err
		}
	}

	copts := []option.RequestOption{}

	if out.baseURL != "" {
		copts = append(copts, option.WithBaseURL(out.baseURL))
	}

	if out.key != "" {
		copts = append(copts, option.WithAPIKey(out.key))
	}

	if out.httpClient != nil {
		copts = append(copts, option.WithHTTPClient(out.httpClient))
	}

	if len(out.middleware) > 0 {
		copts = append(copts, option.WithMiddleware(out.middleware...))
	}

	out.openaiC = openai.NewClient(copts...)

	return out, nil
}

func GenerateSchema[T any]() any {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

type Output[T any] struct {
	Data T `json:"data"`
}

func Single[T any](
	ctx context.Context, c *LLMClient,
	model, desc string,
	system []string,
	context ...string,
) (T, error) {
	msgs := []openai.ChatCompletionMessageParamUnion{}
	for _, c := range system {
		msgs = append(msgs, openai.ChatCompletionMessageParamUnion{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: openai.String(c),
				},
			},
		})
	}

	for _, c := range context {
		msgs = append(msgs, openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(c),
				},
			},
		})
	}

	var schema = GenerateSchema[*Output[T]]()
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "request",
		Description: openai.String(desc),
		Schema:      schema,
		Strict:      openai.Bool(false),
	}

	var data T
	chat, err := c.openaiC.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: msgs,
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		},
		Model: model,
	})
	if err != nil {
		return data, err
	}

	if len(chat.Choices) == 0 {
		return data, errors.New("no choices returned")
	}

	output := &Output[T]{}
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &output)
	if err != nil {
		return data, err
	}

	return output.Data, nil
}
