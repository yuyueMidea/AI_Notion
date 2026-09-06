可以。对于前端开发来说，Chrome DevTools 的 **Sources（源代码）面板**其实是最值得深入掌握的调试工具之一。

很多开发者对 Sources 的使用还停留在：

> 打开 Sources → 找到 JS → 打断点 → 看变量。

但真正把 Sources 用熟之后，它可以变成一套完整的**程序运行过程观测系统**：

> **断点 → 调用栈 → 作用域 → 单步执行 → 事件监听 → 异步链路 → DOM/JS 联动 → Source Map → Overrides → Snippets → Blackbox → 性能定位**

尤其你平时涉及 Vue、WebSocket、文件上传、Three.js、复杂前端交互等项目，Sources 的价值会非常大。

---

# 一、先建立一个正确认识：Sources 到底是什么？

可以把 Sources 理解成：

> **浏览器里观察 JavaScript 程序“运行过程”的实验室。**

前端代码从：

```text
Vue / React / 原生 JS
        ↓
Webpack / Vite / Rollup
        ↓
Bundle
        ↓
Source Map
        ↓
浏览器加载
        ↓
JavaScript Engine 执行
```

Sources 看到的并不一定就是你磁盘上的原始代码。

例如你写：

```js
function login() {
    const username = input.value
    api.login(username)
}
```

Vite/Webpack 可能最终生成：

```js
(()=>{const e=t.value;return fetch("/api/login",{method:"POST",body:e})})()
```

而 Sources 通过 **Source Map**，让你能够重新看到：

```text
src/
├── views/
│   └── Login.vue
├── api/
│   └── user.js
└── utils/
    └── request.js
```

所以 Sources 实际上连接了：

```text
你的源代码
      ↓
构建系统
      ↓
Source Map
      ↓
浏览器运行时代码
      ↓
JavaScript 引擎
```

这也是为什么理解 Sources，会顺带帮助你理解：

* Vite
* Webpack
* Source Map
* JavaScript Runtime
* Event Loop
* Promise
* async/await
* Vue Runtime

---

# 二、Sources 面板主要有哪些东西？

打开：

```text
F12
→ Sources
```

通常可以看到几个核心区域：

```text
┌─────────────────────────────────────────────┐
│ Sources                                     │
├──────────────┬────────────────┬─────────────┤
│ 文件树       │     源代码      │ Debug 区域   │
│              │                │             │
│ Page         │ 代码内容        │ Watch       │
│ Filesystem   │                │ Scope       │
│ Overrides    │                │ Call Stack   │
│ Snippets     │                │ Breakpoints  │
└──────────────┴────────────────┴─────────────┘
```

其中最重要的是：

### 左侧

负责：

> **找代码**

### 中间

负责：

> **看代码、打断点、执行代码**

### 右侧

负责：

> **理解代码现在运行到了哪里**

真正高频的其实是右侧。

---

# 三、第一核心能力：断点调试

这是 Sources 最基本，也是最重要的能力。

比如：

```js
function calculate(a, b) {
    const result = a + b
    return result
}
```

你在：

```js
const result = a + b
```

这一行点击左侧行号。

出现：

```text
🔵 23
```

说明断点建立。

然后执行：

```js
calculate(10, 20)
```

程序就会停下来。

---

# 四、为什么“断点”如此重要？

因为普通：

```js
console.log()
```

只能告诉你：

> “某个时间点发生了什么。”

而 Debugger 可以告诉你：

> **程序现在正在发生什么。**

例如：

```js
function checkout(cart) {
    const total = calculateTotal(cart)

    if (total > 1000) {
        createDiscount()
    }

    submitOrder(cart)
}
```

如果订单金额计算错了。

你当然可以：

```js
console.log(cart)
console.log(total)
```

但更好的方式是：

```text
checkout()
   ↓
calculateTotal()
   ↓
total = ?
   ↓
if
   ↓
submitOrder()
```

一步一步观察。

---

# 五、Sources 中最重要的功能：单步执行

程序停下来之后，你会看到几个非常重要的按钮。

一般可以理解为：

```text
▶ Resume
↓ Step over
↓ Step into
↗ Step out
```

这四个必须熟练。

---

# 六、Step Over：下一步

假设：

```js
const user = getUser()
const name = user.name
console.log(name)
```

当前停在：

```js
const user = getUser()
```

点击：

> Step over

程序执行：

```js
getUser()
```

但是：

> **不会进入 getUser() 内部。**

直接到：

```js
const name = user.name
```

适合：

> 我知道这个函数大概没问题，我只想继续往下走。

---

# 七、Step Into：钻进去

假设：

```js
const user = getUser()
```

点击：

> Step into

那么会进入：

```js
function getUser() {
    return request('/api/user')
}
```

然后继续：

```js
request()
```

再进去：

```js
fetch()
```

甚至可能进入框架代码。

这时候你会发现一个非常重要的问题：

> **千万不要无脑 Step Into。**

因为可能一路进入：

```text
你的代码
 ↓
Axios
 ↓
Interceptor
 ↓
Promise
 ↓
Vue
 ↓
Runtime
 ↓
浏览器内部代码
```

很快就迷路了。

所以后面必须学：

> **Blackbox**

---

# 八、Step Out：从函数里面出来

例如：

```js
function calculate() {
    const a = 10
    const b = 20
    return a + b
}
```

你 Step Into 进去之后发现：

> “这个函数没问题，我不想继续看了。”

直接：

> Step out

就会回到：

```js
const result = calculate()
```

---

# 九、真正厉害的调试方式：Step Over + Step Into 结合

实际项目中通常这样：

```text
业务代码
 ↓
Step Into
 ↓
进入关键函数
 ↓
Step Over
 ↓
观察内部变量
 ↓
发现异常
 ↓
Step Into
 ↓
继续追踪
```

而不是：

```text
疯狂点击 Step Into
```

---

# 十、第二核心能力：Scope

当程序暂停时，右侧会出现：

> Scope

这是极其重要的区域。

例如：

```js
function test() {
    const name = 'Karl'
    const age = 30

    debugger
}
```

停下来以后，你会看到：

```text
Scope

Local
    name: "Karl"
    age: 30

Script
    ...

Global
    window
    document
    location
    ...
```

这实际上对应 JavaScript 的：

> **词法作用域 / 执行上下文**

---

# 十一、Scope 能帮助你理解闭包

例如：

```js
function createCounter() {
    let count = 0

    return function () {
        count++
        return count
    }
}
```

运行：

```js
const counter = createCounter()

counter()
```

在内部断点。

你可能会看到：

```text
Local
    ...

Closure
    count: 0

Global
    window
```

这里的：

```text
Closure
```

就是非常重要的东西。

它直接把：

> JavaScript 闭包

变成了一个可以观察的东西。

---

# 十二、第三核心能力：Call Stack

我认为：

> **Call Stack 是 Sources 里最值得深入理解的功能之一。**

例如：

```js
function A() {
    B()
}

function B() {
    C()
}

function C() {
    debugger
}

A()
```

暂停之后：

```text
Call Stack

C
B
A
<anonymous>
```

这就是：

```text
A()
 ↓
B()
 ↓
C()
 ↓
debugger
```

---

# 十三、Call Stack 到底有什么用？

当你遇到：

```text
某个函数为什么执行了？
```

这是非常关键的。

例如：

```js
saveData()
```

你发现：

> 我根本没有调用 saveData() 啊！

那么在：

```js
saveData()
```

里面打断点。

程序停下来。

看：

```text
Call Stack
```

可能发现：

```text
saveData
handleSubmit
onClick
Vue event handler
```

于是你就知道：

```text
用户点击
 ↓
Vue事件
 ↓
handleSubmit
 ↓
saveData
```

---

# 十四、复杂项目里 Call Stack 尤其重要

例如：

```text
WebSocket.onmessage
 ↓
handleMessage
 ↓
updateStore
 ↓
Pinia action
 ↓
component update
```

你可以完整追踪：

> 一个事件到底是怎么把整个系统推动起来的。

这比：

```js
console.log("hello")
```

强太多。

---

# 十五、第四核心能力：Watch

右侧：

> Watch

可以添加表达式。

例如：

```js
user
```

或者：

```js
user.profile.name
```

或者：

```js
items.length
```

甚至：

```js
cart.reduce((sum, item) => sum + item.price, 0)
```

程序每次暂停时都会重新计算。

---

# 十六、Watch 特别适合复杂业务

例如你有：

```js
const state = {
    user,
    cart,
    orders,
    permissions
}
```

你正在追：

> 为什么按钮突然变成 disabled？

Watch：

```js
user.permissions.includes('admin')
```

或者：

```js
!user || !user.permissions
```

每一步都能观察。

这比不断：

```js
console.log(...)
```

舒服很多。

---

# 十七、Conditional Breakpoint：条件断点

这是生产项目中非常好用的功能。

假设：

```js
for (const item of items) {
    process(item)
}
```

有：

```text
10000 个 item
```

你如果普通断点：

```text
第1个停
第2个停
第3个停
...
```

非常痛苦。

你可以设置条件：

```js
item.id === 9999
```

那么：

> 只有满足条件才暂停。

---

# 十八、条件断点是排查“偶现 Bug”的神器

例如：

```js
function updateUser(user) {
    ...
}
```

Bug：

> 某个特定用户才出现。

可以设置：

```js
user.id === 12345
```

然后：

```text
运行整个系统
       ↓
大量调用
       ↓
只有 user.id === 12345
       ↓
暂停
```

这就是：

> **定点捕获异常现场。**

---

# 十九、Logpoint：不用暂停的“断点日志”

Sources 还可以创建：

> Logpoint

它和：

```js
console.log()
```

类似。

但是优势是：

> **不用修改源码。**

例如：

```js
user.id
```

每次执行到这里就记录。

这对于临时调试非常舒服。

调试完成后：

> 删除 Logpoint。

源代码完全不需要改。

---

# 二十、Exception Breakpoints

这是非常容易被忽略，但特别有价值的功能。

可以让 DevTools：

> **在异常真正发生的位置停下来。**

例如：

```js
const user = null

user.profile.name
```

普通情况下可能最终只看到：

```text
Cannot read properties of null
```

如果打开：

> Pause on exceptions

浏览器会直接停在：

```js
user.profile.name
```

这一行。

这叫：

> **现场抓捕。**

而不是：

> 事后分析。

---

# 二十一、Pause on caught exceptions

还有一个更深入的选项：

> Pause on caught exceptions

例如：

```js
try {
    JSON.parse(data)
} catch (error) {
    console.log(error)
}
```

如果错误被 catch 了：

```text
程序不会崩
```

但是你可能仍然想知道：

> 错误最初到底在哪里发生？

开启这个功能之后：

```text
JSON.parse()
 ↓
Exception
 ↓
立即暂停
```

非常适合排查：

* Promise
* API
* JSON
* 第三方库
* WebSocket
* 数据解析

---

# 二十二、Event Listener Breakpoints

这个是 Sources 里另一个重量级功能。

你可以告诉浏览器：

> 某种事件发生时，直接暂停。

例如：

```text
Event Listener Breakpoints

Mouse
    click
    dblclick
    mousedown

Keyboard
    keydown
    keyup

Control
    submit
    reset

Clipboard
    copy
    paste

Drag / Drop
    dragstart
    drop

Timer
    setTimeout
    setInterval

Animation
    animationstart
    animationend
```

---

# 二十三、它解决什么问题？

比如：

> 点击这个按钮到底执行了什么？

你不知道事件绑定在哪里。

可能是：

```js
button.onclick = ...
```

也可能：

```js
addEventListener(...)
```

也可能：

```vue
@click="handleClick"
```

甚至：

```text
第三方组件
 ↓
事件代理
 ↓
Vue
 ↓
你的组件
```

你可以开启：

```text
Mouse
 ☑ click
```

然后：

> 点击按钮。

浏览器直接暂停。

然后看：

```text
Call Stack
```

你就能找到：

> **这个 click 最终是谁处理的。**

---

# 二十四、Timer Breakpoints

比如：

```js
setTimeout(() => {
    update()
}, 3000)
```

或者：

```js
setInterval(() => {
    refresh()
}, 1000)
```

如果页面：

> 不知道为什么一直在执行某个逻辑。

可以使用：

> Timer Breakpoints

追踪：

```text
setTimeout
setInterval
```

特别适合排查：

* 页面卡顿
* 重复请求
* 定时刷新
* 内存泄漏
* 轮询
* 动画问题

---

# 二十五、XHR / Fetch Breakpoints

这个功能我非常推荐前端开发者熟练掌握。

你可以设置：

```text
XHR/fetch Breakpoints
```

例如：

```text
/api/user
```

那么任何请求 URL 包含：

```text
/api/user
```

程序就会暂停。

---

# 二十六、它比 Network 更适合回答一个问题

Network 擅长：

> **请求发生了什么？**

Sources 擅长：

> **是谁发起了这个请求？**

例如你发现：

```text
/api/order
```

突然发送了。

Network：

```text
POST /api/order
```

但是你不知道：

```text
是谁调用的？
```

设置：

```text
XHR/fetch breakpoint

/api/order
```

然后：

```text
请求发生
 ↓
Debugger暂停
 ↓
Call Stack
```

可能看到：

```text
submitOrder
handleSubmit
onClick
```

这就直接找到源头。

---

# 二十七、Promise / async 调试

现代前端大量代码都是：

```js
async function load() {
    const user = await getUser()
    const orders = await getOrders(user.id)

    return orders
}
```

Debug 时尤其值得观察：

```text
Call Stack
Scope
Async
```

你可以一步一步：

```text
load()
 ↓
await getUser()
 ↓
恢复执行
 ↓
await getOrders()
 ↓
恢复执行
```

理解：

> **异步程序实际上是怎么跑起来的。**

---

# 二十八、这是理解 Event Loop 的最好实践之一

例如：

```js
console.log('A')

setTimeout(() => {
    console.log('B')
}, 0)

Promise.resolve().then(() => {
    console.log('C')
})

console.log('D')
```

输出：

```text
A
D
C
B
```

不要只背：

> 微任务优先于宏任务。

直接在 Sources 里打断点，观察：

```text
同步执行
 ↓
Promise callback
 ↓
Timer callback
```

你会真正理解：

```text
Call Stack
Microtask Queue
Task Queue
```

之间的关系。

---

# 二十九、Debug Vue 组件

你经常使用 Vue 3 的话，Sources 非常值得结合 Vue DevTools 使用。

例如：

```vue
<script setup>
import { ref } from 'vue'

const count = ref(0)

function increment() {
    count.value++
}
</script>
```

在：

```js
count.value++
```

打断点。

点击按钮。

你可以看到：

```text
Scope
    count
```

并观察：

```text
count.value
```

变化。

---

# 三十、继续追 Vue 的响应式更新

例如：

```js
count.value++
```

如果你 Step Into，可能会进入：

```text
Vue runtime
 ↓
RefImpl
 ↓
trigger
 ↓
effect
 ↓
scheduler
 ↓
component update
```

这对于理解 Vue 响应式系统非常有帮助。

但这里有个原则：

> **不要一上来就钻进 Vue 源码。**

应该：

```text
自己的业务代码
 ↓
找到关键调用
 ↓
确认确实需要
 ↓
再进入 Vue Runtime
```

否则非常容易迷失。

---

# 三十一、Source Map：Sources 的灵魂

现代前端开发必须理解 Source Map。

比如你写：

```text
src/
  main.js
  App.vue
  components/
```

浏览器实际执行可能是：

```text
assets/index-Bx82d.js
```

为什么 Sources 还能看到：

```text
App.vue
```

因为：

```text
Source Map
```

它建立了：

```text
编译后代码
      ↕
原始源代码
```

之间的映射。

---

# 三十二、Source Map 出问题会发生什么？

你可能看到：

```text
assets/index-abc123.js
```

然后：

```text
100000 行压缩代码
```

而不是：

```text
src/components/User.vue
```

这时候调试体验会非常糟糕。

所以前端工程师需要理解：

```text
Vite
Webpack
Rollup
Source Map
```

之间的关系。

---

# 三十三、Pretty Print

如果你遇到：

```js
(()=>{var e=...;function t(){...}})()
```

可以使用：

> Pretty print `{}`

让代码变成：

```js
(() => {
    var e = ...

    function t() {
        ...
    }
})()
```

注意：

> Pretty Print 只是格式化，并不会真正恢复变量名和原始结构。

所以：

```text
Pretty Print ≠ Source Map
```

---

# 三十四、Blackbox：大型项目必学

如果你使用：

* Vue
* React
* Axios
* Element Plus
* Ant Design
* Lodash
* Three.js
* WebSocket libraries

Step Into 很容易进入第三方代码。

比如：

```text
your code
 ↓
axios
 ↓
interceptor
 ↓
promise
 ↓
...
```

这时候使用：

> Blackbox script

把第三方脚本排除掉。

于是：

```text
你的业务代码
 ↓
第三方库
 ↓
直接跳过去
 ↓
你的代码
```

这会让 Debug 体验发生巨大提升。

---

# 三十五、Sources + Network 是非常强的组合

很多前端 Bug 都应该：

```text
Network
+
Sources
```

一起使用。

例如：

> 页面加载用户信息失败。

先 Network：

```text
GET /api/user
Status 500
```

然后 Sources：

```text
XHR/fetch breakpoint
/api/user
```

再次请求。

程序停下来。

看：

```text
Call Stack
```

找到：

```js
loadUser()
```

再看：

```text
Scope
```

发现：

```js
userId = undefined
```

于是完整链路变成：

```text
页面初始化
 ↓
loadUser()
 ↓
userId undefined
 ↓
GET /api/user?id=undefined
 ↓
服务器 500
```

这比单独看 Network 快很多。

---

# 三十六、Sources + Elements

这也是非常经典的组合。

比如：

> 点击按钮之后页面样式发生异常。

Elements：

```text
DOM
 ↓
class
 ↓
style
```

Sources：

```text
click
 ↓
handler
 ↓
修改 class
```

最终你可以找到：

```js
element.classList.add('active')
```

是谁执行的。

---

# 三十七、Sources + Console

暂停之后，Console 其实变成了：

> **一个可以直接操作当前程序现场的 REPL。**

例如当前作用域有：

```js
const user = {
    name: 'Karl',
    age: 30
}
```

暂停以后，在 Console 可以直接：

```js
user
```

或者：

```js
user.name
```

甚至：

```js
user.age + 10
```

这非常强。

因为：

> **你不是在模拟程序，而是在程序真实暂停的现场操作它。**

---

# 三十八、debugger 语句

前端调试中我非常推荐：

```js
debugger
```

例如：

```js
function submitOrder(order) {
    debugger

    return api.submit(order)
}
```

打开 DevTools 后执行：

```text
submitOrder()
```

程序会自动暂停。

这比：

```js
console.log(order)
```

更加完整。

你可以直接：

```text
看 Scope
看 Call Stack
看 Watch
单步执行
修改变量
进入函数
退出函数
```

调试完删除：

```js
debugger
```

即可。

---

# 三十九、一个非常实用的调试技巧：运行时修改变量

程序暂停时，可以在 Console 中尝试修改当前可访问的变量。

例如：

```js
let total = 100
```

暂停后：

```js
total = 1000
```

然后继续运行。

这样你可以测试：

> 如果 total = 1000，后面的程序会发生什么？

这是一种非常有价值的：

> **运行时实验。**

而不是：

```text
修改源码
 ↓
保存
 ↓
重新编译
 ↓
刷新
 ↓
测试
```

---

# 四十、Live Edit / 修改代码后重新执行

在某些情况下，你还可以直接编辑 Sources 中的代码进行快速验证。

例如：

```js
if (user.isAdmin) {
```

临时改成：

```js
if (true) {
```

看看后续逻辑。

这种方式适合：

> 快速验证假设。

但不要误认为：

> 浏览器修改 = 项目源代码修改。

它只是当前运行环境中的调试实验。

---

# 四十一、Local Overrides：非常强大的功能

Sources 还有：

> Overrides

可以让你：

> **把网络加载的资源替换成你本地修改的版本。**

例如线上页面：

```text
https://example.com/app.js
```

你可以建立本地 Override。

然后：

```text
线上代码
 ↓
浏览器加载
 ↓
本地 Override
 ↓
使用你的修改版本
```

这对于：

* UI 调试
* CSS 修改
* JS 行为验证
* 线上 Bug 分析
* 第三方页面测试

非常有用。

---

# 四十二、Snippets：Sources 里的隐藏宝藏

Sources：

```text
Snippets
```

可以创建：

```js
const buttons = document.querySelectorAll('button')

buttons.forEach(button => {
    button.style.outline = '2px solid red'
})
```

然后：

> Run

直接在当前页面运行。

---

# 四十三、Snippets 非常适合做“调试脚本”

比如你经常需要：

```js
localStorage.clear()
location.reload()
```

可以做一个：

```js
function reset() {
    localStorage.clear()
    sessionStorage.clear()
    location.reload()
}

reset()
```

保存成：

```text
reset-page.js
```

以后直接：

```text
Sources
→ Snippets
→ Run
```

---

# 四十四、你甚至可以建立自己的 Debug Toolkit

例如：

```text
Snippets
├── clear-storage.js
├── inspect-vue.js
├── find-hidden-elements.js
├── highlight-elements.js
├── export-localstorage.js
├── websocket-monitor.js
└── performance-test.js
```

久而久之：

> Sources 不再只是“调试面板”，而会变成你的浏览器开发工具箱。

---

# 四十五、一个非常实用的完整调试流程

假设：

> 用户点击“提交订单”，但是偶尔失败。

不要直接：

```js
console.log()
```

推荐：

### 第一步

Elements：

确认按钮：

```text
DOM
class
disabled
```

---

### 第二步

Sources：

Event Listener Breakpoint：

```text
Mouse → click
```

---

### 第三步

点击按钮。

---

### 第四步

查看：

```text
Call Stack
```

例如：

```text
handleClick
 ↓
submitOrder
 ↓
createOrder
```

---

### 第五步

Step Into：

```js
submitOrder()
```

---

### 第六步

Scope：

观察：

```js
order
user
token
```

---

### 第七步

设置 Watch：

```js
order.items.length
```

```js
order.total
```

```js
user.id
```

---

### 第八步

XHR/fetch breakpoint：

```text
/api/order
```

---

### 第九步

程序停在：

```js
fetch('/api/order', ...)
```

---

### 第十步

检查：

```js
body
headers
token
```

---

### 第十一步

Network 查看：

```text
Request
Response
Status
Timing
```

最终得到完整链路：

```text
用户点击
   ↓
Vue click
   ↓
handleClick
   ↓
submitOrder
   ↓
参数构造错误
   ↓
fetch
   ↓
/api/order
   ↓
400
```

这才是真正的：

> **系统化 Debug。**

---

# 四十六、不同 Bug 应该使用什么 Sources 技巧？

| 问题          | 推荐功能                      |
| ----------- | ------------------------- |
| 函数为什么执行     | Call Stack                |
| 函数什么时候执行    | Breakpoint                |
| 某个参数为什么错    | Scope / Watch             |
| 某个用户才出错     | Conditional Breakpoint    |
| 点击按钮发生了什么   | Event Listener Breakpoint |
| 谁发起 API 请求  | XHR/Fetch Breakpoint      |
| 定时器不断执行     | Timer Breakpoint          |
| 页面莫名报错      | Exception Breakpoint      |
| Promise 出问题 | Async / Call Stack        |
| 进入第三方库太深    | Blackbox                  |
| Bundle 看不懂  | Source Map                |
| 压缩代码难看      | Pretty Print              |
| 想临时修改线上代码   | Overrides                 |
| 想保存调试脚本     | Snippets                  |
| 想直接实验变量     | Console + Scope           |
| 想快速停在指定位置   | `debugger`                |

---

# 四十七、进一步：把 Sources 和 JavaScript 原理对应起来

如果你真正想把 Sources 用到高手级别，建议把这些概念对应起来：

```text
Sources
│
├── Breakpoint
│      ↓
│   执行暂停
│
├── Scope
│      ↓
│   Lexical Environment
│
├── Call Stack
│      ↓
│   Execution Context
│
├── Closure
│      ↓
│   闭包
│
├── Async Stack
│      ↓
│   Event Loop / Promise
│
├── Event Listener
│      ↓
│   DOM Event System
│
├── XHR / Fetch
│      ↓
│   Network + JS Runtime
│
├── Source Map
│      ↓
│   编译 / Bundler
│
└── Blackbox
       ↓
    调试边界
```

这样你就不是在：

> “学习 DevTools 按钮”。

而是在：

> **通过 DevTools 观察 JavaScript Runtime。**

这是两个完全不同的层次。

---

# 四十八、如果你是 Vue 开发者，我建议重点练这套组合

你平时主要做 Vue 3 的话，我会把 Sources 学习路线压缩成：

```text
① debugger
     ↓
② Breakpoint
     ↓
③ Conditional Breakpoint
     ↓
④ Scope
     ↓
⑤ Watch
     ↓
⑥ Call Stack
     ↓
⑦ Step Over / Into / Out
     ↓
⑧ Event Listener Breakpoint
     ↓
⑨ XHR / Fetch Breakpoint
     ↓
⑩ Exception Breakpoint
     ↓
⑪ Async Stack
     ↓
⑫ Blackbox
     ↓
⑬ Source Map
     ↓
⑭ Overrides
     ↓
⑮ Snippets
```

掌握这一套之后，绝大多数前端 Bug 都可以从：

> “我不知道哪里有问题。”

变成：

> “我知道程序在哪里开始偏离预期。”

这就是 Debugging 能力真正的提升。

---

# 四十九、再往深一层：Sources 的核心其实只有三个问题

如果把整个 Sources 面板抽象掉，你会发现所有功能最终都在回答三个问题：

### ① 程序在哪里？

```text
Call Stack
Breakpoint
Source
```

↓

> **Where?**

---

### ② 程序为什么到了这里？

```text
Call Stack
Event Listener
XHR Breakpoint
Async Stack
```

↓

> **Why?**

---

### ③ 程序此刻是什么状态？

```text
Scope
Watch
Console
Closure
```

↓

> **What?**

所以真正成熟的 Debug 思维其实是：

```text
Where
 ↓
Why
 ↓
What
 ↓
Hypothesis
 ↓
Experiment
 ↓
Fix
```

而不是：

```text
报错
 ↓
console.log
 ↓
猜
 ↓
改
 ↓
刷新
 ↓
再猜
```

---

# 五十、我最推荐你做一个“Sources 实战训练”

可以专门做一个小型 Vue 3 项目，故意制造 **10～20 个 Bug**，然后只允许自己使用 DevTools 解决。

例如：

```text
01  点击事件失效
02  函数被调用两次
03  参数突然变 undefined
04  Promise 数据错乱
05  API 被重复请求
06  setInterval 没有停止
07  某个用户才出现 Bug
08  Vue 响应式没有更新
09  WebSocket 消息处理错误
10  页面异常报错被 catch
11  第三方组件触发了错误事件
12  某个请求不知道是谁发起的
13  async/await 执行顺序异常
14  闭包变量错误
15  线上 Bundle 无法阅读
16  某段代码只在特定条件执行
17  内存对象意外持续增长
18  页面某个按钮被莫名修改
```

然后分别使用：

```text
Breakpoint
Conditional Breakpoint
Logpoint
Call Stack
Scope
Watch
Event Listener Breakpoint
XHR/Fetch Breakpoint
Timer Breakpoint
Exception Breakpoint
Async Stack
Blackbox
Source Map
Overrides
Snippets
```

把这套训练做完，你对 Sources 的理解会从：

> **“会用断点”**

直接提升到：

> **“能够沿着运行时因果链定位 Bug”。**

而这实际上已经开始接近高级前端工程师的 Debugging 能力了。
