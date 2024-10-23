#!/bin/sh

set -e

# Function to download a file
download_file() {
    URL=$1
    OUTPUT=$2
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$OUTPUT" "$URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$OUTPUT" "$URL"
    elif command -v fetch >/dev/null 2>&1; then
        fetch -o "$OUTPUT" "$URL"
    else
        echo "Error: No supported download tool found (curl, wget, or fetch)."
        echo "Please install one of these tools and try again."
        exit 1
    fi
}

# Determine OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS" = "darwin" ]; then
    OS="darwin"
elif [ "$OS" = "linux" ]; then
    OS="linux"
else
    echo "Unsupported OS: $OS"
    exit 1
fi

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Create .gommit directory if it doesn't exist
mkdir -p .gommit

# Download gommit
GOMMIT_URL="https://github.com/Moukrea/gommit/releases/latest/download/gommit-$OS-$ARCH"
download_file "$GOMMIT_URL" ".gommit/gommit"

# Make gommit executable
chmod +x .gommit/gommit

# Create temporary files for hook content
cat > /tmp/gommit_hook << 'EOF'
# <<<< Gommit managed block

# Set your custom hooks here

# Gommit commit message validation
./.gommit/gommit $1
exit $?
# >>>> Gommit managed block
EOF

# Handle commit-msg hook
DEST_FILE=".git/hooks/commit-msg"

if [ -f "$DEST_FILE" ]; then
    echo "Existing commit-msg hook found."
    if grep -q "# <<<< Gommit managed block" "$DEST_FILE" && grep -q "# >>>> Gommit managed block" "$DEST_FILE"; then
        # Create a temp file for comparison
        sed -n '/# <<<< Gommit managed block/,/# >>>> Gommit managed block/p' "$DEST_FILE" > /tmp/existing_block
        if diff -q /tmp/gommit_hook /tmp/existing_block >/dev/null 2>&1; then
            echo "Gommit hook is up to date. No changes needed."
        else
            # Create a new temporary file with the content before the managed block
            sed '/# <<<< Gommit managed block/,$d' "$DEST_FILE" > "${DEST_FILE}.tmp"
            
            # Append the new managed block
            cat /tmp/gommit_hook >> "${DEST_FILE}.tmp"
            
            # Append the content after the managed block
            sed -n '/# >>>> Gommit managed block/,$p' "$DEST_FILE" | sed '1d' >> "${DEST_FILE}.tmp"
            
            # Replace the original file
            mv "${DEST_FILE}.tmp" "$DEST_FILE"
            echo "Updated existing Gommit managed block in commit-msg hook."
        fi
    else
        printf "Gommit hook not found. Choose action (append/skip): "
        read -r choice
        case "$choice" in
            append)
                cat /tmp/gommit_hook >> "$DEST_FILE"
                echo "Appended Gommit managed block to existing commit-msg hook."
                ;;
            skip)
                echo "Skipped modifying commit-msg hook."
                ;;
            *)
                echo "Invalid choice. Skipping commit-msg hook modification."
                ;;
        esac
    fi
else
    printf "#!/bin/sh\n" > "$DEST_FILE"
    cat /tmp/gommit_hook >> "$DEST_FILE"
    chmod +x "$DEST_FILE"
    echo "Created new commit-msg hook with Gommit managed block."
fi

# Clean up temporary files
rm -f /tmp/gommit_hook /tmp/existing_block

echo "Gommit has been set up successfully."