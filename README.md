<div align="center">
  <img src="assets/golazo-logo.png" alt="Golazo logo" width="150">
  <h1>Golazo 中文版</h1>
</div>

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
![macOS](https://img.shields.io/badge/macOS-000000?logo=apple&logoColor=white)
![Linux](https://img.shields.io/badge/Linux-FCC624?logo=linux&logoColor=black)
![Windows](https://img.shields.io/badge/Windows-0078D6?logo=windows&logoColor=white)

一个中文化的足球实时比分终端应用。基于 [0xjuanma/golazo](https://github.com/0xjuanma/golazo) 修改。

</div>

## 项目说明

这是 Golazo 的中文汉化版，适合希望在终端里用中文查看足球赛况的用户。

主要改动：

- 界面文案中文化
- 命令行帮助中文化
- 支持动态实体名翻译：球队、球员、联赛、球场、裁判
- 自动翻译结果会缓存到本地，减少重复联网请求
- 新增只翻译界面的入口，保留球队/球员等专名原文

## 汉化版声明

本项目是基于原开源项目 [Golazo](https://github.com/0xjuanma/golazo) 的非官方中文本地化版本。

- 本项目不是原作者发布的官方版本。
- 原项目版权归原作者 Juanma Roca 及相关贡献者所有。
- 本项目仅在 MIT License 允许范围内进行本地化修改、再分发和发布。
- 若原项目继续更新，本仓库可能不会与上游保持实时同步。
- 足球数据来自项目原有数据源，本项目不对第三方数据的完整性、准确性或可用性作保证。
- 动态实体名翻译依赖外部翻译服务和本地缓存，翻译结果可能不完全准确。

## 功能

- 实时比赛列表和逐分钟事件
- 已完赛比赛结果
- 比赛详情、数据统计、积分榜、阵容
- 官方集锦和进球回放链接
- 进球桌面通知
- 65+ 联赛偏好设置
- 两种中文模式：
  - 完整中文化：界面和实体名都翻译
  - 仅界面中文化：球队、球员、联赛、球场、裁判名保留原文

## 安装

### 从源码构建

```bash
git clone https://github.com/gigode/golazo-cn.git
cd golazo-cn
go build
./golazo
```

### 安装到本地 PATH

```bash
go build
install -m 0755 ./golazo ~/.local/bin/golazo
ln -sf ~/.local/bin/golazo ~/.local/bin/golazo-ui-only
```

如果 `~/.local/bin` 不在 `PATH` 中，需要先加入 shell 配置。

## 使用

完整中文化模式：

```bash
golazo
```

仅界面中文化，保留球队/球员/联赛/球场/裁判名原文：

```bash
golazo-ui-only
```

或者：

```bash
golazo --ui-only
```

显示帮助：

```bash
golazo --help
```

## 动态翻译缓存

完整中文化模式下，首次遇到未收录的英文实体名时会尝试自动翻译成简体中文，并缓存到：

```text
~/.cache/golazo/translations_zh.json
```

后续再次遇到同名实体时会直接使用本地缓存。翻译服务不可用时会保留原文，不影响应用运行。

## 快捷键

- `↑` / `↓` 或 `j` / `k`：导航
- `Enter`：选择
- `/`：筛选
- `Tab`：聚焦详情
- `Esc`：返回或关闭弹窗
- `r`：刷新
- `q`：退出

## 原项目

- 上游仓库：[0xjuanma/golazo](https://github.com/0xjuanma/golazo)
- 原作者：[@0xjuanma](https://github.com/0xjuanma)

## 版权与许可证

本项目保留原项目的 MIT License。

原始版权声明：

```text
MIT License

Copyright (c) 2025 Juanma Roca
```

完整许可证内容见 [LICENSE](LICENSE)。
