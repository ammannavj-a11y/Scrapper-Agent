# ── AI AGENT DOCKERFILE ─────────────────────────────────────────────────────
FROM node:20-alpine

LABEL description="AI Agent — JARVIS/FRIDAY/EDITH style multilingual AI"

WORKDIR /app

# Copy package files
COPY package.json ./

# Install only production dependencies (just dotenv)
RUN npm install --production

# Copy all application files
COPY . .

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s \
  CMD wget -qO- http://localhost:3000 || exit 1

# Start server
CMD ["node", "server.js"]
