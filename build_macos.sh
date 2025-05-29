#!/bin/bash
set -e

# Build script for macOS application

echo "Building OpenRouter Proxy for macOS..."

# Ensure dependencies are up to date
echo "Updating dependencies..."
go mod tidy

# Build the application
echo "Building application..."
go build -o OpenRouterProxy

# Create application bundle structure
echo "Creating application bundle..."
APP_NAME="OpenRouterProxy.app"
CONTENTS_DIR="$APP_NAME/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"

# Remove existing bundle if it exists
rm -rf "$APP_NAME"

# Create directory structure
mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR"

# Copy binary to MacOS directory
cp OpenRouterProxy "$MACOS_DIR/"

# Create Info.plist
cat > "$CONTENTS_DIR/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>OpenRouterProxy</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.openrouterproxy.app</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>OpenRouterProxy</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2023. All rights reserved.</string>
</dict>
</plist>
EOF

# Create a simple icon (placeholder)
# In a real application, you would use a proper icon file
echo "Note: Using placeholder icon. Replace with a proper icon for production."

# Create a DMG (optional)
echo "Creating DMG..."
if command -v create-dmg &> /dev/null; then
    create-dmg \
        --volname "OpenRouterProxy" \
        --volicon "OpenRouterProxy.app/Contents/Resources/AppIcon.icns" \
        --window-pos 200 120 \
        --window-size 600 400 \
        --icon-size 100 \
        --icon "OpenRouterProxy.app" 175 190 \
        --hide-extension "OpenRouterProxy.app" \
        --app-drop-link 425 190 \
        "OpenRouterProxy.dmg" \
        "OpenRouterProxy.app"
else
    echo "create-dmg not found. Skipping DMG creation."
    echo "To create a DMG, install create-dmg: brew install create-dmg"
fi

echo "Build complete!"
echo "Application bundle created at: $APP_NAME"
if [ -f "OpenRouterProxy.dmg" ]; then
    echo "DMG created at: OpenRouterProxy.dmg"
fi