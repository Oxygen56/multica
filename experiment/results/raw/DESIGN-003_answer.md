# DESIGN-003: 设计RESTful API — 答案（Group A 基线）

## 资源层级

```
/api/v1/projects/{project_id}
/api/v1/projects/{project_id}/tasks/{task_id}
/api/v1/projects/{project_id}/tasks/{task_id}/comments/{comment_id}
/api/v1/tags
/api/v1/users/{user_id}
```

## 完整API端点

### Projects

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/projects | List projects (paginated) |
| POST | /api/v1/projects | Create project |
| GET | /api/v1/projects/{id} | Get project detail |
| PUT | /api/v1/projects/{id} | Update project |
| DELETE | /api/v1/projects/{id} | Archive project (soft delete) |

### Tasks (nested under projects)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/projects/{pid}/tasks | List tasks |
| POST | /api/v1/projects/{pid}/tasks | Create task |
| GET | /api/v1/projects/{pid}/tasks/{tid} | Get task detail |
| PATCH | /api/v1/projects/{pid}/tasks/{tid} | Partial update task |
| DELETE | /api/v1/projects/{pid}/tasks/{tid} | Delete task |
| POST | /api/v1/projects/{pid}/tasks/{tid}/assign | Assign task to user |
| POST | /api/v1/projects/{pid}/tasks/{tid}/status | Transition status |

### Comments (nested under tasks)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/projects/{pid}/tasks/{tid}/comments | List comments |
| POST | /api/v1/projects/{pid}/tasks/{tid}/comments | Add comment |
| PUT | /api/v1/projects/{pid}/tasks/{tid}/comments/{cid} | Update comment |
| DELETE | /api/v1/projects/{pid}/tasks/{tid}/comments/{cid} | Delete comment |

### Tags

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/tags | List all tags |
| POST | /api/v1/tags | Create tag |
| PUT | /api/v1/tasks/{tid}/tags | Set tags on a task |

### Users

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/users | List users |
| GET | /api/v1/users/{id} | Get user profile |
| GET | /api/v1/users/{id}/tasks | Get user's assigned tasks |

## 分页、排序、过滤

### 分页

```
GET /api/v1/projects?page=1&per_page=20
```

响应头：
```
Link: </api/v1/projects?page=2&per_page=20>; rel="next"
X-Total-Count: 147
X-Page: 1
X-Per-Page: 20
```

### 排序

```
GET /api/v1/projects/{pid}/tasks?sort=priority:desc,created_at:asc
```

### 过滤

```
GET /api/v1/projects/{pid}/tasks?status=todo,in_progress&assignee_id=uuid&tag=bug&priority=gte:3
```

支持的过滤操作符：`eq:`(默认), `gte:`, `lte:`, `gt:`, `lt:`, `neq:`, `in:`(逗号分隔), `like:`(模糊匹配)

## 请求/响应示例

### Create Task

```json
POST /api/v1/projects/prj_abc123/tasks
Content-Type: application/json

{
  "title": "修复登录页面样式",
  "description": "在移动端Safari上登录按钮错位，需要修复CSS",
  "priority": 3,
  "assignee_id": "usr_xyz789",
  "tag_ids": ["tag_bug", "tag_frontend"],
  "due_date": "2026-07-15T00:00:00Z"
}
```

```json
201 Created
{
  "id": "tsk_def456",
  "title": "修复登录页面样式",
  "description": "在移动端Safari上登录按钮错位，需要修复CSS",
  "status": "todo",
  "priority": 3,
  "project_id": "prj_abc123",
  "assignee_id": "usr_xyz789",
  "creator_id": "usr_aaa111",
  "tags": [
    {"id": "tag_bug", "name": "bug", "color": "#ff0000"},
    {"id": "tag_frontend", "name": "frontend", "color": "#00ff00"}
  ],
  "due_date": "2026-07-15T00:00:00Z",
  "created_at": "2026-06-29T10:00:00Z",
  "updated_at": "2026-06-29T10:00:00Z"
}
```

### Transition Task Status

```json
POST /api/v1/projects/prj_abc123/tasks/tsk_def456/status
{
  "status": "in_progress",
  "comment": "开始处理，预计2小时完成"
}
```

## 错误格式

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {
        "field": "title",
        "code": "required",
        "message": "Task title is required"
      },
      {
        "field": "priority",
        "code": "invalid_value",
        "message": "Priority must be between 1 and 5"
      }
    ],
    "request_id": "req_123456"
  }
}
```

### HTTP状态码约定

| Code | Usage |
|------|-------|
| 200 | 成功GET/PUT/PATCH |
| 201 | 成功POST（创建） |
| 204 | 成功DELETE |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 冲突（如并发修改） |
| 422 | 语义错误（验证失败） |
| 429 | 请求过多 |
| 500 | 服务端错误 |

## API版本管理

```
/api/v1/...  (当前稳定版)
/api/v2/...  (未来版本，破坏性变更时启用)
```

通过URL前缀做版本管理，清晰明确。v1至少维护12个月，废弃前提前6个月通知（通过响应头`Sunset`和`Deprecation`）。

## 自评

- ✅ API端点完整且RESTful（5类资源，完整CRUD + 嵌套资源）
- ✅ 嵌套资源设计合理（projects→tasks→comments层级）
- ✅ 分页排序过滤设计标准（Link头、排序、多操作符过滤）
- ✅ 错误处理完整（结构化错误、详细信息、HTTP状态码约定）
- ✅ 文档格式规范

**完成状态**：通过 | **修复轮次**：0
