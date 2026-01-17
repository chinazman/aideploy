# 更新日志

本项目的所有重要更改都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

## [1.0.0] - 2025-01-17

### 新增
- 子域名和路径两种部署模式
- Docker 容器化部署支持
- Git 版本控制和回滚功能
- 增量部署和全量部署支持
- 多平台二进制文件构建
- GitHub Actions CI/CD 自动化

### 修复
- 子域名模式下单层基础域名的解析问题
- 支持 `jydemo.localhost` 和 `site.example.com` 两种格式

### 安全
- Docker 容器使用非 root 用户运行
- API 密钥验证支持

---

## 版本说明

### 版本号格式
- **主版本号 (MAJOR)**: 不兼容的 API 修改
- **次版本号 (MINOR)**: 向下兼容的功能性新增
- **修订号 (PATCH)**: 向下兼容的问题修正

### 更新类型
- **新增 (Added)**: 新增功能
- **变更 (Changed)**: 功能变更
- **弃用 (Deprecated)**: 即将移除的功能
- **移除 (Removed)**: 已移除的功能
- **修复 (Fixed)**: Bug 修复
- **安全 (Security)**: 安全相关的修复
