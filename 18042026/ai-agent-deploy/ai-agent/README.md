# 🤖 AI AGENT — J.A.R.V.I.S / F.R.I.D.A.Y / E.D.I.T.H

> An advanced personal AI agent with a sci-fi HUD interface.  
> Multilingual • Voice Input/Output • Online + Offline Modes • Customizable

---

## ✨ FEATURES

| Feature | Detail |
|---------|--------|
| **AI Engine** | Claude AI (Anthropic) — online mode |
| **Offline Mode** | Built-in responses work WITHOUT internet |
| **Languages** | 🇺🇸 English • 🇮🇳 Hindi • 🇮🇳 Telugu • 🇳🇱 Dutch |
| **Voice Input** | Microphone / Speech-to-text (Chrome/Edge) |
| **Voice Output** | Text-to-speech in all 4 languages |
| **Agent Names** | JARVIS, FRIDAY, EDITH, ARIA, ORION, NOVA + Custom |
| **Voice Gender** | Male & Female voices |
| **Interface** | Animated HUD with circular waveform visualizer |
| **Deployment** | Browser file / Node.js server / Docker |

---

## 🚀 QUICK START — 3 Methods

---

### METHOD 1: Direct Browser (Simplest — No Install Required)

Best for personal local use.

**Step 1:** Open `config.js` in a text editor.

**Step 2:** Replace `YOUR_API_KEY_HERE` with your Anthropic API key:
```js
apiKey: "sk-ant-api03-...",
apiEndpoint: "https://api.anthropic.com/v1/messages",
```

**Step 3:** Get a free API key at: https://console.anthropic.com/settings/keys

**Step 4:** Double-click `index.html` to open in Chrome or Edge.

> ⚠️ **Note:** Voice recognition requires Chrome or Edge browser.  
> ⚠️ **Note:** API calls from `file://` may be blocked by some browsers. Use Method 2 (Node.js) for reliable local use.

---

### METHOD 2: Node.js Server (Recommended — Most Reliable)

Runs a local web server. API key stays secure on the server.

#### Prerequisites
- Node.js v16+ — Download: https://nodejs.org

#### Windows
```bat
deploy.bat
```
*(Double-click `deploy.bat` — it does everything automatically)*

#### Mac / Linux
```bash
chmod +x deploy.sh
./deploy.sh
```

#### Manual steps (any OS)
```bash
# 1. Copy environment file
cp .env.example .env

# 2. Edit .env and add your API key
#    ANTHROPIC_API_KEY=sk-ant-api03-...

# 3. Install dependencies (only 'dotenv' package)
npm install

# 4. Start the server
node server.js

# 5. Open browser
# http://localhost:3000
```

---

### METHOD 3: Docker (Production / Server Deployment)

#### Prerequisites
- Docker Desktop: https://www.docker.com/products/docker-desktop

```bash
# 1. Set up environment
cp .env.example .env
# Edit .env: ANTHROPIC_API_KEY=sk-ant-api03-...

# 2. Build and run
docker-compose up -d

# 3. Open browser
# http://localhost:3000

# 4. Stop when done
docker-compose down
```

Manual Docker (without docker-compose):
```bash
docker build -t ai-agent .
docker run -d -p 3000:3000 -e ANTHROPIC_API_KEY=sk-ant-... ai-agent
```

---

### METHOD 4: Deploy to Cloud (Public URL)

#### Vercel (Free)
```bash
npm install -g vercel
vercel deploy
# Set env var ANTHROPIC_API_KEY in Vercel dashboard
```

#### Render.com (Free)
1. Push this folder to a GitHub repo
2. Create new "Web Service" on render.com
3. Set Build Command: `npm install`
4. Set Start Command: `node server.js`
5. Add env var: `ANTHROPIC_API_KEY`

#### Railway.app
```bash
npm install -g @railway/cli
railway login
railway init
railway up
# Add ANTHROPIC_API_KEY in Railway dashboard
```

---

## 🔑 GET YOUR API KEY

1. Go to: **https://console.anthropic.com/settings/keys**
2. Sign up / log in (free account)
3. Click **"Create Key"**
4. Copy the key (starts with `sk-ant-api03-...`)
5. Add it to `config.js` (Method 1) or `.env` (Methods 2-4)

---

## 🎮 HOW TO USE

### Voice Commands
1. Click the **🎤 microphone button** (or press it)
2. Speak your command
3. The agent will respond in text AND voice

### Text Commands
1. Type in the input box at the bottom right
2. Press **Enter** or click **▶**

### Language Switching
- Click **EN / हि / తె / NL** buttons in the top bar
- Interface AND voice switch immediately

### Change Agent Name
- Click a **preset** in the left panel (JARVIS, FRIDAY, EDITH, etc.)
- Or type any name in **CUSTOM AGENT NAME** field and click SET
- Or click the agent name in the top-left to toggle the panel

### Voice Gender
- Click **♂ MALE** or **♀ FEMALE** in the left panel
- System will pick the best matching voice for your browser

### Stop Speaking
- Click **🔇** button that appears during speech

---

## 💬 SAMPLE COMMANDS (ALL LANGUAGES)

### English
- "Hello JARVIS, what can you do?"
- "Write me a Python script to sort a list"
- "Explain quantum computing in simple terms"
- "What time is it?" *(works offline)*
- "Tell me a joke" *(works offline)*
- "Calculate 245 times 678" *(works offline)*

### Hindi (हिंदी)
- "नमस्ते, आप क्या कर सकते हैं?"
- "मुझे Python में एक script लिखकर दो"
- "समय क्या है?" *(ऑफलाइन काम करता है)*

### Telugu (తెలుగు)
- "నమస్కారం, మీరు ఏమి చేయగలరు?"
- "నాకు ఒక Python script రాయండి"
- "సమయం ఎంత?" *(ఆఫ్‌లైన్‌లో పని చేస్తుంది)*

### Dutch (Nederlands)
- "Hallo, wat kun jij allemaal doen?"
- "Schrijf een Python script voor mij"
- "Hoe laat is het?" *(werkt offline)*

---

## 🔌 OFFLINE MODE

The agent works WITHOUT internet for:
- ✅ Current time and date (in your language)
- ✅ Basic math calculations
- ✅ Jokes (in all 4 languages)
- ✅ Greetings and identity
- ✅ Help and capabilities info
- ✅ Version information

For full AI intelligence (coding help, research, writing, etc.), internet is required.

The agent **automatically detects** when you go offline/online.

---

## ⚙️ CONFIGURATION

### config.js options
```js
window.AI_CONFIG = {
  apiKey: "YOUR_KEY",                          // Anthropic API key
  model: "claude-sonnet-4-20250514",           // AI model
  maxTokens: 1500,                             // Response length
  defaultAgent: "J.A.R.V.I.S",               // Starting agent name
  defaultLanguage: "en",                       // en | hi | te | nl
  defaultVoiceGender: "male",                  // male | female
  autoSpeak: true,                             // Auto voice responses
  apiEndpoint: "https://api.anthropic.com/v1/messages", // or "/api/chat" for server mode
};
```

### For Node.js server mode, add to config.js:
```js
apiEndpoint: "/api/chat",  // Use the local proxy instead of direct API
apiKey: "",                // Not needed — key is in .env on server
```

---

## 🛠 TROUBLESHOOTING

| Problem | Solution |
|---------|----------|
| Voice input not working | Use Chrome or Edge browser (Firefox not supported) |
| No voice output | Check browser audio permissions; try a different voice gender |
| API not responding | Verify your API key in config.js or .env |
| "Offline mode" even with internet | Check API key is correct; check browser console for errors |
| Port 3000 in use | Run: `PORT=3001 node server.js` |
| Telugu/Hindi voice sounds wrong | Your OS may not have that language pack; try English voice |

---

## 📁 FILE STRUCTURE

```
ai-agent/
├── index.html          ← Main app (open this in browser)
├── config.js           ← Your API key & settings go here
├── server.js           ← Node.js proxy server
├── package.json        ← Node.js project file
├── .env.example        ← Copy to .env for server mode
├── Dockerfile          ← Docker container definition
├── docker-compose.yml  ← Docker Compose config
├── deploy.sh           ← Mac/Linux quick start
├── deploy.bat          ← Windows quick start
└── README.md           ← This file
```

---

## 🔒 SECURITY NOTES

- **Method 1 (Direct):** API key visible in `config.js` — OK for personal local use only
- **Method 2-4 (Server):** API key in `.env` — safe, never sent to browser
- Never commit `.env` to git (it's in `.gitignore`)
- The `index.html` API key field should be empty when using server mode

---

## 📜 LICENSE

MIT License — free for personal and commercial use.

---

*Built with ❤️ | Claude AI by Anthropic | Supports: English, Hindi, Telugu, Dutch*
