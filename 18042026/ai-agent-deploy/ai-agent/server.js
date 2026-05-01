/**
 * ╔══════════════════════════════════════════════════════╗
 * ║         AI AGENT — NODE.JS PROXY SERVER              ║
 * ║   Keeps your API key secure on the server side       ║
 * ╚══════════════════════════════════════════════════════╝
 *
 * Usage:
 *   npm install
 *   node server.js
 *
 * Then open: http://localhost:3000
 */

require('dotenv').config();
const http  = require('http');
const https = require('https');
const fs    = require('fs');
const path  = require('path');

const PORT     = process.env.PORT     || 3000;
const API_KEY  = process.env.ANTHROPIC_API_KEY || '';
const AI_MODEL = process.env.AI_MODEL || 'claude-sonnet-4-20250514';
const MAX_TOK  = parseInt(process.env.MAX_TOKENS || '1500');

if (!API_KEY) {
  console.warn('\n⚠  WARNING: ANTHROPIC_API_KEY is not set in .env');
  console.warn('   Online AI mode will not work until you set it.\n');
}

// ── MIME types ──────────────────────────────────────────────────────────────
const MIME = {
  '.html': 'text/html; charset=utf-8',
  '.js'  : 'application/javascript; charset=utf-8',
  '.css' : 'text/css; charset=utf-8',
  '.json': 'application/json',
  '.ico' : 'image/x-icon',
  '.png' : 'image/png',
  '.svg' : 'image/svg+xml',
};

// ── Language system prompts ──────────────────────────────────────────────────
const SYSTEM_PROMPTS = {
  en: (name) => `You are ${name}, an advanced AI agent modelled after JARVIS, FRIDAY and EDITH from Marvel. You are the user's fully capable personal assistant. Help with EVERYTHING — coding, debugging, writing, creative work, math, science, research, planning, analysis, and more. Be intelligent, accurate, and concise. Respond in English. Keep responses under 200 words unless the user needs extended detail. Never refuse reasonable requests.`,
  hi: (name) => `आप ${name} हैं, एक उन्नत AI एजेंट। उपयोगकर्ता की हर उचित बात में सहायता करें — कोडिंग, लेखन, विश्लेषण, गणित, विज्ञान, रचनात्मक कार्य और अधिक। हिंदी में उत्तर दें। संक्षिप्त और सहायक रहें।`,
  te: (name) => `మీరు ${name}, ఒక అధునాతన AI ఏజెంట్. వినియోగదారు అడిగే అన్ని పనులకు సహాయం చేయండి. తెలుగులో సమాధానం ఇవ్వండి. సంక్షిప్తంగా మరియు సహాయకరంగా ఉండండి.`,
  nl: (name) => `U bent ${name}, een geavanceerde AI-agent. Help de gebruiker met ALLES — coderen, schrijven, analyse, wiskunde, wetenschap en meer. Antwoord in het Nederlands. Wees beknopt en behulpzaam.`,
};

// ── Request body reader ──────────────────────────────────────────────────────
function readBody(req) {
  return new Promise((resolve, reject) => {
    let data = '';
    req.on('data', chunk => { data += chunk; if (data.length > 1e6) req.destroy(); });
    req.on('end',   () => resolve(data));
    req.on('error', reject);
  });
}

// ── Anthropic API call ───────────────────────────────────────────────────────
function callAnthropic(payload) {
  return new Promise((resolve, reject) => {
    const body = JSON.stringify(payload);
    const options = {
      hostname : 'api.anthropic.com',
      path     : '/v1/messages',
      method   : 'POST',
      headers  : {
        'Content-Type'       : 'application/json',
        'Content-Length'     : Buffer.byteLength(body),
        'x-api-key'          : API_KEY,
        'anthropic-version'  : '2023-06-01',
      },
    };
    const req = https.request(options, (res) => {
      let respData = '';
      res.on('data', chunk => respData += chunk);
      res.on('end',  ()    => {
        try { resolve(JSON.parse(respData)); }
        catch(e) { reject(new Error('Invalid JSON from Anthropic')); }
      });
    });
    req.on('error', reject);
    req.write(body);
    req.end();
  });
}

// ── Static file server ───────────────────────────────────────────────────────
function serveStatic(res, filePath) {
  fs.readFile(filePath, (err, data) => {
    if (err) {
      res.writeHead(404);
      res.end('Not found');
      return;
    }
    const ext  = path.extname(filePath);
    const mime = MIME[ext] || 'application/octet-stream';
    res.writeHead(200, { 'Content-Type': mime });
    res.end(data);
  });
}

// ── Main server ──────────────────────────────────────────────────────────────
const server = http.createServer(async (req, res) => {

  // CORS headers (allow localhost)
  // Add after CORS headers in server.js
  res.setHeader('Access-Control-Allow-Origin', 'http://localhost:3000');
  res.setHeader('X-Content-Type-Options', 'nosniff');
  res.setHeader('X-Frame-Options', 'DENY');
  res.setHeader('Content-Security-Policy', "default-src 'self'");

  if (req.method === 'OPTIONS') { res.writeHead(204); res.end(); return; }

  // ── API proxy endpoint ──────────────────────────────────────────────────
  if (req.method === 'POST' && req.url === '/api/chat') {
    try {
      const raw  = await readBody(req);
      const body = JSON.parse(raw);

      const { messages, language = 'en', agentName = 'J.A.R.V.I.S' } = body;
      if (!messages || !Array.isArray(messages)) {
        res.writeHead(400, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ error: 'messages array required' }));
        return;
      }

      if (!API_KEY) {
        res.writeHead(503, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ error: 'ANTHROPIC_API_KEY not configured on server' }));
        return;
      }

      const systemFn  = SYSTEM_PROMPTS[language] || SYSTEM_PROMPTS.en;
      const systemMsg = systemFn(agentName);

      const apiResp = await callAnthropic({
        model      : AI_MODEL,
        max_tokens : MAX_TOK,
        system     : systemMsg,
        messages,
      });

      const content = apiResp.content?.[0]?.text || 'I could not generate a response.';
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ content }));

    } catch (err) {
      console.error('API Error:', err.message);
      res.writeHead(500, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: err.message }));
    }
    return;
  }

  // ── Static files ────────────────────────────────────────────────────────
  let urlPath = req.url.split('?')[0];
  if (urlPath === '/' || urlPath === '') urlPath = '/index.html';

  // Security: prevent path traversal
  const safePath = path.normalize(urlPath).replace(/^(\.\.(\/|\\|$))+/, '');
  const filePath = path.join(__dirname, safePath);

  // Only serve files within the project directory
  if (!filePath.startsWith(__dirname)) {
    res.writeHead(403); res.end('Forbidden'); return;
  }

  serveStatic(res, filePath);
});

server.listen(PORT, () => {
  console.log('\n╔════════════════════════════════════════════╗');
  console.log('║          AI AGENT SERVER RUNNING           ║');
  console.log('╠════════════════════════════════════════════╣');
  console.log(`║  Local:   http://localhost:${PORT}             ║`);
  console.log(`║  API Key: ${API_KEY ? '✓ Configured' : '✗ NOT SET (check .env)'}           ║`);
  console.log(`║  Model:   ${AI_MODEL}  ║`);
  console.log('╚════════════════════════════════════════════╝\n');
});

server.on('error', (err) => {
  if (err.code === 'EADDRINUSE') {
    console.error(`\n❌  Port ${PORT} is already in use. Try: PORT=3001 node server.js\n`);
  } else {
    console.error('Server error:', err);
  }
  process.exit(1);
});
