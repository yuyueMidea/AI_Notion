好，这一部分我直接给你一套**可以落地用的「Claude 工作系统（Go + 前端开发版）」**，不是概念，是你可以**直接复制用 + 接入你项目**的。

---

# 🚀 一、整体设计目标（针对你）

结合你当前背景（前端 + Go + 系统设计 + 爬虫 + 低代码）：

👉 目标是让 Claude 变成：

* 后端架构师（Go）
* 前端工程助手（React/Vue）
* API 设计师
* 文档生成器
* Debug 助手
* 自动化任务执行器

---

# 📁 二、完整目录结构（可直接用）

```bash
Claude-Cowork/
├── ABOUT_ME/
│   └── about-me.md
│
├── RULES/
│   ├── anti-ai-style.md
│   ├── coding-style.md
│   └── tech-stack.md
│
├── TEMPLATES/
│   ├── task-backend.md
│   ├── task-frontend.md
│   ├── task-api.md
│   ├── task-debug.md
│   ├── task-refactor.md
│   └── task-system-design.md
│
├── PROJECTS/
│   └── your-project/
│       ├── brief.md
│       ├── api.md
│       ├── schema.sql
│       ├── frontend.md
│       └── tasks/
│
└── CLAUDE_OUTPUTS/
    ├── backend/
    ├── frontend/
    ├── api/
    ├── debug/
    └── docs/
```

---

# 🧠 三、核心文件（直接给你模板）

---

## ✅ 1. about-me.md（你的“AI人格配置”）

```md
# ABOUT ME

## Role
- Fullstack Developer (Frontend + Go Backend)
- Focus on system design, automation, and efficiency

## Tech Stack
- Frontend: Vue3 / React / JS / Tailwind
- Backend: Go (Gin / Fiber)
- DB: MySQL / SQLite
- Tools: Docker / Redis / Playwright

## Coding Preferences
- Clean architecture
- Modular design
- No over-engineering
- Prefer simple & readable code

## Output Expectations
- Always give runnable code
- Include file structure
- Explain key logic only (no fluff)
- Prefer practical over theoretical

## Current Focus
- Build scalable backend systems
- Improve dev efficiency with AI
- Build reusable components & systems
```

---

## ✅ 2. anti-ai-style.md（去AI味）

```md
# DO NOT USE

Avoid these words:
- delve
- moreover
- furthermore
- it's worth noting
- in conclusion

Avoid:
- long explanations
- generic advice
- vague wording

Use:
- direct answer
- code first
- practical steps
```

---

## ✅ 3. coding-style.md（统一代码风格）

```md
# CODING STYLE

## Go
- Use standard project layout
- Separate: handler / service / repo
- Return structured error

## Frontend
- Functional components
- Keep logic simple
- Avoid unnecessary state

## General
- Always include:
  - file structure
  - example usage
```

---

## ✅ 4. tech-stack.md（上下文约束）

```md
# TECH STACK

Backend:
- Go + Gin
- RESTful API

Frontend:
- Vue3 / React (no UI framework preferred)

Database:
- MySQL / SQLite

Other:
- Redis for cache
- Docker for deployment
```

---

# 🧩 四、任务模板（核心能力）

👉 以后你**不再写 prompt，只填模板**

---

## 🔧 1. 后端开发模板（task-backend.md）

```md
# TASK: BACKEND FEATURE

## Feature
{{功能描述}}

## Requirements
- Use Go + Gin
- RESTful API
- Clean architecture

## Input
{{请求参数}}

## Output
{{返回结构}}

## Constraints
- Must be runnable
- Include router + handler + service + repo

## Output Format
1. Project structure
2. Code
3. How to run
```

---

## 🎨 2. 前端模板（task-frontend.md）

```md
# TASK: FRONTEND COMPONENT

## Feature
{{功能}}

## Tech
- Vue3 / React

## Requirements
- No UI library OR specify if needed
- Clean UI
- Reusable

## Output
1. Component code
2. Usage example
3. State logic
```

---

## 🔌 3. API设计模板（task-api.md）

```md
# TASK: API DESIGN

## Feature
{{业务}}

## Requirements
- RESTful
- Clear naming
- Expandable

## Output
- API list
- Request/Response
- DB schema
```

---

## 🐞 4. Debug模板（task-debug.md）

```md
# TASK: DEBUG

## Problem
{{错误描述}}

## Code
{{代码}}

## Expect
- Find root cause
- Fix code
- Explain briefly
```

---

## 🔁 5. 重构模板（task-refactor.md）

```md
# TASK: REFACTOR

## Code
{{原代码}}

## Goals
- Simplify
- Improve readability
- Better structure

## Output
- New code
- What improved
```

---

## 🏗️ 6. 系统设计模板（task-system-design.md）

```md
# TASK: SYSTEM DESIGN

## Goal
{{系统目标}}

## Scale
{{用户量}}

## Requirements
- High performance
- Scalable

## Output
1. Architecture diagram (text)
2. Tech choices
3. Key modules
4. Data flow
```

---

# ⚙️ 五、使用方式（真实工作流）

---

## 🧪 场景1：你要做一个接口

👉 不再写 prompt，直接：

```md
# 使用 task-api.md

Feature: 用户登录系统（JWT）
```

Claude 会输出：

* API设计
* 表结构
* token方案

---

## 🧪 场景2：写后端代码

```md
# 使用 task-backend.md

Feature: 用户登录接口
```

👉 输出：

* 完整 Go 项目结构
* handler + service + repo

---

## 🧪 场景3：前端页面

```md
Feature: 登录页面 + 表单验证
```

👉 输出：

* Vue/React组件
* 表单逻辑

---

## 🧪 场景4：查bug

直接丢代码 + 报错
👉 Claude帮你定位 + 修复

---

# 🤖 六、进阶玩法（你会很需要）

---

## 🔥 1. 自动生成代码流

流程：

```
API设计 → 后端代码 → 前端页面 → 联调
```

👉 全部 Claude 自动串起来

---

## 🔥 2. 项目级上下文

把这些文件丢进 PROJECTS：

* api.md
* schema.sql
* frontend.md

👉 Claude 会理解整个项目

---

## 🔥 3. 自动日报 / 周报

```md
总结今天代码变更 + 输出文档
```

👉 自动生成开发日志

---

# 🧠 七、本质升级（非常关键）

你现在从：

❌ 写 prompt
➡️

升级为：

✅ **构建 AI 操作系统**

---
