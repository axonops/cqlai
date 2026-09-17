package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
)

// The provider table is the one list of providers.
//
// It used to be six lists: what PREFERENCES offers, which block of cqlai.json
// a provider reads, what counts as configured, how its client is built, how a
// conversation with it continues, and how it is asked to explain something. A
// provider in some of those and not the others is one that is offered and then
// fails on the first question, and nothing in the build says so - which is
// what "gemini" was, and still is, with the difference that it is now one
// line of the table saying so rather than a gap in five switches.
//
// These tests walk the table and check every one of those follows from it.

// TestEveryProviderIsReachable: a provider with a transport can be talked to,
// and one without says so instead of failing halfway.
func TestEveryProviderIsReachable(t *testing.T) {
	answers := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"an answer"}}]}`))
	}))
	defer answers.Close()

	for _, p := range providerTable() {
		name := string(p.name)

		settings := &config.AIConfig{Provider: name, Model: "a-model", APIKey: "k", URL: answers.URL}
		if p.speaks == anthropicAPI {
			// Anthropic is the one provider that cannot be pointed at a test
			// server, so it is asked without a key: it then stops before it
			// leaves the process, which is enough to show it was dispatched.
			settings.APIKey = ""
		}

		_, err := Explain(context.Background(), settings, TraceInstructions, "something")

		if p.speaks == noAPI {
			assert.Nil(t, p.continues, "%s has nothing to talk to, so nothing continues a conversation with it", name)
			assert.ErrorContains(t, err, "cannot be asked to explain", "%s", name)

			_, started := GetConversationManager().StartConversation(name, "a-model", "k", "", "a request", "")
			assert.ErrorContains(t, started, name, "%s cannot start a conversation either", name)
			continue
		}

		assert.NotNil(t, p.continues, "%s can be talked to, so a conversation with it must continue somewhere", name)
		if err != nil {
			assert.NotContains(t, err.Error(), "cannot be asked to explain",
				"%s can be talked to, so Explain must not refuse it", name)
		}

		conv, started := GetConversationManager().StartConversation(name, "a-model", "k", "", "a request", "")
		require.NoError(t, started, "%s", name)
		assert.Equal(t, name, conv.Provider)
	}
}

// TestAProviderThatCanBeTalkedToHasSomewhereToTalkTo.
func TestAProviderThatCanBeTalkedToHasSomewhereToTalkTo(t *testing.T) {
	for _, p := range providerTable() {
		name := string(p.name)

		switch p.speaks {
		case openAIAPI:
			assert.NotEmpty(t, p.url, "%s is reached over HTTP, so the table must say where", name)
		case anthropicAPI:
			assert.Empty(t, p.url, "the Anthropic SDK knows where Anthropic is")
		case noAPI:
			assert.Empty(t, p.url, "%s has nowhere to send a question", name)
		}

		if p.speaks != noAPI {
			assert.NotEmpty(t, p.model, "%s needs a model to fall back on", name)
		}
	}
}

// TestProvidersAreTheTable: what PREFERENCES offers is what the table holds,
// in the table's order.
func TestProvidersAreTheTable(t *testing.T) {
	var expected []string
	for _, p := range providerTable() {
		expected = append(expected, string(p.name))
	}

	assert.Equal(t, expected, Providers())
}

// TestEveryProviderBlockIsARealBlock: the name PREFERENCES addresses a
// provider's settings by, and the accessor the CHAT tab reads them through,
// are the same field of cqlai.json.
func TestEveryProviderBlockIsARealBlock(t *testing.T) {
	mine := &config.AIProviderConfig{APIKey: "its own key"}

	for _, p := range providerTable() {
		name := string(p.name)

		if p.section == nil {
			assert.Empty(t, p.field, "%s has no block, so there is no field to name", name)
			continue
		}
		require.NotEmpty(t, p.field, "%s reads a block, so PREFERENCES needs its name", name)

		settings := &config.AIConfig{}
		field := reflect.ValueOf(settings).Elem().FieldByName(p.field)
		require.True(t, field.IsValid(), "config.AIConfig has no field %q", p.field)
		field.Set(reflect.ValueOf(mine))

		assert.Same(t, mine, p.section(settings),
			"%s reads a different block from the one PREFERENCES writes", name)
	}
}

// TestProviderBlocksAreWhatPreferencesOffers.
func TestProviderBlocksAreWhatPreferencesOffers(t *testing.T) {
	blocks := ProviderBlocks()
	require.NotEmpty(t, blocks)

	for _, block := range blocks {
		p, known := providerNamed(Provider(block.Name))
		require.True(t, known, "%s is offered a block and is not a provider", block.Name)
		assert.Equal(t, p.field, block.Field)
	}
}

// TestAProvidersOwnBlockOverridesTheGeneralSettings, for every provider that
// has one.
func TestAProvidersOwnBlockOverridesTheGeneralSettings(t *testing.T) {
	for _, p := range providerTable() {
		if p.section == nil {
			continue
		}
		name := string(p.name)

		settings := &config.AIConfig{
			Provider: name,
			APIKey:   "the general key",
			Model:    "the general model",
			URL:      "the general URL",
		}
		reflect.ValueOf(settings).Elem().FieldByName(p.field).Set(reflect.ValueOf(
			&config.AIProviderConfig{APIKey: "its own key", Model: "its own model", URL: "its own URL"},
		))

		resolved := ConvertDBConfigToAIConfig(settings)

		assert.Equal(t, "its own key", resolved.APIKey, "%s", name)
		assert.Equal(t, "its own model", resolved.Model, "%s", name)
		assert.Equal(t, "its own URL", resolved.URL, "%s", name)
	}
}

// TestWhatIsNotConfiguredComesFromTheTable.
func TestWhatIsNotConfiguredComesFromTheTable(t *testing.T) {
	for _, p := range providerTable() {
		name := string(p.name)

		resolved := ConvertDBConfigToAIConfig(&config.AIConfig{Provider: name})

		assert.Equal(t, p.model, resolved.Model, "%s falls back to the model in the table", name)
		assert.Equal(t, p.url, resolved.URL, "%s falls back to the URL in the table", name)
	}
}

// TestNoProviderAtAllIsMock: nothing built from an empty configuration reaches
// a provider by accident.
func TestNoProviderAtAllIsMock(t *testing.T) {
	assert.Equal(t, string(ProviderMock), ConvertDBConfigToAIConfig(nil).Provider)
	assert.Equal(t, string(ProviderMock), ConvertDBConfigToAIConfig(&config.AIConfig{}).Provider)
}

// TestAnUnknownProviderKeepsWhatItWasGiven: a name nothing here knows is left
// alone rather than quietly turned into something else.
func TestAnUnknownProviderKeepsWhatItWasGiven(t *testing.T) {
	resolved := ConvertDBConfigToAIConfig(&config.AIConfig{
		Provider: "telepathy", APIKey: "k", Model: "m",
	})

	assert.Equal(t, "telepathy", resolved.Provider)
	assert.Equal(t, "k", resolved.APIKey)
	assert.Equal(t, "m", resolved.Model)
	assert.Empty(t, resolved.URL)
}

// TestConfiguredIsWhatTheProviderNeeds: a key, a URL, or nothing, according to
// the table rather than to a switch of its own.
func TestConfiguredIsWhatTheProviderNeeds(t *testing.T) {
	for _, p := range providerTable() {
		name := string(p.name)

		switch p.needs {
		case needsNothing:
			assert.True(t, IsConfigured(&config.AIConfig{Provider: name}), "%s needs nothing", name)
		case needsKey:
			assert.False(t, IsConfigured(&config.AIConfig{Provider: name, URL: "u"}), "%s needs a key", name)
			assert.True(t, IsConfigured(&config.AIConfig{Provider: name, APIKey: "k"}), "%s has a key", name)
		case needsURL:
			assert.False(t, IsConfigured(&config.AIConfig{Provider: name, APIKey: "k"}), "%s needs a URL", name)
			assert.True(t, IsConfigured(&config.AIConfig{Provider: name, URL: "u"}), "%s has a URL", name)
		}
	}
}

// TestAConversationIsContinuedByTheProviderItWasStartedWith.
func TestAConversationIsContinuedByTheProviderItWasStartedWith(t *testing.T) {
	_, _, err := (&AIConversation{Provider: "telepathy", MaxRounds: 5}).
		Continue(context.Background(), "a question")
	assert.ErrorContains(t, err, "telepathy")

	_, _, err = (&AIConversation{Provider: "mock", MaxRounds: 5}).
		Continue(context.Background(), "a question")
	assert.ErrorContains(t, err, "mock", "a provider with no transport cannot continue anything")
}

// TestTheOpenAIAPIIsOneConversation.
//
// OpenRouter had its own copy of it - 314 lines identical to the character
// apart from the names in it - and Ollama had a second, hand-written, which
// posted to the OpenAI-compatible endpoint and then read the reply as though
// it were Ollama's own API. They differ in one thing, the URL, so that is the
// only thing that differs here.
func TestTheOpenAIAPIIsOneConversation(t *testing.T) {
	openAI, _ := providerNamed(ProviderOpenAI)

	urls := map[string]bool{}
	for _, name := range []Provider{ProviderOpenRouter, ProviderGemini, ProviderOllama} {
		p, known := providerNamed(name)
		require.True(t, known)

		assert.Equal(t, openAIAPI, p.speaks, "%s", name)
		assert.Equal(t,
			reflect.ValueOf(openAI.continues).Pointer(),
			reflect.ValueOf(p.continues).Pointer(),
			"%s is continued by the same function as OpenAI", name)

		assert.NotEqual(t, openAI.url, p.url, "%s is somewhere else", name)
		assert.False(t, urls[p.url], "%s shares a URL with another provider", name)
		urls[p.url] = true
	}
}

// TestAConversationGoesWhereTheProviderLives: OpenRouter and OpenAI share the
// conversation, so the base URL is what has to keep them apart.
func TestAConversationGoesWhereTheProviderLives(t *testing.T) {
	var asked struct {
		Model string `json:"model"`
	}

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&asked))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}]}`))
	}))
	defer provider.Close()

	conv, err := GetConversationManager().StartConversation(
		"openrouter", "a-model", "k", provider.URL, "count the rows", "")
	require.NoError(t, err)
	assert.Equal(t, provider.URL, conv.BaseURL)

	_, _, _ = conv.Continue(context.Background(), "count the rows")
	assert.Equal(t, "a-model", asked.Model, "the question went to the URL it was started with")
}

// TestTheDefaultURLIsUsedWhenNoneIsConfigured.
func TestTheDefaultURLIsUsedWhenNoneIsConfigured(t *testing.T) {
	for _, name := range []string{"openai", "openrouter", "gemini", "ollama"} {
		p, _ := providerNamed(Provider(name))

		conv, err := GetConversationManager().StartConversation(name, "a-model", "k", "", "a request", "")
		require.NoError(t, err)
		assert.Equal(t, p.url, conv.BaseURL, "%s", name)
		assert.True(t, strings.HasPrefix(conv.BaseURL, "http"), "%s", name)
	}
}

// TestEveryProviderOnTheOpenAIAPIIsAskedAndReadTheSameWay.
//
// Ollama is the one this was written for. It posted to the OpenAI-compatible
// endpoint - /v1/chat/completions - and then read the reply into a struct
// shaped like Ollama's own API, which has the message at the top level rather
// than under choices. Every reply therefore came back empty, the plan never
// parsed, and the conversation asked for clarification until it ran out of
// rounds. Gemini had no client at all.
func TestEveryProviderOnTheOpenAIAPIIsAskedAndReadTheSameWay(t *testing.T) {
	for _, name := range []string{"openai", "openrouter", "gemini", "ollama"} {
		t.Run(name, func(t *testing.T) {
			var path, authorization string

			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				path = r.URL.Path
				authorization = r.Header.Get("Authorization")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","tool_calls":[{"id":"1","type":"function","function":{"name":"submit_query_plan","arguments":"{\"operation\":\"SELECT\",\"keyspace\":\"app\",\"table\":\"users\",\"read_only\":true,\"confidence\":0.9}"}}]}}]}`))
			}))
			defer provider.Close()

			key := "k"
			if name == "ollama" {
				key = "" // it runs locally and takes no key
			}

			conv, err := GetConversationManager().StartConversation(
				name, "a-model", key, provider.URL, "show me the users", "")
			require.NoError(t, err)

			plan, interaction, err := conv.Continue(context.Background(), "show me the users")

			require.NoError(t, err)
			assert.Nil(t, interaction)
			require.NotNil(t, plan, "the reply is read as chat completions, which is what the endpoint answers")
			assert.Equal(t, "SELECT", plan.Operation)
			assert.Equal(t, "app", plan.Keyspace)
			assert.Equal(t, "users", plan.Table)

			assert.Equal(t, "/chat/completions", path)
			// Ollama takes no key, and its endpoint does not look at the
			// header; the others need it on every request.
			assert.Equal(t, strings.TrimSpace("Bearer "+key), authorization)
		})
	}
}

// TestGeminiIsAskedWhereGeminiIs.
func TestGeminiIsAskedWhereGeminiIs(t *testing.T) {
	gemini, known := providerNamed(ProviderGemini)
	require.True(t, known)

	assert.Equal(t, "https://generativelanguage.googleapis.com/v1beta/openai", gemini.url,
		"the OpenAI-compatible endpoint, not the Gemini API itself")
	assert.NotEqual(t, "gemini-pro", gemini.model, "gemini-pro has been retired")
	assert.True(t, IsConfigured(&config.AIConfig{Provider: "gemini", APIKey: "k"}))
}
