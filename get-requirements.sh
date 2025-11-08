#!/bin/bash
# get-requirements.sh
# Downloads all Piper TTS requirements if they're not already present
# Linux/macOS version of get-requirements.ps1

set -e

# Parse command line arguments
LANGUAGE="${1:-all}"  # Default to "all" if no argument provided

echo "=== Pejelagarto Translator - Dependency Checker ==="
if [ "$LANGUAGE" != "all" ]; then
    echo "Language: $LANGUAGE (single language mode)"
else
    echo "Language: All languages"
fi
echo ""

# Determine the requirements directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REQUIREMENTS_DIR="${SCRIPT_DIR}/tts/requirements"
PIPER_DIR="${REQUIREMENTS_DIR}/piper"
LANGUAGES_DIR="${PIPER_DIR}/languages"

# Create directories if they don't exist
if [ ! -d "$REQUIREMENTS_DIR" ]; then
    echo "Creating requirements directory..."
    mkdir -p "$REQUIREMENTS_DIR"
fi

if [ ! -d "$PIPER_DIR" ]; then
    echo "Creating piper directory..."
    mkdir -p "$PIPER_DIR"
fi

if [ ! -d "$LANGUAGES_DIR" ]; then
    echo "Creating languages directory..."
    mkdir -p "$LANGUAGES_DIR"
fi

# Function to download a file
download_file() {
    local url="$1"
    local output_path="$2"
    
    echo "  Downloading from: $url"
    echo "  Saving to: $output_path"
    
    if curl -L -f -o "$output_path" "$url"; then
        echo "  ✓ Downloaded successfully"
        return 0
    else
        echo "  ✗ Failed to download"
        return 1
    fi
}

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Detect OS
OS=$(uname -s)
case "$OS" in
    Linux)
        OS_LOWER="linux"
        PIPER_BINARY="piper"
        ;;
    Darwin)
        OS_LOWER="macos"
        PIPER_BINARY="piper"
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

# Check for Piper binary
echo "Checking Piper binary..."
PIPER_EXE="${REQUIREMENTS_DIR}/${PIPER_BINARY}"

if [ ! -f "$PIPER_EXE" ]; then
    echo "Piper binary not found. Downloading..."
    
    TAR_PATH="${REQUIREMENTS_DIR}/piper_${OS_LOWER}_${ARCH}.tar.gz"
    URL="https://github.com/rhasspy/piper/releases/latest/download/piper_${OS_LOWER}_${ARCH}.tar.gz"
    
    if download_file "$URL" "$TAR_PATH"; then
        echo "Extracting Piper binary..."
        tar -xzf "$TAR_PATH" -C "$REQUIREMENTS_DIR"
        rm "$TAR_PATH"
        
        # Make binary executable
        chmod +x "$PIPER_EXE"
        
        if [ -f "$PIPER_EXE" ]; then
            echo "✓ Piper binary extracted successfully"
        else
            echo "✗ Failed to extract Piper binary"
            exit 1
        fi
    else
        echo "✗ Failed to download Piper"
        exit 1
    fi
else
    echo "✓ Piper binary already exists"
fi

# Check for espeak-ng-data
echo ""
echo "Checking espeak-ng-data..."
ESPEAK_DATA="${REQUIREMENTS_DIR}/espeak-ng-data"

if [ ! -d "$ESPEAK_DATA" ]; then
    echo "espeak-ng-data not found. Downloading..."
    
    TAR_PATH="${REQUIREMENTS_DIR}/espeak-ng-data.tar.gz"
    URL="https://github.com/rhasspy/piper/releases/download/v1.2.0/espeak-ng-data.tar.gz"
    
    if download_file "$URL" "$TAR_PATH"; then
        echo "Extracting espeak-ng-data..."
        tar -xzf "$TAR_PATH" -C "$REQUIREMENTS_DIR"
        rm "$TAR_PATH"
        
        if [ -d "$ESPEAK_DATA" ]; then
            echo "✓ espeak-ng-data extracted successfully"
        else
            echo "✗ Failed to extract espeak-ng-data"
            exit 1
        fi
    else
        echo "✗ Failed to download espeak-ng-data"
        exit 1
    fi
else
    echo "✓ espeak-ng-data already exists"
fi

# Language models configuration
declare -A LANGUAGES=(
    ["russian"]="ru_RU-dmitri-medium"
    ["german"]="de_DE-thorsten-medium"
    ["turkish"]="tr_TR-dfki-medium"
    ["portuguese"]="pt_BR-faber-medium"
    ["french"]="fr_FR-siwis-medium"
    ["hindi"]="hi_HI-medium"
    ["romanian"]="ro_RO-mihai-medium"
    ["icelandic"]="is_IS-bui-medium"
    ["swahili"]="sw_CD-lanfrica-medium"
    ["swedish"]="sv_SE-nst-medium"
    ["vietnamese"]="vi_VN-vivos-medium"
    ["czech"]="cs_CZ-jirka-medium"
    ["chinese"]="zh_CN-huayan-medium"
    ["norwegian"]="no_NO-talesyntese-medium"
    ["hungarian"]="hu_HU-anna-medium"
    ["kazakh"]="kk_KZ-iseke-x_low"
)

# Download language models
echo ""
echo "Checking language models..."

# Filter languages based on parameter
if [ "$LANGUAGE" != "all" ]; then
    # Check if language exists in array
    if [ -z "${LANGUAGES[$LANGUAGE]}" ]; then
        echo "Error: Unknown language '$LANGUAGE'"
        echo "Available languages: ${!LANGUAGES[@]}"
        exit 1
    fi
    LANGUAGES_TO_DOWNLOAD=("$LANGUAGE")
else
    LANGUAGES_TO_DOWNLOAD=("${!LANGUAGES[@]}")
fi

for lang_name in "${LANGUAGES_TO_DOWNLOAD[@]}"; do
    model_name="${LANGUAGES[$lang_name]}"
    lang_dir="${LANGUAGES_DIR}/${lang_name}"
    model_file="${lang_dir}/model.onnx"
    config_file="${lang_dir}/model.onnx.json"
    
    if [ -f "$model_file" ] && [ -f "$config_file" ]; then
        echo "✓ ${lang_name} (${model_name}) already exists"
        continue
    fi
    
    echo "Downloading ${lang_name} (${model_name})..."
    
    # Create language directory
    mkdir -p "$lang_dir"
    
    # Download model.onnx
    model_url="https://huggingface.co/rhasspy/piper-voices/resolve/main/${model_name//-/_}.onnx"
    if ! download_file "$model_url" "$model_file"; then
        echo "✗ Failed to download model for ${lang_name}"
        rm -rf "$lang_dir"
        continue
    fi
    
    # Download model.onnx.json
    config_url="https://huggingface.co/rhasspy/piper-voices/resolve/main/${model_name//-/_}.onnx.json"
    if ! download_file "$config_url" "$config_file"; then
        echo "✗ Failed to download config for ${lang_name}"
        rm -rf "$lang_dir"
        continue
    fi
    
    echo "✓ ${lang_name} downloaded successfully"
done

echo ""
echo "=== Dependency check complete ==="
echo "All requirements are ready at: ${REQUIREMENTS_DIR}"
