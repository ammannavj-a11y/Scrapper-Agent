# ────────────────────────────────────────────────────────────────────────────
# Stage 1 — Builder
# ────────────────────────────────────────────────────────────────────────────
FROM python:3.11-slim AS builder

WORKDIR /build

# System deps for compilation
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc g++ libpq-dev curl \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies into /build/venv
COPY requirements-runtime.txt .
RUN python -m venv /build/venv \
    && /build/venv/bin/pip install --upgrade pip setuptools wheel \
    && /build/venv/bin/pip install --no-cache-dir -r requirements-runtime.txt

# Download spaCy model
RUN /build/venv/bin/python -m spacy download en_core_web_sm || true

# ────────────────────────────────────────────────────────────────────────────
# Stage 2 — Runtime (minimal attack surface)
# ────────────────────────────────────────────────────────────────────────────
FROM python:3.11-slim AS runtime

# AppSec: run as non-root user
RUN groupadd -r appuser && useradd -r -g appuser -d /app -s /sbin/nologin appuser

# Only runtime libs
RUN apt-get update && apt-get install -y --no-install-recommends \
    libpq5 curl \
    && rm -rf /var/lib/apt/lists/* \
    # Remove SUID/SGID binaries
    && find / -perm /6000 -type f -exec chmod a-s {} \; 2>/dev/null || true

WORKDIR /app

# Copy venv from builder
COPY --from=builder /build/venv /venv

# Copy application source
COPY --chown=appuser:appuser . /app

# Create data directory
RUN mkdir -p /app/data && chown appuser:appuser /app/data

# Environment
ENV PATH="/venv/bin:$PATH" \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONFAULTHANDLER=1

# AppSec: read-only filesystem where possible
# Switch to non-root user
USER appuser

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

EXPOSE 8000

# Start with Uvicorn (use gunicorn in production via CMD override in k8s)
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000", \
     "--workers", "2", "--loop", "uvloop", "--http", "httptools", \
     "--access-log", "--log-level", "info"]
