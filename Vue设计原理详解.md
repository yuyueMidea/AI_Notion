# Vue 的设计原理详解

Vue 的设计可以从"设计哲学"和"核心技术实现"两个层面来理解。下面从整体架构到具体实现细节展开聊。

## 一、整体设计哲学:渐进式框架

Vue 最核心的设计理念是**渐进式(Progressive)**——你可以只把它当一个模板引擎用,也可以搭配路由、状态管理、构建工具打造完整的工程化应用。这种设计体现在:

- **核心库只关注视图层**,不像 Angular 那样"大而全"
- **功能可插拔**:Vue Router、Vuex/Pinia、Vue CLI 都是独立的包
- **API 设计上做兼容取舍**:Vue3 保留 Options API 的同时引入 Composition API,让老项目平滑迁移

Vue3 的源码采用 **Monorepo** 架构,拆分为多个职责单一的包:

```
@vue/reactivity    响应式系统(可独立使用,不依赖 Vue)
@vue/runtime-core  平台无关的运行时核心(组件、diff、生命周期)
@vue/runtime-dom   浏览器平台特定实现(DOM API 操作)
@vue/compiler-core 平台无关的编译器
@vue/compiler-dom  针对浏览器的编译器
@vue/compiler-sfc  单文件组件(.vue)编译支持
```

这种分层解耦让 `runtime-core` 可以配合不同的 renderer 跑在小程序、Canvas 等非 DOM 平台上(这也是 Vue3 支持自定义渲染器的基础)。

## 二、响应式系统:Vue 的灵魂

这是 Vue 区别于其他框架最核心的部分。

### Vue2:Object.defineProperty

Vue2 通过 `Object.defineProperty` 劫持对象的 getter/setter:

```javascript
function defineReactive(obj, key, val) {
  const dep = new Dep() // 依赖收集器
  Object.defineProperty(obj, key, {
    get() {
      if (Dep.target) dep.depend() // 收集当前正在渲染的 Watcher
      return val
    },
    set(newVal) {
      if (newVal === val) return
      val = newVal
      dep.notify() // 通知所有依赖此数据的 Watcher 更新
    }
  })
}
```

核心是 **依赖收集(track)+ 派发更新(trigger)** 的模型:每个响应式数据对应一个 `Dep`,每个组件渲染时对应一个 `Watcher`,访问数据时收集依赖,修改数据时触发更新。

**局限性**:
- 无法监听对象新增/删除属性(需要 `Vue.set`/`Vue.delete`)
- 数组的索引赋值、`length` 修改无法被拦截(需要重写数组方法)
- 初始化时需要递归遍历所有属性,性能有损耗

### Vue3:Proxy + Reflect

Vue3 用 `Proxy` 重写了响应式系统,从根源上解决了上述问题:

```javascript
function reactive(target) {
  return new Proxy(target, {
    get(target, key, receiver) {
      track(target, key) // 依赖收集
      const result = Reflect.get(target, key, receiver)
      return typeof result === 'object' ? reactive(result) : result // 懒代理,访问到才递归
    },
    set(target, key, value, receiver) {
      const result = Reflect.set(target, key, value, receiver)
      trigger(target, key) // 触发更新
      return result
    },
    deleteProperty(target, key) {
      const result = Reflect.deleteProperty(target, key)
      trigger(target, key)
      return result
    }
  })
}
```

`track`/`trigger` 底层用 **WeakMap → Map → Set** 三层结构维护依赖关系:

```
targetMap (WeakMap)
  └── target对象 → depsMap (Map)
                     └── key → dep (Set,存放依赖此key的 effect)
```

配合 `effect` 函数(替代 Vue2 的 Watcher),实现了更细粒度、更灵活的响应式追踪,也支持了 `ref`、`computed`、`watch` 等 API。

**Proxy 的优势**:
- 可以拦截整个对象,新增/删除属性天然支持
- 数组的各种操作都能正确拦截
- 懒代理(访问到嵌套属性才递归转换),初始化性能更好

## 三、虚拟DOM与Diff算法

### 为什么需要虚拟DOM

直接操作真实 DOM 代价高(重排重绘)。虚拟DOM 用 JS 对象描述 DOM 结构:

```javascript
{
  tag: 'div',
  props: { class: 'container' },
  children: [{ tag: 'span', children: '内容' }]
}
```

状态变化时先在内存中对比新旧虚拟DOM(Diff),计算出最小的真实DOM操作再统一执行,减少直接操作真实DOM的次数。

### Diff 算法核心思路

Vue 采用**同层比较**(不跨层级比较,降低复杂度到 O(n)):

1. 只比较同一父节点下的子节点
2. 双端比较:同时从新旧两组节点的头尾开始比较,尽量复用节点、减少移动
3. `key` 的作用:帮助 Diff 算法识别节点是否可复用,避免"就地复用"导致的状态错乱(比如列表里带输入框的情况)

## 四、Vue3 编译时优化(这是很多人忽略但非常关键的部分)

Vue3 相比 Vue2 一个巨大的突破是**编译器和运行时的协同优化**,把一部分本该运行时做的工作提前到编译阶段完成。

### 1. 静态提升(Static Hoisting)

不会变化的节点在编译时被提升到 render 函数外,只创建一次:

```javascript
// 编译前
<div><span>静态文本</span>{{ dynamic }}</div>

// 编译后(伪代码)
const hoisted = createVNode('span', null, '静态文本') // 只创建一次
function render() {
  return createVNode('div', null, [hoisted, toDisplayString(dynamic)])
}
```

### 2. Patch Flag(补丁标记)

编译器给动态节点打上标记,告诉运行时"这个节点只有 class 会变"或"只有文本会变",Diff 时直接跳过静态属性的比较:

```javascript
createVNode('div', { class: dynamicClass }, text, 2 /* PatchFlags.CLASS */)
```

### 3. Block Tree

把动态节点收集到一个扁平数组中(而不是每层都递归遍历整棵树),更新时只需遍历这个"动态节点列表",跳过所有静态子树。

这三项优化让 Vue3 的 Diff 复杂度和更新性能相比 Vue2 有数量级的提升。

## 五、组件化设计

- **单向数据流**:props 向下传递,事件向上传递,保证数据流向可预测
- **生命周期钩子**:本质是在组件实例的特定阶段执行注册的回调数组
- **组件通信**:props/emit、provide/inject(依赖注入,解决跨层级传递)、插槽(slot,实现内容分发)

## 六、Composition API 的设计初衷

Options API(`data`、`methods`、`computed` 分开写)在复杂组件中会导致**同一逻辑关注点被拆散在不同选项里**,难以复用(只能靠 mixin,还有命名冲突、来源不清晰的问题)。

Composition API 让你按**逻辑关注点**组织代码:

```javascript
function useMouse() {
  const x = ref(0), y = ref(0)
  const update = (e) => { x.value = e.pageX; y.value = e.pageY }
  onMounted(() => window.addEventListener('mousemove', update))
  onUnmounted(() => window.removeEventListener('mousemove', update))
  return { x, y }
}
// 组件里直接复用整套逻辑
const { x, y } = useMouse()
```

本质上这是把 React Hooks 的"逻辑复用"思路,结合 Vue 自身的响应式系统实现的一套方案,但没有 React 那种"闭包陷阱"和依赖数组的心智负担,因为 Vue 的响应式是自动追踪依赖的。

---

## 继续深入 响应式系统的 `computed`/`watch` 具体实现

# Computed 与 Watch 的实现原理

这两者都建立在同一套底层设施——`ReactiveEffect` 之上,但解决的问题和设计取舍完全不同。先从共同基石讲起。

## 一、共同基石:ReactiveEffect

```javascript
class ReactiveEffect {
  constructor(fn, scheduler) {
    this.fn = fn
    this.scheduler = scheduler // 有scheduler时,依赖变化不直接重跑fn,而是走scheduler
    this.deps = [] // 反向记录:自己被哪些dep(Set)收集了,用于stop时清理
    this.active = true
  }
  
  run() {
    if (!this.active) return this.fn()
    
    const lastEffect = activeEffect
    activeEffect = this // 设置"当前正在收集依赖的effect"
    try {
      cleanupEffect(this) // 每次运行前先清空旧依赖,避免残留(处理条件分支导致的依赖变化)
      return this.fn()
    } finally {
      activeEffect = lastEffect // 用栈的方式还原,支持effect嵌套
    }
  }
  
  stop() {
    if (this.active) {
      cleanupEffect(this)
      this.active = false
    }
  }
}
```

**关键点**:普通渲染场景没有 `scheduler`,数据变化时 `trigger` 直接调用 `effect.run()` 重新渲染。而 `computed` 和 `watch` 都传入了自定义 `scheduler`——这是它们跟"自动重渲染"行为分道扬镳的开关。

## 二、computed 的实现:懒计算 + 双重依赖角色

computed 最精妙的地方在于它**同时扮演两个角色**:
- 对自己依赖的响应式数据来说,它是**订阅者(effect)**
- 对使用它的地方(比如 render 函数、别的 computed)来说,它是**发布者(有自己的 dep)**

```javascript
class ComputedRefImpl {
  constructor(getter, setter) {
    this._value = undefined
    this._dirty = true      // 脏标记:true表示需要重新计算
    this.dep = undefined    // 收集"谁依赖了这个computed"
    this.__v_isRef = true
    
    this.effect = new ReactiveEffect(getter, () => {
      // 这个scheduler不会在依赖变化时立刻重算!
      // 只做两件事:标脏 + 通知下游
      if (!this._dirty) {
        this._dirty = true
        triggerRefValue(this) // 通知依赖当前computed的effect(比如渲染函数)
      }
    })
    this.effect.computed = this
  }
  
  get value() {
    trackRefValue(this) // 如果是在别的effect中访问,把那个effect收集进this.dep
    if (this._dirty) {
      this._dirty = false
      this._value = this.effect.run() // 惰性求值:只有脏了才真正执行getter
    }
    return this._value
  }
  
  set value(newValue) {
    this.setter(newValue) // 支持writable computed
  }
}
```

### 为什么这样设计是关键优化

假设有 `count` → `double = computed(() => count.value * 2)` → 模板里用了 `double`。

**如果没有 dirty 标记**,每次 `count` 变化都要立刻重算 `double`,即使这一刻没人读取它,也白算了。

**Vue 的做法**是:`count` 变化 → 触发 `double.effect` 的 scheduler → **只是打个脏标记、通知渲染 effect 该更新了** → 真正的计算被推迟到渲染函数**真正读取 `double.value` 的那一刻**才发生。这是典型的"拉(pull)优于推(push)"思路,避免不必要的重复计算,也支持了 `double.value` 被连续读取多次时只算一遍(缓存)。

### effect 栈解决嵌套问题

如果一个 computed 依赖另一个 computed(`c2 = computed(() => c1.value + 1)`),访问 `c2.value` 时会触发 `c1.effect` 的执行,此时 `activeEffect` 需要正确地在 `c1` 和 `c2` 之间切换——这就是为什么 `run()` 里用 `lastEffect` 做"入栈/出栈"式还原,而不是简单赋值。

## 三、watch 的实现:effect + 调度时机的组合

`watch` 本质上是**一个自定义了 getter 和 scheduler 的 effect**,但比 computed 多了几个工程化的能力:数据源归一化、深度遍历、清理函数、刷新时机控制。

### 1. 数据源归一化(source → getter)

```javascript
function doWatch(source, cb, { deep, immediate, flush = 'pre' } = {}) {
  let getter
  
  if (isRef(source)) {
    getter = () => source.value
  } else if (isReactive(source)) {
    getter = () => source
    if (!deep) deep = true // 直接watch一个reactive对象,隐式开启深度监听
  } else if (isFunction(source)) {
    getter = source // watch(() => xxx.value, cb) 这种写法
  } else if (isArray(source)) {
    getter = () => source.map(s => isRef(s) ? s.value : s) // 支持watch多个源
  }
  
  if (cb && deep) {
    const baseGetter = getter
    getter = () => traverse(baseGetter()) // 深度遍历,强制访问所有嵌套属性完成依赖收集
  }
  ...
}
```

### 2. deep 的核心:traverse 强制触发嵌套依赖收集

```javascript
function traverse(value, seen = new Set()) {
  if (!isObject(value) || seen.has(value)) return value
  seen.add(value) // 防止循环引用死循环
  
  if (isRef(value)) {
    traverse(value.value, seen)
  } else if (isArray(value)) {
    for (let i = 0; i < value.length; i++) traverse(value[i], seen)
  } else if (isPlainObject(value)) {
    for (const key in value) traverse(value[key], seen) // 递归访问每个属性,触发get拦截器
  }
  return value
}
```

原理很直白:响应式系统只有在**属性被"访问(get)"** 时才会 track。浅层 watch 只访问了对象引用本身,不会递归进去。`traverse` 就是暴力地把对象每一层都"读一遍",借助 Proxy 的 get 拦截,把所有嵌套属性的依赖都收集到当前 watch 的 effect 里。

这也解释了为什么**深度 watch 大对象性能开销大**——本质上是一次完整的对象遍历。

### 3. cleanup(副作用清理)——解决竞态问题

```javascript
let cleanup
const onCleanup = (fn) => { cleanup = fn }

const job = () => {
  if (!effect.active) return
  if (cb) {
    const newValue = effect.run()
    if (deep || hasChanged(newValue, oldValue)) {
      if (cleanup) cleanup() // 先执行上一次注册的清理函数
      cb(newValue, oldValue, onCleanup) // 把onCleanup传给回调,供用户注册清理逻辑
      oldValue = newValue
    }
  } else {
    effect.run() // watchEffect场景,没有cb
  }
}
```

典型使用场景(解决竞态请求):

```javascript
watch(id, async (newId, oldId, onCleanup) => {
  let cancelled = false
  onCleanup(() => { cancelled = true }) // 若id再次变化,上次请求的结果会被标记作废
  
  const data = await fetchData(newId)
  if (!cancelled) result.value = data
})
```

### 4. flush 时机——scheduler 的真正用武之地

```javascript
const effect = new ReactiveEffect(getter, () => {
  if (flush === 'sync') {
    job() // 同步:数据变了立刻执行,不等待
  } else if (flush === 'post') {
    queuePostFlushCb(job) // 等组件DOM更新完成后再执行
  } else {
    queuePreFlushCb(job) // 默认pre:组件更新前执行(但仍是异步/微任务)
  }
})

if (cb) {
  oldValue = immediate ? INITIAL_WATCHER_VALUE : effect.run()
} else {
  effect.run() // watchEffect:构造时就立即run一次,建立依赖
}

if (immediate) job() // immediate: true 时,首次立刻执行一次回调
```

这里能看出 `watch` 和 `computed` 在调度策略上的本质区别:**computed 的 scheduler 只标脏、不执行**;**watch 的 scheduler 是把 job 塞进队列、异步批量执行**。

## 四、为什么需要异步队列(避免同步 watch 的性能陷阱)

如果每次数据变化都同步触发 watch 回调,同一个事件循环里连续修改 10 次响应式数据,回调会被触发 10 次。Vue 用微任务队列做去重合并:

```javascript
const queue = []
let isFlushPending = false
const resolvedPromise = Promise.resolve()

function queueJob(job) {
  if (!queue.includes(job)) { // 去重:同一个job在一个tick内只入队一次
    queue.push(job)
    queueFlush()
  }
}

function queueFlush() {
  if (!isFlushPending) {
    isFlushPending = true
    resolvedPromise.then(flushJobs) // 借助微任务,等同步代码全部执行完再统一处理
  }
}

function flushJobs() {
  queue.sort((a, b) => a.id - b.id) // 按组件id排序,保证父组件先于子组件更新
  for (const job of queue) job()
  queue.length = 0
  isFlushPending = false
}
```

这套机制跟 `nextTick` 是同一个队列体系——这也是为什么"修改数据后立刻读 DOM 拿不到最新值,要用 `nextTick`"的根本原因。

## 五、watch vs watchEffect 一览

| | watch | watchEffect |
|---|---|---|
| 依赖收集方式 | 显式指定 source | 自动追踪函数内访问的所有响应式数据 |
| 是否拿到 oldValue | 是 | 否 |
| 是否需要 deep | 需手动开启(或watch reactive对象隐式开启) | 天然是"深度"的,因为整个函数体都会执行 |
| 首次是否执行 | 默认不执行(除非 immediate: true) | 定义时立刻执行一次 |

本质上 `watchEffect` 就是 `doWatch` 内部 `cb` 为 `undefined` 的那条分支,复用了同一套 effect + 调度逻辑。
