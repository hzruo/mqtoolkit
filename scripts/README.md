# MQ Toolkit 图标管理

本目录包含用于管理 MQ Toolkit 应用图标的脚本和工具。

## 📁 文件说明

- `generate-icons.sh` - 图标生成脚本，从 SVG 源文件生成各平台所需的图标文件

## 🎨 图标生成

### 前置要求

在运行图标生成脚本之前，请确保安装了以下工具之一：

**macOS:**
```bash
brew install imagemagick librsvg
```

**Ubuntu/Debian:**
```bash
sudo apt-get install imagemagick librsvg2-bin
```

**Windows:**
- 下载并安装 [ImageMagick](https://imagemagick.org/script/download.php#windows)

### 使用方法

1. 确保项目根目录有 `logo.svg` 文件
2. 运行图标生成脚本：

```bash
./scripts/generate-icons.sh
```

### 生成的文件

脚本会生成以下图标文件：

| 文件路径 | 用途 | 尺寸 |
|---------|------|------|
| `build/appicon.png` | macOS 应用图标 | 512x512 |
| `build/windows/icon.ico` | Windows 应用图标 | 多尺寸 ICO |
| `build/linux/icon.png` | Linux 应用图标 | 512x512 |
| `frontend/src/assets/logo.svg` | 前端 Logo | 矢量 |
| `frontend/public/favicon.png` | 网页 Favicon | 32x32 |

## 🔧 自定义图标

### 修改源图标

1. 编辑项目根目录的 `logo.svg` 文件
2. 运行 `./scripts/generate-icons.sh` 重新生成所有图标
3. 运行 `wails build` 构建应用以应用新图标

### 手动替换图标

如果需要手动替换特定平台的图标：

**macOS:**
- 替换 `build/appicon.png`

**Windows:**
- 替换 `build/windows/icon.ico`

**Linux:**
- 替换 `build/linux/icon.png`

## 📱 图标规范

### 设计要求

- **格式**: SVG（源文件）
- **尺寸**: 矢量格式，建议基于 512x512 画布
- **颜色**: 支持全彩，建议使用品牌色彩
- **背景**: 透明背景
- **风格**: 简洁、现代、易识别

### 平台特定要求

**macOS:**
- 推荐尺寸: 512x512, 1024x1024
- 格式: PNG
- 圆角: 系统自动处理

**Windows:**
- 推荐尺寸: 16x16, 32x32, 48x48, 256x256
- 格式: ICO（包含多个尺寸）
- 背景: 透明

**Linux:**
- 推荐尺寸: 512x512
- 格式: PNG
- 背景: 透明

## 🚀 构建应用

生成图标后，使用以下命令构建应用：

```bash
# 构建当前平台
wails build

# 构建特定平台
wails build -platform windows/amd64
wails build -platform darwin/amd64
wails build -platform linux/amd64
```

## 🔍 故障排除

### 图标不显示

1. 确认图标文件存在且路径正确
2. 检查 `wails.json` 中的 `iconPath` 配置
3. 重新运行图标生成脚本
4. 清理构建缓存后重新构建

### 生成脚本失败

1. 检查是否安装了 ImageMagick 或 librsvg
2. 确认 `logo.svg` 文件存在且有效
3. 检查文件权限
4. 查看错误信息并根据提示解决

### 图标质量问题

1. 确保源 SVG 文件质量良好
2. 检查 SVG 是否使用了不支持的特性
3. 考虑手动优化特定尺寸的图标

## 📝 注意事项

- 修改图标后需要重新构建应用才能生效
- 不同平台的图标要求可能不同，建议测试各平台效果
- SVG 源文件应保持简洁，避免过于复杂的图形
- 建议定期备份自定义图标文件
