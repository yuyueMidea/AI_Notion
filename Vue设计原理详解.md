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

