package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	anthropicoption "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/openai/openai-go"
	openaioption "github.com/openai/openai-go/option"

	"github.com/axonops/cqlai/internal/config"
)

// Asking the model to explain something, in plain words.
//
// Everything else here asks for a query plan: a question goes out with the
// tools, and what comes back is parsed. An explanation has nothing to parse -
// it is read - and the tool loop is in the way of getting one.
//
// The clients are built the way the conversation builds them, from the SDKs
// directly. There is a second set of provider clients in this package behind
// an AIClient interface, and nothing reaches it: going through that instead
// put a second copy of four SDKs' worth of machinery in the binary - five
// megabytes - for the same three requests.

// Explain asks the configured provider to explain something.
func Explain(ctx context.Context, providerConfig *config.AIConfig, instructions, subject string) (string, error) {
	if strings.TrimSpace(subject) == "" {
		return "", fmt.Errorf("there is nothing to explain")
	}

	settings := ConvertDBConfigToAIConfig(providerConfig)
	key, model := settings.APIKey, settings.Model

	var said string
	var err error

	switch Provider(settings.Provider) {
	case ProviderAnthropic:
		said, err = explainWithAnthropic(ctx, key, model, instructions, subject)
	case ProviderOpenAI:
		said, err = explainWithOpenAI(ctx, key, model, openAiBaseURL, settings.URL, instructions, subject)
	case ProviderOpenRouter:
		said, err = explainWithOpenAI(ctx, key, model, openRouterBaseURL, settings.URL, instructions, subject)
	case ProviderOllama:
		said, err = explainWithOpenAI(ctx, key, model, ollamaBaseURL, settings.URL, instructions, subject)
	default:
		return "", fmt.Errorf("the %s provider cannot be asked to explain things", settings.Provider)
	}

	if err != nil {
		return "", err
	}
	if strings.TrimSpace(said) == "" {
		return "", fmt.Errorf("the answer came back empty")
	}
	return said, nil
}

// explainWithAnthropic asks Claude.
func explainWithAnthropic(ctx context.Context, key, model, instructions, subject string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("no Anthropic API key is configured")
	}
	if model == "" {
		model = DefaultAnthropicModel
	}

	client := anthropic.NewClient(anthropicoption.WithAPIKey(key))
	response, err := askAnthropic(ctx, &client, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: anthropicMaxTokens,
		System:    []anthropic.TextBlockParam{{Text: instructions}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(subject)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("anthropic API error: %v", err)
	}

	var text strings.Builder
	for _, content := range response.Content {
		if content.Type == "text" {
			text.WriteString(content.Text)
		}
	}
	return text.String(), nil
}

// explainWithOpenAI asks anything that speaks the OpenAI API: OpenAI itself,
// OpenRouter, and Ollama.
func explainWithOpenAI(ctx context.Context, key, model, standardURL, configuredURL, instructions, subject string) (string, error) {
	url := configuredURL
	if url == "" {
		url = standardURL
	}
	if model == "" {
		return "", fmt.Errorf("no model is configured")
	}

	client := openai.NewClient(
		openaioption.WithAPIKey(key),
		openaioption.WithBaseURL(url),
	)

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(instructions),
			openai.UserMessage(subject),
		},
	})
	if err != nil {
		return "", fmt.Errorf("API error: %v", err)
	}
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("the answer came back empty")
	}
	return completion.Choices[0].Message.Content, nil
}

// TraceInstructions is what to make of a query trace.
//
// The shape of the answer is dictated, not suggested. Asked for an explanation
// in a word limit, the models write an essay about the query: paragraphs that
// restate the trace the reader already has in front of them. What is wanted is
// the few lines that are not in the trace - where the time went, what that
// means, and what to do - in a pane a few lines high beside it.
const TraceInstructions = `You are reading a Cassandra query trace from system_traces, for an engineer who has the trace on screen beside your answer.

Answer in exactly these three sections, with these headings, in this order, and write nothing else:

TIME
The steps that took the time, longest share first, at most four lines. One line each, in this shape:
  1453us  submitting range requests  Native-Transport-Requests-1
Use the elapsed figures from the trace. Where two steps are on different threads, keep the thread name.
A trace with a Page column is one query read a page at a time, each page a request of its own: start this section with a line saying what a page costs and how many there were, in this shape, and then the steps of the page that cost the most.
  6 pages  ~1.0ms each  6.1ms total

FINDINGS
What the trace shows that the timings do not say on their own, at most four lines, each starting with "- ". Tombstones read, a read repair, a range scan where a partition lookup was possible, more replicas than the consistency level needs, a timeout, a retry, a page size so small that the query cost a request per handful of rows. If there is nothing of the kind, write exactly "- Nothing remarkable."

WHAT TO DO
The changes worth making, at most three lines, each starting with "- ", each a change someone could make: an index on a named column, a different partition key, a consistency level, a page size. Name the column or the setting. If the query is already doing the right thing, write exactly "- Nothing to change."

Rules:
- No preamble, no closing summary, no markdown, no bold, no numbered lists.
- Every line under 100 characters.
- Say microseconds as us and milliseconds as ms, one decimal place at most.
- Never restate the whole trace. The reader has it.
- Where the trace covers several pages, the numbers to reason about are the totals across them, not one page's: a page that read 65 rows is not a query that read 65 rows.`

// SchemaInstructions is what to make of a table's definition.
//
// The same discipline as the trace: the shape of the answer is dictated, and
// it is read in a pane beside the definition it is about rather than instead
// of it.
const SchemaInstructions = `You are reading a Cassandra table or keyspace definition, for an engineer who has it on screen beside your answer.

Answer in exactly these three sections, with these headings, in this order, and write nothing else:

WHAT IT IS
One line: what the partition key is, what the rows inside a partition are ordered by, and what the table appears to model. Two lines at most.

RISKS
What will go wrong with this definition, at most four lines, each starting with "- ", worst first. The things that punish a Cassandra cluster months later:
- a partition that grows without bound, or one row per partition
- a partition key with too few distinct values, or one value that takes most of the traffic
- a clustering order that cannot serve the query the table is obviously for
- a collection that grows without bound, or a non-frozen collection read whole
- a secondary index on a high-cardinality or very low-cardinality column
- a compaction strategy that does not match how the data is written and expired: time series without TWCS, heavy overwrites without LCS
- a default TTL that will leave tombstones in the read path
- text where a timeuuid, timestamp or uuid is meant
If nothing of the kind is there, write exactly "- Nothing that will bite."

WHAT TO CHANGE
The changes worth making, at most three lines, each starting with "- ", each naming the column or the setting. Say when a change means writing the data again into a new table, because in Cassandra it usually does. If the definition is sound, write exactly "- Nothing to change."

Rules:
- No preamble, no closing summary, no markdown, no bold, no numbered lists.
- Every line under 100 characters.
- Judge what is in front of you. Do not invent queries the table might serve; where the shape only makes sense for a query you cannot see, say so as a risk.
- Never restate the definition. The reader has it.`

// ReviewSchema asks what a definition will do to whoever runs it.
func ReviewSchema(ctx context.Context, providerConfig *config.AIConfig, definition string) (string, error) {
	return Explain(ctx, providerConfig, SchemaInstructions, definition)
}

// AnalyseTrace asks what a query trace shows.
func AnalyseTrace(ctx context.Context, providerConfig *config.AIConfig, trace string) (string, error) {
	return Explain(ctx, providerConfig, TraceInstructions, trace)
}
