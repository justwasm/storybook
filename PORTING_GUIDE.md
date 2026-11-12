# Bubble Tea v2 Component & Example Porting Guide
## 组件与示例接入与移植全景指南

> **基座定位**：`github.com/justwasm/storybook` 是一个面向 **Charmbracelet Bubble Tea v2 (`charm.land/bubbletea/v2`)** 的现代化通用组件展示框架（Substrate / Harness）。  
> 本指南旨在规范独立 TUI 组件、微应用（Micro-apps）及官方 Showcases 的零侵入接入流程，确保遵循 Bubble Tea v2 标准、高内聚低耦合、以及可无缝独立抽离迁移。

---

## 目录
1. [核心架构与设计哲学](#1-核心架构与设计哲学)
2. [独立程序 vs 嵌入式基座的 5 大阻抗失真](#2-独立程序-vs-嵌入式基座的-5-大阻抗失真)
3. [零侵入标准接入流程 (3 步法)](#3-零侵入标准接入流程-3-步法)
4. [7 类典型场景的最佳实战配方](#4-7-类典型场景的最佳实战配方)
5. [双轴状态协议 (Dual-Axis Protocol)](#5-双轴状态协议-dual-axis-protocol)
6. [API 参考与基座扩展点](#6-api-参考与基座扩展点)
7. [反模式与质量自检清单](#7-反模式与质量自检清单)

---

## 1. 核心架构与设计哲学

Storybook 基座遵循以下四大承重设计原则：

```
┌────────────────────────────────────────────────────────┐
│                   Top Status / Brand                   │
├───────────────────────────────┬────────────────────────┤
│ ▼ Category A                  │ [1] [2] [3] (States)   │
│   ● active-story              │                        │
│   • another-story             │      Right Canvas      │
│ ▼ Category B                  │     (Sub-Rectangle)    │
│   • list-item                 │                        │
├───────────────────────────────┴────────────────────────┤
│ [1] Live Demo   [2] Source Code   [3] Info / Arch      │
└────────────────────────────────────────────────────────┘
```

1. **宿主隔离（Host Isolation）**：
   - 任何子组件的 `tea.Quit`、`panic` 或未捕获事件，都不得导致主宿主进程崩溃。
   - 退出指令会被宿主安全拦截并呈现沙箱退出横幅（`MockQuit`），支持按 `r` 或空格一键恢复。
2. **零侵入适配（Zero-Invasive Adaptation）**：
   - 原始业务模型（`tea.Model`）**无需引用任何 Storybook 专有包**即可独立存在。
   - 适配层作为单独的注册声明注入，保证组件随时可以作为独立的 Bubble Tea CLI 工具发布或被其他项目导入。
3. **双轴导航矩阵（Dual-Axis Matrix）**：
   - **X-轴（变种轴，按键 `1..9`）**：切换同一组件的不同状态预设（如 Default、Hover、Disabled、Loading）。
   - **Y-轴（视点轴，按键 `v/d/s/i`）**：切换实时交互（Live Demo）、语法高亮源码（Source Code）、架构说明（Architecture Info）。
4. **插件化发现（Plug-and-Play via Blank Imports）**：
   - 通过 Go 的 `_ "my/component/plugin"` 机制实现高内聚自动注册，彻底解耦主程序与各示例包。

---

## 2. 独立程序 vs 嵌入式基座的 5 大阻抗失真

在独立 CLI（`examples/*/main.go`）中运行完全正常的代码，直接移入多组件宿主时，必定会遇到以下 5 大本质阻抗失真。基座已通过内置中间层解决了绝大部分，但开发者需理解其原理：

### 阻抗 1：全屏主权假设（The Full-Screen Sovereign Fallacy）
* **现象**：表格被截断、`glamour` Markdown 无法贴边排版或滚动条错乱。
* **原因**：独立程序独占整个终端，直接拿 OS 的 `SIGWINCH` 物理窗口尺寸；而在基座中，组件只占有右侧的局部矩形视口（扣除了侧边栏 33 列、边框 1 列、顶部面包屑与底部状态栏）。
* **基座解决方案**：基座在挂载时及终端缩放时，自动计算局部可用矩形并向组件发送经过裁切计算的 `tea.WindowSizeMsg{Width, Height}`，或调用 `SetSize(w, h)`。

### 阻抗 2：焦点与按键路由失真（Focus Multiplexing）
* **现象**：`chat`、`textarea` 或 `textinput` 点进去了却无法输入字符。
* **原因**：`bubbles` 中的输入组件内部有 `focused bool` 守卫，失焦状态下其 `Update(msg)` 会在第一行代码静默丢弃字符按键。
* **基座解决方案**：基座提供 `Focusable` 自动嗅探与 `WithAutoFocus(true)`。当进入交互模式（`INTERACTIVE`）或组件初始化时，适配层自动调用其 `Focus()` 方法激活光标。

### 阻抗 3：异步命令与冷启动陷阱（The Cold-Start Zero-State Trap）
* **现象**：`file-picker` 突然只显示一行 `Bummer. No Files Found.`。
* **原因**：`filepicker` 的列表展开依赖 `fp.Init()` 返回的异步 `tea.Cmd` 执行 `readDir`。在宿主快速切换视图或轻量预览时，如果首帧未预先读取，组件就会停留在零高度的空状态。
* **最佳实践**：对重度依赖 I/O 初始化展开尺寸的组件，提供初始同步数据填充或显式设置初始高度 `SetHeight`。

### 阻抗 4：鼠标绝对坐标偏置（Coordinate Offset）
* **现象**：`clickable`、`mouse` 例子中的按钮明明点在上面，却毫无反应。
* **原因**：原始 `tea.MouseMsg.X/Y` 是屏幕绝对坐标（从 `(0, 0)` 开始算）。而组件画布左上角在 `(34, 3)`。如果直传绝对坐标，组件内部 Hit-Test 全部落空。
* **基座解决方案**：基座内建 `TranslateMouseMsg(msg, offsetX, offsetY)`，在将鼠标事件投递给组件前，已透明转换为以画布左上角为 `(0, 0)` 的局部相对坐标。

### 阻抗 5：生命周期与转义序列污染（Lifecycle & Escape Code Sandboxing）
* **现象**：部分终端特性例子调用了全屏切换或退出指令，导致整个基座布局崩溃。
* **基座解决方案**：基座 `ModelAdapter` 拦截任何返回 `tea.QuitMsg` 的 `tea.Cmd`，转换为受控的 `MockQuit` 状态，防止子组件意外中止宿主进程。

---

## 3. 零侵入标准接入流程 (3 步法)

### 第一步：保留纯净的业务模型（Zero Coupling）
确保你的组件或示例完全遵循标准 Bubble Tea v2，不要引入任何与 storybook 相关的依赖：

```go
// pkg/counter/counter.go
package counter

import (
	"fmt"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Count int
}

func New() Model {
	return Model{Count: 0}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k", "+":
			m.Count++
		case "down", "j", "-":
			m.Count--
		case "r":
			m.Count = 0
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	return tea.NewView(fmt.Sprintf("Counter Value: %d\n\n(↑/k: Inc • ↓/j: Dec • r: Reset)", m.Count))
}
```

### 第二步：编写单文件注册插件（High Cohesion）
在组件目录下或单独的 `plugins` 目录下，添加 `story.go`：

```go
// pkg/counter/story.go
package counter

import (
	_ "embed"
	tea "charm.land/bubbletea/v2"
	"github.com/justwasm/storybook"
)

//go:embed counter.go
var sourceCode string

func init() {
	storybook.RegisterModel(
		"counter",                        // 唯一标识符 (URL/搜索友好)
		"Interactive Counter",            // 显示名称
		storybook.CategoryBasics,         // 所属分类
		func() tea.Model { return New() },// 模型工厂
		storybook.WithLevel(storybook.LevelAtom),
		storybook.WithDescription("A minimal state counter demonstrating zero-invasive Bubble Tea v2 integration."),
		storybook.WithSourceCode(sourceCode),
		storybook.WithUpstream("pkg/counter/counter.go"),
		storybook.WithKeyFeatures(
			"Pure Bubble Tea v2 Model",
			"Zero external dependencies outside Bubble Tea",
			"Automatic mock-quit and restart integration",
		),
	)
}
```

### 第三步：在聚合包中引入（Blank Import）
在你的展示端（如 `cmd/storybook/main.go`）聚合引入：

```go
package main

import (
	_ "myproject/pkg/counter" // 自动触发 init() 注册
	"github.com/justwasm/storybook"
)

func main() {
	if err := storybook.Run(storybook.WithTitle("My Component Storybook")); err != nil {
		panic(err)
	}
}
```

---

## 4. 7 类典型场景的最佳实战配方

### 配方 1：响应式 / 容器自适应组件（Responsive Layouts）
* **适用**：`glamour`、`pager`、`table`、`list`、`split-editors`。
* **要点**：
  * 在 `Update(tea.WindowSizeMsg)` 中计算子视口；
  * `ModelAdapter` 会自动向组件派发 `tea.WindowSizeMsg`；
  * 若实现了 `storybook.Resizable` 接口，基座亦可直接调用 `SetSize(w, h)`。

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.SetWidth(msg.Width - 2)
		m.viewport.SetHeight(msg.Height - 4)
		return m, nil
	}
	// ...
}
```

### 配方 2：表单与输入框组件（Focus & Form Inputs）
* **适用**：`textarea`、`textinput`、`chat`。
* **要点**：
  * 实现 `Focus() tea.Cmd` 或 `Focus()`；
  * `ModelAdapter` 默认开启 `WithAutoFocus(true)`，在进入交互模式或初始化时会自动为输入控件唤醒焦点，防止字符被吞。

```go
func (m *ChatModel) Focus() tea.Cmd {
	return m.textarea.Focus()
}
```

### 配方 3：鼠标交互与命中检测（Mouse & Hitbox）
* **适用**：`clickable`、`mouse`、`tabs`。
* **要点**：
  * 无需任何特殊改动！基座自动完成屏幕绝对坐标到局部画布相对坐标的映射。
  * 编写你的组件时，假定左上角就是 `(0, 0)` 即可。

### 配方 4：持续动画与定时器（Tick & Animation）
* **适用**：`doom-fire`、`spinner`、`splash`、`realtime`。
* **要点**：
  * 确保在离开页面或切换时，Tick 消息不会继续无限递归调度造成内存与 CPU 泄漏；
  * 实现 `storybook.Restarter` 接口或借助 `WithResetFactory`，使得重启时能够从初始帧干净重开。

### 配方 5：异步 I/O 与文件系统（Async File I/O）
* **适用**：`file-picker`、`http`、`exec`。
* **要点**：
  * 避免构造时直接空高度（`height = 0`）；
  * 建议在构造函数或 `Init()` 中指定安全的高度下限（如 `SetHeight(15)`），并在接收到 `WindowSizeMsg` 后动态放大。

---

## 5. 双轴状态协议 (Dual-Axis Protocol)

Storybook 支持为单个组件注册多个状态变种（通过键盘 `1..9` 快速切换）。

### 声明状态变种：
```go
storybook.RegisterModel(
	"btn",
	"Button",
	storybook.CategoryBasics,
	func() tea.Model { return NewButton() },
	storybook.WithStates(
		storybook.State{ID: "primary", Name: "Primary", Description: "Standard filled button"},
		storybook.State{ID: "secondary", Name: "Secondary", Description: "Outlined button"},
		storybook.State{ID: "disabled", Name: "Disabled", Description: "Unclickable muted button"},
	),
)
```

### 组件支持状态响应：
只需让你的 Model 实现 `storybook.Stateful` 接口：

```go
func (m *ButtonModel) SetState(stateIndex int) {
	switch stateIndex {
	case 0:
		m.Variant = VariantPrimary
	case 1:
		m.Variant = VariantSecondary
	case 2:
		m.Disabled = true
	}
}
```
切换状态时，基座会自动保持在 `[1] Live Demo` 视图，无需重新加载整个故事。

---

## 6. API 参考与基座扩展点

### 注册函数
* `storybook.RegisterModel(id, name, cat, factory, opts...) Story`：推荐的标准接入 API。
* `storybook.Register(Story)`：底层全功能注册函数。
* `storybook.WrapModel(tea.Model, opts...) *ModelAdapter`：底层包装适配器。

### 常用配置选项 (`StoryOption`)
* `WithLevel(AtomicLevel)`：指定原子设计级别（`LevelAtom`、`LevelMolecule`、`LevelOrganism` 等）。
* `WithDescription(string)`：简短用途概述。
* `WithSourceCode(string)`：源码展示字符串（建议结合 `//go:embed`）。
* `WithUpstream(string)`：上游源码路径。
* `WithKeyFeatures(...string)`：核心能力亮点列表。
* `WithStates(...State)`：声明多状态变种列表。

### 包装适配选项 (`WrapOption`)
* `WithAutoFocus(bool)`：开启/关闭自动焦点激活（默认开启）。
* `WithInitialSize(width, height int)`：设置预挂载视口宽高。
* `WithResetFactory(func() tea.Model)`：提供无状态重建工厂，支持一键重置。
* `WithCustomQuitKey(...string)`：设置触发模拟退出的自定义按键。

### 组件可选契约接口（Duck Typing）
| 接口定义 | 说明 |
|---|---|
| `Resizable { SetSize(w, h int) }` | 显式接收局部画布尺寸变化 |
| `Focusable { Focus() tea.Cmd }` | 获得焦点生命周期并返回命令 |
| `SimpleFocusable { Focus() }` | 获得焦点生命周期（无返回值） |
| `Stateful { SetState(idx int) }` | 响应双轴矩阵的变种切换 (`1..9`) |
| `Restarter { Restart() tea.Cmd }` | 自定义内部状态重置逻辑 |
| `InteractiveHandler { HandleKey(tea.KeyMsg) (tea.Cmd, bool) }` | 接管交互模式下的原始按键 |
| `MouseHandler { HandleMouse(tea.MouseMsg) (tea.Cmd, bool) }` | 接管局部相对坐标鼠标事件 |

---

## 7. 反模式与质量自检清单

### ❌ 严禁的反模式 (Antipatterns)
1. **严禁调用 `os.Exit(0)`**：必须返回 `tea.Quit`，由基座统一拦截为 `MockQuit`。
2. **严禁调用 `term.GetSize()` 绕过基座**：直接探测 OS 终端会破坏右侧画布的边界约束。
3. **严禁假定鼠标事件为屏幕绝对坐标**：基座已处理好偏移，禁止手动二次加减侧边栏宽度。
4. **严禁在组件包内部反向 import 宿主主包**：保持业务组件依赖单向、纯净。

###  质量自检 Checklist (上线前核对)
* [ ] 独立性：组件是否能在不依赖 storybook 的情况下通过 `tea.NewProgram(m).Run()` 独立启动？
* [ ] 视口适应：在终端窗口大幅拉大或缩小时，组件是否能流畅跟随伸缩而不产生越界折行？
* [ ] 交互模式：按下 `Enter` 激活 `● INTERACTIVE` 模式后，按键输入与鼠标点击是否精准响应？
* [ ] 安全退出：按 `q` 或 `Ctrl+C` 是否正确展示 MockQuit 横幅，按 `r` 能否完整复原？
* [ ] 坐标检测：如果有点击区域，鼠标点击命中是否与视觉位置 100% 吻合？
* [ ] 源码展示：按 `s` 切换到 Source Code Tab 时，是否能高亮查阅组件的完整源码？
