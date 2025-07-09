# 发布指南

## 📦 自动发布流程

本项目使用 GitHub Actions 实现基于 Git 标签的自动构建和发布。

### 🚀 发布新版本

#### 方法1: 使用发布脚本（推荐）

```bash
# 运行发布脚本
./scripts/release.sh
```

脚本会引导您：
1. 选择版本类型（补丁/次要/主要版本）
2. 自动更新 `wails.json` 中的版本号
3. 创建并推送 Git 标签
4. 触发 GitHub Actions 自动构建

#### 方法2: 手动创建标签

```bash
# 更新版本号（可选）
# 编辑 wails.json 中的 productVersion

# 提交更改
git add .
git commit -m "chore: bump version to v1.0.0"

# 创建标签
git tag -a v1.0.0 -m "Release v1.0.0"

# 推送标签
git push origin main
git push origin v1.0.0
```

### 📋 版本号规范

使用 [语义化版本](https://semver.org/lang/zh-CN/) 规范：

- **v1.0.0**: 主要版本（重大更改，可能不兼容）
- **v1.1.0**: 次要版本（新功能，向后兼容）
- **v1.0.1**: 补丁版本（Bug 修复，向后兼容）

### 🔄 自动构建流程

当推送标签时，GitHub Actions 会自动：

1. **多平台构建**：
   - Windows (amd64)
   - macOS (amd64 + arm64)
   - Linux (amd64)

2. **生成发布包**：
   - Windows: `.zip` 格式
   - macOS: `.zip` 格式（包含 .app）
   - Linux: `.tar.gz` 格式

3. **创建 GitHub Release**：
   - 自动生成发布说明
   - 上传所有平台的构建文件
   - 包含下载链接和使用说明

### 📁 构建产物

每个平台的构建产物：

```
MQToolkit-windows-amd64.zip     # Windows 版本
MQToolkit-darwin-amd64.zip      # macOS Intel 版本
MQToolkit-darwin-arm64.zip      # macOS Apple Silicon 版本
MQToolkit-linux-amd64.tar.gz    # Linux 版本
```

### 🛠️ 本地测试构建

在发布前，可以本地测试多平台构建：

```bash
# 测试所有平台构建
./scripts/build-all.sh

# 测试单个平台
wails build -platform darwin/amd64
wails build -platform windows/amd64
wails build -platform linux/amd64
```

### 📊 监控构建状态

- **Actions 页面**: 查看构建进度和日志
- **Releases 页面**: 查看发布的版本
- **构建时间**: 通常需要 10-15 分钟完成所有平台

### ⚠️ 注意事项

1. **标签格式**: 必须以 `v` 开头，如 `v1.0.0`
2. **分支要求**: 建议在 `main` 或 `master` 分支创建标签
3. **权限要求**: 需要仓库的 push 权限
4. **构建依赖**: GitHub Actions 会自动安装所需依赖

### 🔧 故障排除

#### 构建失败
- 检查 Actions 页面的错误日志
- 确认代码在本地能正常构建
- 检查依赖是否正确安装

#### 发布失败
- 确认 GitHub Token 权限
- 检查标签格式是否正确
- 确认没有重复的标签名

#### macOS 签名问题
- 当前使用临时签名（adhoc）
- 用户首次运行需要手动允许
- 生产环境建议使用开发者证书

### 📝 发布检查清单

发布前确认：

- [ ] 代码已测试并正常工作
- [ ] 更新了版本号和更新日志
- [ ] 本地构建测试通过
- [ ] 提交了所有更改
- [ ] 选择了正确的版本号
- [ ] 标签格式正确（vX.Y.Z）

### 🔗 相关链接

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [Wails 构建文档](https://wails.io/docs/guides/building/)
- [语义化版本规范](https://semver.org/lang/zh-CN/)
