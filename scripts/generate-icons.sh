#!/bin/bash

# MQ Toolkit 图标生成脚本
# 从 logo.svg 生成各种尺寸和格式的图标文件

set -e

# 检查是否安装了 ImageMagick 或 rsvg-convert
if ! command -v convert &> /dev/null && ! command -v rsvg-convert &> /dev/null; then
    echo "错误: 需要安装 ImageMagick 或 librsvg"
    echo "macOS: brew install imagemagick librsvg"
    echo "Ubuntu: sudo apt-get install imagemagick librsvg2-bin"
    exit 1
fi

# 创建临时目录和构建目录
TEMP_DIR="temp_icons"
mkdir -p "$TEMP_DIR"
mkdir -p "build"

# 源文件
SOURCE_SVG="logo.svg"

if [ ! -f "$SOURCE_SVG" ]; then
    echo "错误: 找不到 $SOURCE_SVG 文件"
    exit 1
fi

echo "🎨 开始生成图标文件..."

# 生成函数
generate_png() {
    local size=$1
    local output=$2
    
    if command -v rsvg-convert &> /dev/null; then
        rsvg-convert -w $size -h $size "$SOURCE_SVG" -o "$output"
    else
        convert -background transparent "$SOURCE_SVG" -resize ${size}x${size} "$output"
    fi
    echo "✅ 生成: $output (${size}x${size})"
}

# 生成不同尺寸的 PNG 文件
echo "📱 生成 PNG 图标..."
generate_png 16 "$TEMP_DIR/icon-16.png"
generate_png 32 "$TEMP_DIR/icon-32.png"
generate_png 48 "$TEMP_DIR/icon-48.png"
generate_png 64 "$TEMP_DIR/icon-64.png"
generate_png 128 "$TEMP_DIR/icon-128.png"
generate_png 256 "$TEMP_DIR/icon-256.png"
generate_png 512 "$TEMP_DIR/icon-512.png"
generate_png 1024 "$TEMP_DIR/icon-1024.png"

# 生成 macOS 应用图标 (appicon.png)
echo "🍎 生成 macOS 应用图标..."
generate_png 512 "build/appicon.png"

# 生成 Windows ICO 文件
echo "🪟 生成 Windows ICO 图标..."
mkdir -p "build/windows"
if command -v magick &> /dev/null; then
    magick "$TEMP_DIR/icon-16.png" "$TEMP_DIR/icon-32.png" "$TEMP_DIR/icon-48.png" "$TEMP_DIR/icon-256.png" "build/windows/icon.ico"
    echo "✅ 生成: build/windows/icon.ico"
elif command -v convert &> /dev/null; then
    convert "$TEMP_DIR/icon-16.png" "$TEMP_DIR/icon-32.png" "$TEMP_DIR/icon-48.png" "$TEMP_DIR/icon-256.png" "build/windows/icon.ico"
    echo "✅ 生成: build/windows/icon.ico"
else
    echo "⚠️  警告: 无法生成 ICO 文件，需要 ImageMagick"
fi

# 生成 Linux 图标
echo "🐧 生成 Linux 图标..."
mkdir -p "build/linux"
generate_png 512 "build/linux/icon.png"

# 复制 SVG 到 frontend 用于网页显示
echo "🌐 复制 SVG 到前端..."
cp "$SOURCE_SVG" "frontend/src/assets/logo.svg" 2>/dev/null || echo "⚠️  前端 assets 目录不存在，跳过"

# 生成 favicon
echo "🌐 生成 Favicon..."
generate_png 32 "frontend/public/favicon.png" 2>/dev/null || echo "⚠️  前端 public 目录不存在，跳过"

# 清理临时文件
echo "🧹 清理临时文件..."
rm -rf "$TEMP_DIR"

echo ""
echo "🎉 图标生成完成！"
echo ""
echo "生成的文件:"
echo "  📁 build/appicon.png          - macOS 应用图标"
echo "  📁 build/windows/icon.ico     - Windows 应用图标"
echo "  📁 build/linux/icon.png       - Linux 应用图标"
echo "  📁 frontend/src/assets/logo.svg - 前端 Logo"
echo "  📁 frontend/public/favicon.png  - 网页 Favicon"
echo ""
echo "💡 提示: 运行 'wails build' 来使用新图标构建应用"
