#!/bin/bash
set -e

# Build script for macOS application

echo "Building OpenRouter Proxy for macOS..."

# Ensure dependencies are up to date
echo "Updating dependencies..."
go mod tidy

# Build the application
echo "Building application..."
go build -o OpenRouterProxy app.go || {
    echo "Build failed!"
    exit 1
}

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

# Create a placeholder icon
echo "Creating placeholder icon..."
# Save the PNG from app.go to a temporary file
cat > temp_icon.png << 'EOF'
$(xxd -p -c 256 getIcon | xxd -r -p)
EOF

# Create an iconset directory
mkdir -p OpenRouterProxy.iconset
# Copy the PNG to multiple sizes (simplified for placeholder)
cp temp_icon.png OpenRouterProxy.iconset/icon_16x16.png
cp temp_icon.png OpenRouterProxy.iconset/icon_32x32.png
cp temp_icon.png OpenRouterProxy.iconset/icon_128x128.png
cp temp_icon.png OpenRouterProxy.iconset/icon_256x256.png
# Convert to icns
iconutil -c icns OpenRouterProxy.iconset -o "$RESOURCES_DIR/AppIcon.icns"
# Clean up temporary files
rm -rf OpenRouterProxy.iconset temp_icon.png

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

# Create a DMG
echo "Creating DMG..."
if command -v create-dmg &> /dev/null; then
    create-dmg \
        --volname "OpenRouterProxy" \
        --volicon "$RESOURCES_DIR/AppIcon.icns" \
        --window-pos 200 120 \
        --window-size 600 400 \
        --icon-size 100 \
        --icon "OpenRouterProxy.app" 175 190 \
        --hide-extension "OpenRouterProxy.app" \
        --app-drop-link 425 190 \
        "OpenRouterProxy.dmg" \
        "$APP_NAME" || {
            echo "DMG creation failed!"
            exit 1
        }
else
    echo "create-dmg not found. Skipping DMG creation."
    echo "To create a DMG, install create-dmg: brew install create-dmg"
    exit 1
fi

echo "Build complete!"
echo "Application bundle created at: $APP_NAME"
echo "DMG created at: OpenRouterProxy.dmg"

# Clean up temporary binary
rm -f OpenRouterProxy