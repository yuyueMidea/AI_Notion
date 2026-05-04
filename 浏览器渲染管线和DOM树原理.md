<img width="1440" height="1040" alt="image" src="https://github.com/user-attachments/assets/c016afc4-be5e-4ed1-98b9-2bad7b59f57d" />


上面是整体流程的鸟瞰图。下面逐层拆解每个阶段的原理。

---

## 第一阶段：解析与 AST

浏览器收到服务器返回的字节流后，第一步是**词法分析（Tokenization）**——将原始字符流切割成有意义的最小单元（Token）。

HTML 的 Token 长这样：开始标签 `<div>`、结束标签 `</div>`、文本内容、注释等。CSS 的 Token 则是选择器、属性名、属性值、冒号、分号等。

切完 Token 之后，解析器进行**语法分析**，把 Token 序列按语法规则组装成 **AST（抽象语法树）**。AST 是一个纯粹的内存中间结构，它忠实描述了代码的嵌套与语义关系，但不包含任何渲染信息。你可以把 AST 想成"翻译官先在脑子里理解了整个句子的语法结构，还没有开始做任何事"。---

## 第二阶段：DOM 树与 CSSOM 树

AST 是临时中间产物。解析器在构建 AST 的同时（或紧接着），会将其转化为两棵真正服务于渲染的"活树"：

**DOM 树（Document Object Model）** 是 HTML 解析的最终产物。每个 HTML 标签对应一个节点，文本内容也是节点，注释也是节点。DOM 是一个**实时的 JavaScript 可操作对象图**——你用 `document.querySelector` 拿到的，就是 DOM 树上的节点。

**CSSOM 树（CSS Object Model）** 是 CSS 解析的产物。它将所有 CSS 规则（来自外链样式表、`<style>` 标签、内联 style 属性）组织成一棵树，并完成**样式继承（cascade）和优先级计算**。`font-size` 在父节点上设置后，子节点的 CSSOM 节点会继承这个值，这正是"层叠样式"中"层叠"的含义。

关键点：**CSSOM 是渲染阻塞的**。在 CSSOM 构建完成之前，浏览器无法进行后续的渲染工作——因为样式不完整，渲染出错了会更糟。这就是为什么 CSS 要放在 `<head>` 中尽早加载。

---

## 第三阶段：合并为 Render Tree

DOM 树 + CSSOM 树 → **Render Tree（渲染树）**。

合并规则有两条值得特别注意：

- `display: none` 的节点**不会进入** Render Tree（它们在 DOM 里存在，但渲染树里没有）
- `visibility: hidden` 的节点**会进入** Render Tree（占位，只是不可见）

Render Tree 的每个节点叫做 **RenderObject**（在 Chromium 里叫 `LayoutObject`），它既包含内容信息（来自 DOM），也包含完整的计算样式（来自 CSSOM）。---

## 第四阶段：渲染管线 — Layout → Paint → Composite

这是最终把像素"画"到屏幕上的三步流水线，每一步的代价都不同。

**Layout（布局 / 回流 Reflow）**：遍历 Render Tree，根据 CSS 盒模型计算每个节点的精确位置和尺寸（x, y, width, height）。这是 CPU 密集型操作，任何会改变"几何关系"的属性（`width`、`margin`、`position`、添加/删除 DOM 节点）都会触发它。

**Paint（绘制 / 重绘 Repaint）**：遍历 Layout 树，生成**绘制记录（Paint Records）**——"在坐标 (x,y) 画一个宽高为 w×h 的矩形，颜色为 #xxx"。这是一系列绘图指令，还不是真正的像素。`color`、`background`、`box-shadow` 等视觉属性的变化会触发重绘（但不一定触发 Reflow）。

**Composite（合成）**：现代浏览器会将页面分成若干**图层（Layers）**，每个图层独立光栅化（Rasterize）成位图，最后由 **GPU** 将所有图层合并输出到屏幕。`transform`、`opacity` 这类属性的动画之所以流畅，是因为它们只在合成阶段发生，完全走 GPU，跳过了 Layout 和 Paint。---

## 把四个角色串起来

| 角色 | 本质 | 存在于哪个阶段 | 谁消费它 |
|---|---|---|---|
| **AST** | 源码的语法结构（内存中间态） | 解析阶段，短暂存在 | 后续树构建器 |
| **DOM 树** | HTML 的节点对象图，JS 可操作 | 持续存在，随 JS 动态变化 | 渲染引擎 + JavaScript |
| **CSSOM 树** | 所有 CSS 规则 + 继承计算结果 | 持续存在，CSS 变化时重建 | 渲染引擎 |
| **渲染管线** | Layout → Paint → Composite 三步 | 每帧触发（~16ms/帧） | GPU / 屏幕 |

---

## 两个最容易混淆的问题

**Reflow vs Repaint 的区别**：Reflow（回流）是 Layout 阶段重新计算，它一定会触发后续的 Paint 和 Composite，代价最高。Repaint（重绘）跳过 Layout，只重新 Paint，代价中等。只修改 `transform`/`opacity` 则只触发 Composite，代价最低，这是 CSS 动画性能优化的核心原则。

**为何 JS 放 `<body>` 末尾**：JS 执行会阻塞 HTML 解析（因为 JS 可能操作 DOM），所以脚本放末尾能让页面先解析完 HTML 构建 DOM，再运行脚本，用户更快看到内容。`defer`/`async` 属性是更现代的解法。
