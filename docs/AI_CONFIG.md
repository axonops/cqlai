# AI Configuration for CQLAI

CQLAI supports AI-powered natural language to CQL query generation. You can configure your preferred AI provider using environment variables.

## Usage

To use AI features, type `.ai` followed by your natural language request:

```
> .ai show all users with age greater than 25
```

This will:
1. Generate a CQL query plan based on your request
2. Show you the generated CQL for preview
3. Allow you to execute, edit, or cancel the query

## Configuration

AI providers are configured in the `cqlai.json` configuration file. Copy `cqlai.json.example` to `cqlai.json` and update the AI section:

### Basic Configuration
```json
{
  "host": "127.0.0.1",
  "port": 9042,
  ...
  "ai": {
    "provider": "mock",  // Options: mock, openai, anthropic, gemini, ollama, openrouter
    "apiKey": "",        // General API key (overridden by provider-specific)
    "model": "",         // General model (overridden by provider-specific)
    "url": ""            // General URL (overridden by provider-specific, for Ollama/OpenRouter)
  }
}
```

### OpenAI Configuration
```json
"ai": {
  "provider": "openai",
  "openai": {
    "apiKey": "your-openai-api-key-here",
    "model": "gpt-4-turbo-preview"  // Optional, defaults to gpt-4-turbo-preview
  }
}
```

### Anthropic Configuration
```json
"ai": {
  "provider": "anthropic",
  "anthropic": {
    "apiKey": "your-anthropic-api-key-here",
    "model": "claude-3-sonnet-20240229"  // Optional
  }
}
```

### Google Gemini Configuration
```json
"ai": {
  "provider": "gemini",
  "gemini": {
    "apiKey": "your-gemini-api-key-here",
    "model": "gemini-pro"  // Optional, defaults to gemini-pro
  }
}
```

### Ollama Configuration
```json
"ai": {
  "provider": "ollama",
  "ollama": {
    "model": "llama3.2",  // Optional, model to use
    "url": "http://localhost:11434/v1"  // Optional, defaults to http://localhost:11434/v1
  }
}
```

### OpenRouter Configuration
```json
"ai": {
  "provider": "openrouter",
  "openrouter": {
    "apiKey": "your-openrouter-api-key-here",
    "model": "anthropic/claude-3-sonnet",  // Optional
    "url": "https://openrouter.ai/api/v1"  // Optional, defaults to https://openrouter.ai/api/v1
  }
}
```

### Mock Provider (Default)
The mock provider is used by default for testing and doesn't require any API keys. It generates simple example queries based on keywords in your request:

```json
"ai": {
  "provider": "mock"
}
```

## Features

### Query Plan Generation
The AI generates a structured query plan that includes:
- Operation type (SELECT, INSERT, UPDATE, DELETE, CREATE, etc.)
- Target keyspace and table
- Columns to select or modify
- WHERE conditions
- ORDER BY clauses
- LIMIT specifications
- Confidence level

### Safety Features
- **Read-only by default**: The AI prefers SELECT queries unless explicitly asked to modify data
- **Dangerous operation warnings**: Destructive operations (DROP, DELETE, TRUNCATE) show warnings
- **Confirmation required**: Dangerous operations require additional confirmation if enabled
- **Schema validation**: Queries are validated against your current Cassandra schema

### Modal Controls
When the AI generates a query, you can:
- **P**: Toggle between viewing the CQL query and the JSON query plan
- **Enter**: Execute the query
- **Tab/Arrow Keys**: Navigate between Cancel, Execute, and Edit buttons
- **Edit**: Put the generated CQL into the input for manual editing
- **Esc**: Cancel without executing

## Implementation Status

Currently implemented:
- ✅ Mock provider for testing
- ✅ OpenAI API integration (GPT-4, GPT-3.5)
- ✅ Anthropic API integration (Claude 3)
- ✅ Google Gemini API integration
- ✅ Ollama support (local models and OpenAI-compatible APIs)
- ✅ OpenRouter integration (multiple model access)
- ✅ Query plan generation and validation
- ✅ CQL rendering from plans
- ✅ UI modal for preview and confirmation
- ✅ Schema context extraction

Future enhancements:
- ⏳ Query optimization suggestions
- ⏳ Natural language explanations of existing queries
- ⏳ Enhanced context awareness with table statistics
- ⏳ Multi-turn conversation support