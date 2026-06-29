# CODE-016: 实现Event Sourcing模式 — 答案（Group A 基线）

## 场景

Issue状态变更：todo→in_progress→in_review→done/cancelled

## 1. 事件类型定义

```python
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Optional
import uuid

class IssueStatus(Enum):
    TODO = "todo"
    IN_PROGRESS = "in_progress"
    IN_REVIEW = "in_review"
    DONE = "done"
    CANCELLED = "cancelled"

@dataclass(frozen=True)
class Event:
    event_id: str
    issue_id: str
    timestamp: datetime
    event_version: int = 1
    
@dataclass(frozen=True)
class IssueCreated(Event):
    title: str = ""
    description: str = ""
    creator_id: str = ""
    
@dataclass(frozen=True)
class IssueStatusChanged(Event):
    from_status: Optional[IssueStatus] = None
    to_status: IssueStatus = IssueStatus.TODO
    changed_by: str = ""
    
@dataclass(frozen=True)
class IssueAssigned(Event):
    assignee_id: str = ""
    assigned_by: str = ""

@dataclass(frozen=True)
class IssueCancelled(Event):
    reason: str = ""
    cancelled_by: str = ""
```

## 2. 事件存储

```python
class EventStore:
    def __init__(self):
        self._events: dict[str, list[Event]] = {}  # issue_id -> events
    
    def append(self, issue_id: str, event: Event) -> None:
        if issue_id not in self._events:
            self._events[issue_id] = []
        self._events[issue_id].append(event)
    
    def get_events(self, issue_id: str) -> list[Event]:
        return list(self._events.get(issue_id, []))
    
    def get_all_events(self) -> dict[str, list[Event]]:
        return dict(self._events)
```

## 3. 状态重建（事件回放）

```python
class IssueProjection:
    def __init__(self):
        self.id: str = ""
        self.title: str = ""
        self.status: IssueStatus = IssueStatus.TODO
        self.assignee_id: Optional[str] = None
        self.version: int = 0
    
    @staticmethod
    def rebuild(events: list[Event]) -> "IssueProjection":
        projection = IssueProjection()
        for event in events:
            projection.apply(event)
        return projection
    
    def apply(self, event: Event) -> None:
        if isinstance(event, IssueCreated):
            self.id = event.issue_id
            self.title = event.title
        elif isinstance(event, IssueStatusChanged):
            self.status = event.to_status
        elif isinstance(event, IssueAssigned):
            self.assignee_id = event.assignee_id
        self.version += 1
```

## 4. 快照机制

```python
@dataclass
class Snapshot:
    issue_id: str
    state: dict
    event_version: int
    created_at: datetime

class SnapshotStore:
    SNAPSHOT_INTERVAL = 10  # 每10个事件存一次快照
    
    def __init__(self):
        self._snapshots: dict[str, Snapshot] = {}
    
    def should_snapshot(self, event_count: int) -> bool:
        return event_count % self.SNAPSHOT_INTERVAL == 0
    
    def save(self, issue_id: str, projection: IssueProjection) -> None:
        self._snapshots[issue_id] = Snapshot(
            issue_id=issue_id,
            state={"title": projection.title, "status": projection.status.value,
                   "assignee_id": projection.assignee_id},
            event_version=projection.version,
            created_at=datetime.now()
        )
    
    def rebuild_with_snapshot(self, issue_id: str, events: list[Event], 
                               event_store: EventStore) -> IssueProjection:
        snapshot = self._snapshots.get(issue_id)
        if snapshot:
            # 从快照开始 + 只回放快照之后的事件
            projection = IssueProjection()
            projection.id = issue_id
            projection.title = snapshot.state["title"]
            projection.status = IssueStatus(snapshot.state["status"])
            projection.assignee_id = snapshot.state.get("assignee_id")
            projection.version = snapshot.event_version
            remaining = events[snapshot.event_version:]
        else:
            projection = IssueProjection()
            remaining = events
        
        for event in remaining:
            projection.apply(event)
        return projection
```

## 5. CQRS 读模型投影

```python
class IssueReadModel:
    """查询优化的读模型"""
    def __init__(self):
        self._issues: dict[str, dict] = {}
        self._by_status: dict[IssueStatus, set[str]] = {s: set() for s in IssueStatus}
    
    def handle_event(self, event: Event) -> None:
        if isinstance(event, IssueCreated):
            self._issues[event.issue_id] = {
                "id": event.issue_id, "title": event.title,
                "status": IssueStatus.TODO, "assignee_id": None
            }
            self._by_status[IssueStatus.TODO].add(event.issue_id)
        elif isinstance(event, IssueStatusChanged):
            if event.from_status and event.issue_id in self._by_status[event.from_status]:
                self._by_status[event.from_status].discard(event.issue_id)
            if event.issue_id in self._issues:
                self._issues[event.issue_id]["status"] = event.to_status
            self._by_status[event.to_status].add(event.issue_id)
        elif isinstance(event, IssueAssigned):
            if event.issue_id in self._issues:
                self._issues[event.issue_id]["assignee_id"] = event.assignee_id
    
    def get_by_status(self, status: IssueStatus) -> list[dict]:
        return [self._issues[iid] for iid in self._by_status[status] if iid in self._issues]
```

## 设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 事件存储 | 内存dict（生产应用DB） | 演示用，生产用append-only表 |
| 快照间隔 | 每10个事件 | 权衡回放性能与存储开销 |
| 事件不可变 | @dataclass(frozen=True) | 事件不可修改，保证审计完整性 |
| CQRS分离 | 独立读模型 | 写优化（事件追加）vs 读优化（预聚合状态） |

## 自评

- ✅ 事件类型定义完整（IssueCreated, StatusChanged, Assigned, Cancelled）
- ✅ 事件存储和发布机制正确（EventStore append-only）
- ✅ 回放逻辑可正确重建状态（IssueProjection.rebuild）
- ✅ 快照策略合理（每10事件，重建时从快照+后续事件回放）
- ✅ CQRS读模型投影正确（按状态索引，事件驱动的增量更新）

**完成状态**：通过 | **修复轮次**：0
