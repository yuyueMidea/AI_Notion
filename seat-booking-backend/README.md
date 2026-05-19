# Seat Booking Backend

一个面向“固定工位 + 可预订工位”的最小可用后端，使用：

- Go 标准库 `net/http`
- `database/sql`
- SQLite
- `modernc.org/sqlite` 纯 Go 驱动

## 1. 当前范围

第一版只做基础闭环：

1. 启动时根据 `config/seat_layout.json` 初始化工位数据
2. 查询工位列表及指定时间段的可用状态
3. 创建预订
4. 查询预订列表
5. 取消预订
6. 基础 CORS，便于本地前端页面调试

暂不包含：

- 登录鉴权
- 审批
- 消息通知
- 复杂组织架构
- 批量 Excel 导入
- 管理后台

## 2. 目录结构

```text
seat-booking-backend/
├── cmd/
│   └── server/
│       └── main.go
├── config/
│   └── seat_layout.json
├── data/
│   └── seat_booking.db          # 首次启动后自动生成
├── internal/
│   ├── config/
│   │   └── layout.go
│   ├── httpapi/
│   │   └── handlers.go
│   ├── model/
│   │   └── model.go
│   └── store/
│       └── store.go
├── go.mod
└── README.md
```

## 3. 启动方式

```bash
cd seat-booking-backend
go mod tidy
go run ./cmd/server
```

默认启动：

```text
http://127.0.0.1:8080
```

健康检查：

```bash
curl http://127.0.0.1:8080/api/health
```

## 4. 配置工位布局

修改：

```text
config/seat_layout.json
```

示例：

```json
{
  "seat_code_prefix": "A-",
  "seat_number_start": 1,
  "seat_number_width": 3,
  "zones": [
    {
      "zone_code": "A1",
      "zone_name": "西北工位区",
      "seat_count": 25,
      "fixed_count": 8
    }
  ],
  "fixed_owner_map": {
    "A-001": "张三"
  }
}
```

含义：

- `seat_count`：该区域总工位数
- `fixed_count`：该区域前 N 个工位初始化为固定工位
- `fixed_owner_map`：指定某个固定工位显示的员工姓名

注意：

- 工位基础数据只在数据库为空时自动初始化。
- 如果已经启动过并生成了 `data/seat_booking.db`，再修改 `seat_layout.json` 不会自动覆盖旧工位。
- 开发阶段要重新生成，可删除 `data/seat_booking.db` 后重启服务。

## 5. 时间格式

接口支持两种时间格式：

1. 前端 `datetime-local` 常用格式：
   ```text
   2026-05-20T09:00
   ```

2. RFC3339：
   ```text
   2026-05-20T09:00:00+08:00
   ```

没有时区的时间按 `Asia/Taipei` 解释。

## 6. API

### 6.1 查询工位

当前占用状态：

```bash
curl http://127.0.0.1:8080/api/seats
```

查询某个预订时间段内的状态：

```bash
curl "http://127.0.0.1:8080/api/seats?start_time=2026-05-20T09:00&end_time=2026-05-20T18:00"
```

返回工位 `availability`：

- `fixed`
- `available`
- `booked`
- `disabled`

### 6.2 创建预订

```bash
curl -X POST http://127.0.0.1:8080/api/reservations \
  -H "Content-Type: application/json" \
  -d '{
    "seat_code": "A-081",
    "booker_name": "王强",
    "start_time": "2026-05-20T09:00",
    "end_time": "2026-05-20T18:00",
    "note": "项目驻场"
  }'
```

如果该工位在这个时间段已被预订，会返回 409。

### 6.3 查询预订列表

查询全部：

```bash
curl http://127.0.0.1:8080/api/reservations
```

查询有效预订：

```bash
curl "http://127.0.0.1:8080/api/reservations?status=active"
```

按工位编号查：

```bash
curl "http://127.0.0.1:8080/api/reservations?seat_code=A-081"
```

按时间范围查重叠预订：

```bash
curl "http://127.0.0.1:8080/api/reservations?start_time=2026-05-20T09:00&end_time=2026-05-20T18:00"
```

### 6.4 取消预订

```bash
curl -X POST http://127.0.0.1:8080/api/reservations/1/cancel
```

## 7. 与前端衔接建议

你现有前端 Demo 目前是：

- 通过 `localStorage` 保存工位和预订
- 预订表单在浏览器本地直接改数据

下一步前端接后端时，最自然的替换方式是：

1. 页面初始化时调用 `GET /api/seats?...`
2. 点击“提交预订”时调用 `POST /api/reservations`
3. 点击“取消预订”时调用 `POST /api/reservations/{id}/cancel`
4. 提交或取消成功后重新请求 `GET /api/seats?...`

## 8. 当前后端设计原则

这版是“毛坯产品”：

- 工位基础数据先从 JSON 初始化
- 数据库只保留 `seats` 和 `reservations`
- 不提前引入审批、权限、消息推送等复杂模块
- 先把“能查、能订、能取消、能防冲突”跑通
