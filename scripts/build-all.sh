#!/bin/bash

# MQ Toolkit 多平台构建脚本
# 用于本地测试多平台构建

set -e

echo "🚀 开始多平台构建 MQ Toolkit..."

# 支持的平台
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "windows/amd64"
)

# 清理之前的构建
echo "🧹 清理之前的构建..."
rm -rf build/bin/*
rm -rf dist/

# 创建分发目录
mkdir -p dist

# 生成图标
echo "🎨 生成应用图标..."
if [ -f "scripts/generate-icons.sh" ]; then
    chmod +x scripts/generate-icons.sh
    ./scripts/generate-icons.sh
else
    echo "⚠️  图标生成脚本不存在，跳过"
fi

# 构建每个平台
for platform in "${PLATFORMS[@]}"; do
    echo ""
    echo "🔨 构建平台: $platform"
    
    # 构建应用
    wails build -platform "$platform"
    
    # 获取平台名称
    platform_name=$(echo "$platform" | tr '/' '-')
    
    # 处理构建结果
    if [ -d "build/bin" ]; then
        cd build/bin
        
        case "$platform" in
            "darwin/"*)
                # macOS: 修复权限并打包为 zip
                echo "🍎 处理 macOS 应用..."
                find . -name "*.app" -exec xattr -cr {} \; 2>/dev/null || true
                find . -name "*.app" -exec codesign --force --deep --sign - {} \; 2>/dev/null || true
                zip -r "../../dist/MQToolkit-$platform_name.zip" *.app
                ;;
            "windows/"*)
                # Windows: 打包为 zip
                echo "🪟 处理 Windows 应用..."
                if command -v 7z &> /dev/null; then
                    7z a "../../dist/MQToolkit-$platform_name.zip" *
                else
                    zip -r "../../dist/MQToolkit-$platform_name.zip" *
                fi
                ;;
            "linux/"*)
                # Linux: 打包为 tar.gz
                echo "🐧 处理 Linux 应用..."
                tar -czf "../../dist/MQToolkit-$platform_name.tar.gz" *
                ;;
        esac
        
        cd ../..
        echo "✅ $platform 构建完成"
    else
        echo "❌ $platform 构建失败"
    fi
done

# 显示构建结果
echo ""
echo "📊 构建结果:"
echo "============"
if [ -d "dist" ]; then
    ls -la dist/
    
    echo ""
    echo "📦 生成的文件:"
    for file in dist/*; do
        if [ -f "$file" ]; then
            size=$(du -h "$file" | cut -f1)
            echo "  $(basename "$file") - $size"
        fi
    done
else
    echo "❌ 没有生成任何文件"
fi

echo ""
echo "🎉 多平台构建完成！"
echo ""
echo "📝 使用说明:"
echo "============"
echo "1. macOS 用户: 解压 .zip 文件，右键点击 .app 选择'打开'"
echo "2. Windows 用户: 解压 .zip 文件，运行 .exe 文件"
echo "3. Linux 用户: 解压 .tar.gz 文件，运行可执行文件"
echo ""
echo "💡 提示: 这些文件可以直接分发给用户使用"
