/**
 * ╔══════════════════════════════════════════════════╗
 * ║          AI AGENT — USER CONFIGURATION           ║
 * ╚══════════════════════════════════════════════════╝
 *
 * STEP 1: Get your Anthropic API key from:
 *         https://console.anthropic.com/settings/keys
 *
 * STEP 2: Get your OpenAI API key from:
 *         https://platform.openai.com/api-keys
 *
 * STEP 3: Replace YOUR_API_KEY_HERE with your Anthropic key
 *         Replace YOUR_OPENAI_API_KEY_HERE with your OpenAI key
 *
 * STEP 4: Choose provider: "anthropic" or "openai"
 *
 * STEP 5: Save this file and open index.html
 */

window.AI_CONFIG = {

  // ── Provider: "anthropic" or "openai" ────────────────────────────────
  provider: "anthropic",

  // ── Anthropic settings ───────────────────────────────────────────────
  anthropic: {
    apiKey: "YOUR_API_KEY_HERE",
    model: "claude-sonnet-4-20250514",
    apiEndpoint: "https://api.anthropic.com/v1/messages",
  },
  window.AI_CONFIG = {
    provider: "anthropic",
    model: "claude-sonnet-4-20250514",
    maxTokens: 1500,
    defaultAgent: "J.A.R.V.I.S",
    defaultLanguage: "en",
    defaultVoiceGender: "male",
    autoSpeak: true,
    apiEndpoint: "/api/chat"  // Use server proxy, NOT direct API
  };
  // ── OpenAI settings ──────────────────────────────────────────────────
  openai: {
    apiKey: "YOUR_OPENAI_API_KEY_HERE",
    model: "gpt-4",
    apiEndpoint: "https://api.openai.com/v1/chat/completions",
  },

  // ── Max response length (tokens). Increase for longer answers ───────────
  maxTokens: 4096,

  // ── Default agent name on startup ───────────────────────────────────────
  defaultAgent: "J.A.R.V.I.S",

  // ── Default language: "en" | "hi" | "te" | "nl" ────────────────────────
  defaultLanguage: "en",

  // ── Default voice gender: "male" | "female" ─────────────────────────────
  defaultVoiceGender: "male",

  // ── If true, agent will speak every response automatically ───────────────
  autoSpeak: true,

};
// ── Clean config export ──────────────────────────────────────────────────
// Remove duplicate declarations above. Keep only ONE config object.
window.AI_CONFIG = {

  // ── REQUIRED: Your Anthropic API key ────────────────────────────────────
  apiKey: "YOUR_API_KEY_HERE",

  // ── AI Model (do not change unless you know what you're doing) ──────────
  model: "claude-sonnet-4-20250514",

  // ── Max response length (tokens). Increase for longer answers ───────────
  maxTokens: 4096,

  // ── Default agent name on startup ───────────────────────────────────────
  defaultAgent: "J.A.R.V.I.S",

  // ── Default language: "en" | "hi" | "te" | "nl" ────────────────────────
  defaultLanguage: "en",

  // ── Default voice gender: "male" | "female" ─────────────────────────────
  defaultVoiceGender: "male",

  // ── If true, agent will speak every response automatically ───────────────
  autoSpeak: true,

  // ── API endpoint — change if using the server proxy (server.js) ─────────
  // For local file:// mode:   "https://api.anthropic.com/v1/messages"
  // For server proxy mode:    "/api/chat"
  apiEndpoint: "https://api.anthropic.com/v1/messages",

};

window.AI_CONFIG = {

  // ── Provider: "anthropic" or "openai" ────────────────────────────────
  provider: "anthropic",

  // ── Anthropic settings ───────────────────────────────────────────────
  anthropic: {
    apiKey: "YOUR_API_KEY_HERE",
    model: "claude-sonnet-4-20250514",
    apiEndpoint: "https://api.anthropic.com/v1/messages",
  },

  // ── OpenAI settings ──────────────────────────────────────────────────
  openai: {
    apiKey: "YOUR_OPENAI_API_KEY_HERE",
    model: "gpt-4",
    apiEndpoint: "https://api.openai.com/v1/chat/completions",
  },

  // ── Max response length (tokens). Increase for longer answers ───────────
  maxTokens: 1500,

  // ── Default agent name on startup ───────────────────────────────────────
  defaultAgent: "J.A.R.V.I.S",

  // ── Default language: "en" | "hi" | "te" | "nl" ────────────────────────
  defaultLanguage: "en",

  // ── Default voice gender: "male" | "female" ─────────────────────────────
  defaultVoiceGender: "male",

  // ── If true, agent will speak every response automatically ───────────────
  autoSpeak: true,

};

window.AI_CONFIG = {

  // ── REQUIRED: Your Anthropic API key ────────────────────────────────────
  apiKey: "YOUR_API_KEY_HERE",

  // ── AI Model (do not change unless you know what you're doing) ──────────
  model: "claude-sonnet-4-20250514",

  // ── Max response length (tokens). Increase for longer answers ───────────
  maxTokens: 1500,

  // ── Default agent name on startup ───────────────────────────────────────
  defaultAgent: "J.A.R.V.I.S",

  // ── Default language: "en" | "hi" | "te" | "nl" ────────────────────────
  defaultLanguage: "en",

  // ── Default voice gender: "male" | "female" ─────────────────────────────
  defaultVoiceGender: "male",

  // ── If true, agent will speak every response automatically ───────────────
  autoSpeak: true,

  // ── API endpoint — change if using the server proxy (server.js) ─────────
  // For local file:// mode:   "https://api.anthropic.com/v1/messages"
  // For server proxy mode:    "/api/chat"
  apiEndpoint: "https://api.anthropic.com/v1/messages",

};
