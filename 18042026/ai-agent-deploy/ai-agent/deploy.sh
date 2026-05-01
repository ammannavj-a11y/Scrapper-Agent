#!/bin/bash
# ╔══════════════════════════════════════════════════════╗
# ║     AI AGENT — QUICK DEPLOY SCRIPT (Mac/Linux)       ║
# ╚══════════════════════════════════════════════════════╝

set -e

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${CYAN}"
echo "╔════════════════════════════════════════════╗"
echo "║            AI AGENT DEPLOYER               ║"
echo "╚════════════════════════════════════════════╝"
echo -e "${NC}"

# Check Node.js
if ! command -v node &>/dev/null; then
  echo -e "${RED}❌  Node.js not found.${NC}"
  echo "    Install from: https://nodejs.org (v16 or higher)"
  exit 1
fi
NODE_VER=$(node -v)
echo -e "${GREEN}✓  Node.js found: $NODE_VER${NC}"

# Check .env
if [ ! -f .env ]; then
  if [ -f .env.example ]; then
    cp .env.example .env
    echo -e "${YELLOW}⚠  Created .env from .env.example${NC}"
    echo -e "${YELLOW}   Please edit .env and add your ANTHROPIC_API_KEY${NC}"
    echo ""
    echo "   Get your key at: https://console.anthropic.com/settings/keys"
    echo ""
    read -p "   Enter your API key now (or press Enter to skip): " key
    if [ -n "$key" ]; then
      sed -i.bak "s/your_api_key_here/$key/" .env
      rm -f .env.bak
      echo -e "${GREEN}✓  API key saved to .env${NC}"
    fi
  fi
fi

# Install dependencies
echo ""
echo -e "${CYAN}Installing dependencies...${NC}"
npm install --production
echo -e "${GREEN}✓  Dependencies installed${NC}"

# Start server
echo ""
echo -e "${GREEN}Starting AI Agent server...${NC}"
echo -e "${CYAN}Open your browser to: http://localhost:3000${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
echo ""

node server.js
