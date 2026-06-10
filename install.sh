#!/usr/bin/env bash
set -e

echo "🔨 Building mdboard..."
go build -o mdboard .

echo "📦 Installing to /usr/local/bin/mdboard..."
sudo mv mdboard /usr/local/bin/mdboard

echo "🔗 Creating 'mdb' alias..."
sudo ln -sf /usr/local/bin/mdboard /usr/local/bin/mdb

echo ""
echo "✅ mdboard installed successfully!"
echo "✨ You can now use the 'mdb' command as a native shortcut."
echo ""
echo "Quick start:"
echo "  mdb config set github_user <your-github-username>"
echo "  mdb new \"My Project\""
echo "  mdb view"
