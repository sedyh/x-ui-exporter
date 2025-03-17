#!/bin/bash

# Check if script is run as root
if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root (sudo)."
    exit 1
fi

echo "Starting installation of 3x-ui-exporter..."

# Create dedicated system user for running the service
if ! id -u x-ui-exporter > /dev/null 2>&1; then
    echo "Creating system user: x-ui-exporter..."
    useradd -r -s /bin/false x-ui-exporter
    if [ $? -ne 0 ]; then
        echo "Failed to create user. Installation aborted."
        exit 1
    fi
fi

# Determine system architecture
echo "Detecting system architecture..."
ARCH=$(uname -m)
case ${ARCH} in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: ${ARCH}"
        exit 1
        ;;
esac
echo "Architecture detected: ${ARCH}"

# Get latest release tag
echo "Fetching latest release information..."
LATEST_RELEASE=$(curl -s https://api.github.com/repos/hteppl/3x-ui-exporter/releases/latest)
if [ $? -ne 0 ]; then
    echo "Failed to fetch release information. Installation aborted."
    exit 1
fi

VERSION=$(echo "${LATEST_RELEASE}" | grep -Po '"tag_name": "\K.*?(?=")')
echo "Latest version: ${VERSION}"

# Download the appropriate archive
TEMP_DIR=$(mktemp -d)
ARCHIVE_NAME="3x-ui-exporter-${VERSION}-linux-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/hteppl/3x-ui-exporter/releases/download/${VERSION}/${ARCHIVE_NAME}"

echo "Downloading binary from: ${DOWNLOAD_URL}"
curl -L -o "${TEMP_DIR}/${ARCHIVE_NAME}" "${DOWNLOAD_URL}"
if [ $? -ne 0 ]; then
    echo "Failed to download binary. Installation aborted."
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# Extract binary
echo "Extracting binary..."
tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}"
if [ $? -ne 0 ]; then
    echo "Failed to extract binary. Installation aborted."
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# Install binary to /usr/local/bin
echo "Installing binary to /usr/local/bin..."
cp "${TEMP_DIR}/x-ui-exporter" /usr/local/bin/
if [ $? -ne 0 ]; then
    echo "Failed to install binary. Installation aborted."
    rm -rf "${TEMP_DIR}"
    exit 1
fi

# Clean up
rm -rf "${TEMP_DIR}"

# Set permissions
chmod 755 /usr/local/bin/x-ui-exporter

# Create config directory
echo "Creating configuration directory..."
mkdir -p /etc/x-ui-exporter/
if [ $? -ne 0 ]; then
    echo "Failed to create config directory. Installation aborted."
    exit 1
fi

# Download config file
echo "Downloading example config from GitHub..."
curl -s -o /etc/x-ui-exporter/config.yaml https://raw.githubusercontent.com/hteppl/3x-ui-exporter/main/config-example.yaml
if [ $? -ne 0 ]; then
    echo "Failed to download config file. Installation aborted."
    exit 1
fi

# Interactive configuration
echo ""
echo "===== 3X-UI Exporter ====="
echo "Provide your 3X-UI panel details:"
echo ""

# Get Panel URL
while true; do
    read -p "Enter Panel URL (e.g., http://example.com:54321): " PANEL_URL
    # Remove trailing slash if present
    PANEL_URL=${PANEL_URL%/}
    
    if [ -z "$PANEL_URL" ]; then
        echo "Error: Panel URL cannot be empty. Please try again."
    elif [[ ! "$PANEL_URL" =~ ^https?:// ]]; then
        echo "Error: Panel URL must start with http:// or https://. Please try again."
    else
        break
    fi
done

# Get credentials
while true; do
    read -p "Enter Panel Username: " PANEL_USERNAME
    if [ -z "$PANEL_USERNAME" ]; then
        echo "Error: Panel Username cannot be empty. Please try again."
    else
        break
    fi
done

while true; do
    read -s -p "Enter Panel Password: " PANEL_PASSWORD
    echo ""
    if [ -z "$PANEL_PASSWORD" ]; then
        echo "Error: Panel Password cannot be empty. Please try again."
    else
        break
    fi
done

# Validate connection to panel
echo "Validating connection to panel..."
TEMP_RESPONSE=$(mktemp)
CURL_EXIT_CODE=0
LOGIN_RESULT=$(curl -s -w "%{http_code}" -o "$TEMP_RESPONSE" -X POST "${PANEL_URL}/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${PANEL_USERNAME}\",\"password\":\"${PANEL_PASSWORD}\"}" || { CURL_EXIT_CODE=$?; echo "000"; })

if [ $CURL_EXIT_CODE -ne 0 ]; then
    echo "Failed to connect to panel. Network error (curl exit code: $CURL_EXIT_CODE)"
    echo "Please check if the panel URL is correct and the server is reachable."
    read -p "Continue anyway? (y/n): " CONTINUE
    if [ "$CONTINUE" != "y" ] && [ "$CONTINUE" != "Y" ]; then
        rm -f "$TEMP_RESPONSE"
        echo "Installation aborted."
        exit 1
    fi
elif [ "$LOGIN_RESULT" == "200" ] || [ "$LOGIN_RESULT" == "303" ]; then
    echo "✅ Successfully connected to panel!"
elif [ "$LOGIN_RESULT" == "401" ] || [ "$LOGIN_RESULT" == "403" ]; then
    echo "Authentication failed. Invalid username or password."
    read -p "Continue anyway? (y/n): " CONTINUE
    if [ "$CONTINUE" != "y" ] && [ "$CONTINUE" != "Y" ]; then
        rm -f "$TEMP_RESPONSE"
        echo "Installation aborted."
        exit 1
    fi
else
    echo "Failed to connect to panel. HTTP status: ${LOGIN_RESULT}"
    echo "Please verify your panel URL and credentials."
    read -p "Continue anyway? (y/n): " CONTINUE
    if [ "$CONTINUE" != "y" ] && [ "$CONTINUE" != "Y" ]; then
        rm -f "$TEMP_RESPONSE"
        echo "Installation aborted."
        exit 1
    fi
fi
rm -f "$TEMP_RESPONSE"

# Update the config file with user input
echo "Updating configuration file with provided details..."
# Escape special characters in variables for sed
PANEL_URL_ESCAPED=$(echo "$PANEL_URL" | sed 's/[\/&]/\\&/g')
PANEL_USERNAME_ESCAPED=$(echo "$PANEL_USERNAME" | sed 's/[\/&]/\\&/g')
PANEL_PASSWORD_ESCAPED=$(echo "$PANEL_PASSWORD" | sed 's/[\/&]/\\&/g')

sed -i "s|panel-base-url:.*|panel-base-url: \"${PANEL_URL_ESCAPED}\"|" /etc/x-ui-exporter/config.yaml
sed -i "s|panel-username:.*|panel-username: \"${PANEL_USERNAME_ESCAPED}\"|" /etc/x-ui-exporter/config.yaml
sed -i "s|panel-password:.*|panel-password: \"${PANEL_PASSWORD_ESCAPED}\"|" /etc/x-ui-exporter/config.yaml

chmod 644 /etc/x-ui-exporter/config.yaml
chown -R x-ui-exporter:x-ui-exporter /etc/x-ui-exporter

# Create systemd service file
echo "Downloading systemd service file from GitHub..."
curl -s -o /etc/systemd/system/x-ui-exporter.service https://raw.githubusercontent.com/hteppl/3x-ui-exporter/main/x-ui-exporter.service

if [ $? -ne 0 ]; then
    echo "Failed to create service file. Installation aborted."
    exit 1
fi

chmod 644 /etc/systemd/system/x-ui-exporter.service

# Reload systemd to recognize the new service
echo "Reloading systemd daemon..."
systemctl daemon-reload
if [ $? -ne 0 ]; then
    echo "Failed to reload systemd. Installation aborted."
    exit 1
fi

# Enable and start the service
echo "Enabling and starting x-ui-exporter service..."
systemctl enable x-ui-exporter.service
if [ $? -ne 0 ]; then
    echo "Failed to enable service. Installation aborted."
    exit 1
fi

systemctl start x-ui-exporter.service
if [ $? -ne 0 ]; then
    echo "Failed to start service. Installation aborted."
    exit 1
fi

# Check if service is running
if systemctl is-active --quiet x-ui-exporter.service; then
    echo ""
    echo "  3X-UI Exporter has been successfully installed and started!"
    echo "   - Binary location: /usr/local/bin/x-ui-exporter"
    echo "   - Config location: /etc/x-ui-exporter/config.yaml"
    echo "   - Service status: Active"
    echo ""
    echo ""
    echo "You can check the service status with: systemctl status x-ui-exporter.service"
    echo "You can view logs with: journalctl -u x-ui-exporter.service"
    echo "Support the project: https://pay.cloudtips.ru/p/67507843"
else
    echo "Installation completed but the service failed to start."
    echo "Please check the logs for more information: journalctl -u x-ui-exporter.service"
    exit 1
fi
