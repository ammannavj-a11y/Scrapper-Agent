#!/bin/bash
set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  GUVID v3.0 — Generating Deployable ZIP"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="$(dirname "$PROJECT_DIR")"
ZIP_NAME="guvid-platform-v3.0.zip"

cd "$PROJECT_DIR"

# Clean up
rm -f "$OUTPUT_DIR/$ZIP_NAME"
find . -name "*.sum" -delete 2>/dev/null || true

zip -r "$OUTPUT_DIR/$ZIP_NAME" . \
  --exclude "*.git*" \
  --exclude "*/tmp/*" \
  --exclude "*__pycache__*" \
  --exclude "*.DS_Store"

SIZE=$(du -sh "$OUTPUT_DIR/$ZIP_NAME" | cut -f1)
FILES=$(unzip -l "$OUTPUT_DIR/$ZIP_NAME" | tail -1 | awk '{print $2}')

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ ENTERPRISE ZIP READY"
echo "   File:  $OUTPUT_DIR/$ZIP_NAME"
echo "   Size:  $SIZE"
echo "   Files: $FILES"
echo ""
echo "Deploy with one command:"
echo ""
echo "   unzip guvid-platform-v3.0.zip -d guvid-platform"
echo "   cd guvid-platform"
echo "   cp .env.example .env"
echo "   make quickstart"
echo ""
echo "   Open http://localhost:3000"
echo ""
echo "Demo Logins (password: Admin@123):"
echo "   HR:          hr@google.com         → google-hr"
echo "   Institution: registrar@iitd.ac.in  → iit-delhi"
echo "   Regulatory:  regulator@dpdp.gov.in → india-regulator"
echo "   Fraud L1:    analyst@fraudmonitoring.in → fraud-monitoring"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
