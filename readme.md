# OpenRouter Proxy for macOS

This repository provides a macOS status bar application that runs a proxy server emulating [Ollama's REST API](https://github.com/ollama/ollama) but forwards requests to [OpenRouter](https://openrouter.ai/). It uses the [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) library under the hood, with minimal code changes to keep the Ollama API calls the same. This allows you to use Ollama-compatible tooling and clients, but run your requests on OpenRouter-managed models.

The original proxy functionality was created by [marknefedov](https://github.com/marknefedov/ollama-openrouter-proxy) and was adapted for use with the [Enchanted project](https://github.com/gluonfield/enchanted/tree/main).

## Description
This macOS application sits in your status bar, allowing you to easily start and stop the proxy server, configure your OpenRouter API key, and manage model filtering. It's perfect for developers who want to use OpenRouter models with tools that support Ollama, such as [Jetbrains AI assistant](https://blog.jetbrains.com/ai/2024/11/jetbrains-ai-assistant-2024-3/#more-control-over-your-chat-experience-choose-between-gemini,-openai,-and-local-models).

## Features
- **Model Filtering**: You can provide a `models-filter` file in the same directory as the proxy. Each line in this file should contain a single model name. The proxy will only show models that match these entries. If the file doesn’t exist or is empty, no filtering is applied.

  **Note**: OpenRouter model names may sometimes include a vendor prefix, for example `deepseek/deepseek-chat-v3-0324:free`. To make sure filtering works correctly, remove the vendor part when adding the name to your `models-filter` file, e.g. `deepseek-chat-v3-0324:free`.

- **Ollama-like API**: The server listens on `11434` and exposes endpoints similar to Ollama (e.g., `/api/chat`, `/api/tags`).
- **Model Listing**: Fetch a list of available models from OpenRouter.
- **Model Details**: Retrieve metadata about a specific model.
- **Streaming Chat**: Forward streaming responses from OpenRouter in a chunked JSON format that is compatible with Ollama’s expectations.

## Usage

1. **Launch the Application**:
   - Run the application from your Applications folder or from the command line.
   - A new icon will appear in your macOS status bar.

2. **Configure API Key**:
   - Click on the status bar icon and select "Configure API Key".
   - Enter your OpenRouter API key when prompted.
   - You can get an API key from [OpenRouter](https://openrouter.ai/).

3. **Start the Proxy Server**:
   - Click on the status bar icon and select "Start Server".
   - The icon will change to indicate that the server is running.

4. **Configure Model Filtering (Optional)**:
   - Click on the status bar icon and select "Edit Model Filter".
   - Add model names to the file, one per line.
   - Save the file and restart the server for changes to take effect.

Once running, the proxy listens on port `11434`. You can make requests to `http://localhost:11434` with your Ollama-compatible tooling.

## Installation

### Option 1: Download the Pre-built Application

1. **Download the Latest Release**:
   - Go to the [Releases](https://github.com/your-username/ollama-openrouter-proxy/releases) page.
   - Download the latest `.dmg` file.

2. **Install the Application**:
   - Open the downloaded `.dmg` file.
   - Drag the application to your Applications folder.

### Option 2: Build from Source

1. **Clone the Repository**:

       git clone https://github.com/your-username/ollama-openrouter-proxy.git
       cd ollama-openrouter-proxy

2. **Install Dependencies**:

       go mod tidy

3. **Build**:

       go build -o OpenRouterProxy

4. **Run the Application**:

       ./OpenRouterProxy
