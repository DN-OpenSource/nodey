#!/bin/bash
echo "🚀 Preparing Release..."

# 1. Check for token
if [ -z "$GITHUB_TOKEN" ]; then
    echo "❌ Error: GITHUB_TOKEN is not set."
    echo "Please export GITHUB_TOKEN='your_token' and try again."
    exit 1
fi

# 2. Push code
echo "📦 Pushing code..."
git push origin main

# 3. Tag
# 3. Tag
echo "🏷️  Tagging v1.0.6..."
git tag -a v1.0.6 -m "Release v1.0.6"
git push origin v1.0.6

# 4. Release
echo "🚀 Running GoReleaser..."
goreleaser release --clean

echo "✅ Done! Users can now run: brew tap DN-OpenSource/tap && brew install nodey"
