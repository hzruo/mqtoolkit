#!/bin/bash

# MQ Toolkit 版本发布脚本
# 用于创建新版本标签并触发 GitHub Actions 自动构建

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 检查是否在 git 仓库中
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "当前目录不是 Git 仓库"
    exit 1
fi

# 检查是否有未提交的更改
if ! git diff-index --quiet HEAD --; then
    print_error "存在未提交的更改，请先提交所有更改"
    git status --porcelain
    exit 1
fi

# 获取当前分支
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
    print_warning "当前不在主分支 ($CURRENT_BRANCH)，建议切换到 main 或 master 分支"
    read -p "是否继续？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 获取最新的标签
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
print_info "当前最新标签: $LATEST_TAG"

# 解析版本号
if [[ $LATEST_TAG =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    MAJOR=${BASH_REMATCH[1]}
    MINOR=${BASH_REMATCH[2]}
    PATCH=${BASH_REMATCH[3]}
else
    print_warning "无法解析当前标签版本号，使用默认版本 0.0.0"
    MAJOR=0
    MINOR=0
    PATCH=0
fi

# 计算建议的版本号
NEXT_PATCH="v$MAJOR.$MINOR.$((PATCH + 1))"
NEXT_MINOR="v$MAJOR.$((MINOR + 1)).0"
NEXT_MAJOR="v$((MAJOR + 1)).0.0"

echo ""
print_info "版本发布选项:"
echo "1. 补丁版本 (Bug 修复): $NEXT_PATCH"
echo "2. 次要版本 (新功能): $NEXT_MINOR"
echo "3. 主要版本 (重大更改): $NEXT_MAJOR"
echo "4. 自定义版本"
echo "5. 退出"

echo ""
read -p "请选择版本类型 (1-5): " -n 1 -r
echo

case $REPLY in
    1)
        NEW_VERSION=$NEXT_PATCH
        ;;
    2)
        NEW_VERSION=$NEXT_MINOR
        ;;
    3)
        NEW_VERSION=$NEXT_MAJOR
        ;;
    4)
        read -p "请输入自定义版本号 (格式: v1.2.3): " NEW_VERSION
        if [[ ! $NEW_VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            print_error "版本号格式错误，应为 vX.Y.Z 格式"
            exit 1
        fi
        ;;
    5)
        print_info "已取消"
        exit 0
        ;;
    *)
        print_error "无效选择"
        exit 1
        ;;
esac

# 检查标签是否已存在
if git tag -l | grep -q "^$NEW_VERSION$"; then
    print_warning "标签 $NEW_VERSION 已存在"
    echo ""
    read -p "是否删除现有标签并重新创建？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "删除现有标签 $NEW_VERSION"

        # 删除本地标签
        git tag -d "$NEW_VERSION" 2>/dev/null || true
        print_success "已删除本地标签"

        # 删除远程标签
        if git push origin --delete "$NEW_VERSION" 2>/dev/null; then
            print_success "已删除远程标签"
        else
            print_warning "远程标签删除失败或不存在"
        fi
    else
        print_info "已取消发布"
        exit 0
    fi
fi

# 确认发布
echo ""
print_info "准备发布版本: $NEW_VERSION"
print_warning "这将会:"
echo "  1. 创建 Git 标签: $NEW_VERSION"
echo "  2. 推送标签到远程仓库"
echo "  3. 触发 GitHub Actions 自动构建和发布"

echo ""
read -p "确认发布？(y/N): " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_info "已取消发布"
    exit 0
fi

# 跳过本地版本号更新，由 GitHub Actions 在构建时处理
print_info "版本号将在 GitHub Actions 构建时自动更新"

# 创建标签
print_info "创建标签 $NEW_VERSION"
git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"
print_success "已创建标签 $NEW_VERSION"

# 推送标签到远程仓库
print_info "推送标签到远程仓库"
git push origin "$NEW_VERSION"
print_success "已推送标签到远程仓库"

echo ""
print_success "🎉 版本 $NEW_VERSION 发布成功！"
echo ""
print_info "接下来的步骤:"
echo "  1. GitHub Actions 将自动开始构建"
echo "  2. 构建完成后会自动创建 Release"
echo "  3. 可以在 GitHub 仓库的 Actions 页面查看构建进度"
echo "  4. 构建完成后在 Releases 页面查看发布的版本"

echo ""
print_info "相关链接:"
echo "  - Actions: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/actions"
echo "  - Releases: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/releases"
