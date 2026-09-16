package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
)

// Asking the model to explain something.
//
// The clients are built the way the conversation builds them, from the SDKs
// directly. There is a second set of provider clients in this package behind
// an AIClient interface and nothing reaches it: going through that instead put
// a second copy of four SDKs' worth of machinery in the binary - five
// megabytes - for the same three requests.

// TestThereIsNothingToExplainWithoutSomethingToExplain.
func TestThereIsNothingToExplainWithoutSomethingToExplain(t *testing.T) {
	_, err := Explain(context.Background(), &config.AIConfig{Provider: "anthropic"}, TraceInstructions, "   ")

	assert.ErrorContains(t, err, "nothing to explain")
}

// TestAProviderThatCannotBeAskedSaysWhichOneItIs.
func TestAProviderThatCannotBeAskedSaysWhichOneItIs(t *testing.T) {
	_, err := Explain(context.Background(), &config.AIConfig{Provider: "mock"}, TraceInstructions, "a trace")

	assert.ErrorContains(t, err, "mock")
	assert.ErrorContains(t, err, "cannot be asked to explain")
}

// TestTheQuestionIsAskedOfTheProviderAndTheAnswerComesBack.
//
// Against a server speaking the OpenAI API, which is what OpenAI, OpenRouter
// and Ollama are all asked through.
func TestTheQuestionIsAskedOfTheProviderAndTheAnswerComesBack(t *testing.T) {
	var asked struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&asked))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"Most of the time went on reading sstables."}}]}`))
	}))
	defer provider.Close()

	said, err := AnalyseTrace(context.Background(), &config.AIConfig{
		Provider: "openai", APIKey: "k", Model: "gpt-4o", URL: provider.URL,
	}, "Read 3 sstables | 10.0.0.2 | 9100")

	require.NoError(t, err)
	assert.Equal(t, "Most of the time went on reading sstables.", said)

	assert.Equal(t, "gpt-4o", asked.Model)
	require.Len(t, asked.Messages, 2)
	assert.Equal(t, "system", asked.Messages[0].Role)
	assert.Contains(t, asked.Messages[0].Content, "system_traces", "the instructions say what it is reading")
	assert.Equal(t, "Read 3 sstables | 10.0.0.2 | 9100", asked.Messages[1].Content)
}

// TestAnEmptyAnswerIsNotAnAnswer.
func TestAnEmptyAnswerIsNotAnAnswer(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"  "}}]}`))
	}))
	defer provider.Close()

	_, err := AnalyseTrace(context.Background(), &config.AIConfig{
		Provider: "ollama", Model: "llama3", URL: provider.URL,
	}, "a trace")

	assert.ErrorContains(t, err, "came back empty")
}

// TestTheTraceInstructionsDictateTheShapeOfTheAnswer.
//
// Asked for an explanation in a word limit, the models write an essay about
// the query: paragraphs restating the trace the reader already has in front of
// them. The three headings, the line shapes and the line counts are what turn
// that into something to read in a pane a few lines high.
func TestTheTraceInstructionsDictateTheShapeOfTheAnswer(t *testing.T) {
	for _, wanted := range []string{
		"system_traces",
		"TIME", "FINDINGS", "WHAT TO DO",
		"at most four lines", "at most three lines",
		"Tombstones read", "read repair", "range scan",
		"No preamble", "no markdown",
		"Never restate the whole trace",
		"Page column", "not one page's",
	} {
		assert.Contains(t, TraceInstructions, wanted, "the instructions should say %q", wanted)
	}
}
