# DDD 重构 Phase 1 实施指南

> **开始日期**: 2026年3月25日  
> **目标**: 建立领域模型基础，完成核心聚合根设计  
> **时间**: 1 周（约 40 小时）  
> **输出**: 可编译、可测试的代码库

---

## 概览

### Phase 1 目标

1. ✅ 建立聚合根实体（Agent、ConversationSession、User、LLMProvider 等）
2. ✅ 定义值对象（AgentIdentity、ChannelConfig、ToolCallState 等）
3. ✅ 设计和实现领域事件
4. ✅ 定义仓储接口（标准 Repository 模式）
5. ✅ 建立基本的 DI 容器框架
6. ✅ 完成单元测试覆盖

### 工作量分解

| 任务 | 天数 | 依赖 | 所有者 |
|------|------|------|--------|
| 1. 领域模型设计评审 | 1 | - | `@lead` |
| 2. 聚合根实现 | 1.5 | 1 | `@dev1` |
| 3. 值对象与 VO 实现 | 1 | 2 | `@dev2` |
| 4. 领域事件定义 | 0.5 | 1 | `@dev1` |
| 5. 仓储接口设计 | 1 | 4 | `@dev2` |
| 6. 单元测试 | 1 | 2-5 | `@qa` |

### 文件清单

使用此清单来跟踪实施进度：

```
Phase 1 文件清单
├── domain/agent/
│   ├── entity.go                 ⬜ 未开始
│   ├── value_objects.go          ⬜ 未开始
│   ├── repository.go             ⬜ 未开始
│   ├── events.go                 ⬜ 未开始
│   └── entity_test.go            ⬜ 未开始
├── domain/session/
│   ├── conversation_session.go   ⬜ 未开始
│   ├── repository.go             ⬜ 未开始
│   ├── events.go                 ⬜ 未开始
│   └── entity_test.go            ⬜ 未开始
├── domain/message/
│   ├── conversation_record.go    ⬜ 未开始
│   ├── repository.go             ⬜ 未开始
│   └── events.go                 ⬜ 未开始
├── domain/user/
│   ├── entity.go                 ⬜ 未开始
│   ├── repository.go             ⬜ 未开始
│   └── events.go                 ⬜ 未开始
├── domain/llm_provider/
│   ├── entity.go                 ⬜ 未开始
│   └── repository.go             ⬜ 未开始
├── domain/events.go              ⬜ 未开始 (共享定义)
├── infrastructure/di/container.go ⬜ 未开始
└── README.md                      ⬜ 未开始
```

---

## 第 1 天：领域模型设计评审与代码架构

### 1.1 建立工作分支

```bash
# 在项目主目录执行
repo_name=$(basename "$(git rev-parse --show-toplevel)")
mkdir -p ../ai-worktrees/${repo_name}
git checkout -b feature/ddd-phase1-domain-model
git worktree add ../ai-worktrees/${repo_name}/ddd-phase1 feature/ddd-phase1-domain-model
cd ../ai-worktrees/${repo_name}/ddd-phase1
```

### 1.2 初始化目录结构

```bash
# 创建必要的目录
mkdir -p internal/domain/{agent,session,message,channel,user,llm_provider,mcp_server,shared}
mkdir -p internal/application
mkdir -p internal/infrastructure/{persistence,eventbus,middleware,sandbox,adapters}
mkdir -p internal/di

# 创建占位符文件
touch internal/domain/{agent,session,message,channel,user,llm_provider,mcp_server}/README.md
```

### 1.3 设置环境

```bash
# 构建项目
make setup
make build

# 如果有错误，检查 go.mod 指向
```

### 1.4 代码审查清单

在开始编码前，评审以下关键设计决定：

- [ ] 聚合根识别是否准确？
- [ ] 值对象的不可变性是否能保证？
- [ ] 实体的身份（ID）生成策略是什么？
- [ ] 事件溯源是否需要完整实现？
- [ ] 乐观锁 vs 悲观锁 for 并发控制？

**默认决定**：
- ✅ 使用 UUID v4 作为聚合根 ID
- ✅ 值对象通过名值对象在实体中表现
- ✅ 暂不全量实现 Event Sourcing（只有事件发布）
- ✅ 乐观锁配合 Version 字段

---

## 第 2-3 天：聚合根实现

### 2.1 Agent 聚合根实现

**文件**: `internal/domain/agent/entity.go`

```go
package agent

import (
    "context"
    "errors"
    "fmt"
    "time"
    
    "github.com/google/uuid"
)

// Agent ID 类型
type AgentID string

func (id AgentID) String() string {
    return string(id)
}

func NewAgentID() AgentID {
    return AgentID(uuid.New().String())
}

// 聚合根：Agent
type Agent struct {
    // 标识
    id       AgentID
    userCode string // 多租户隔离键
    version  int64  // 乐观锁版本
    
    // 核心属性
    identity     *Identity          // 值对象
    capabilities *CapabilitiesSet   // 值对象
    personality  *PersonalityConfig // 值对象
    
    // 元数据
    createdAt time.Time
    updatedAt time.Time
    
    // 事件
    uncommittedEvents []interface{}
}

// 构造函数
func NewAgent(userCode string, identity *Identity) *Agent {
    agent := &Agent{
        id:                NewAgentID(),
        userCode:          userCode,
        version:           1,
        identity:          identity,
        capabilities:      NewCapabilitiesSet(),
        personality:       NewDefaultPersonalityConfig(),
        createdAt:         time.Now(),
        updatedAt:         time.Now(),
        uncommittedEvents: make([]interface{}, 0),
    }
    
    agent.RaiseDomainEvent(&AgentCreatedEvent{
        AgentID:   agent.id.String(),
        UserCode:  userCode,
        Identity:  identity,
        CreatedAt: agent.createdAt,
    })
    
    return agent
}

// 聚合根 ID 访问器
func (a *Agent) ID() AgentID {
    return a.id
}

func (a *Agent) UserCode() string {
    return a.userCode
}

func (a *Agent) Version() int64 {
    return a.version
}

// 业务规则：添加能力
func (a *Agent) AddCapability(cap *Capability) error {
    if cap == nil {
        return ErrInvalidCapability
    }
    
    if a.capabilities.Contains(cap.Name) {
        return fmt.Errorf("capability %s already exists", cap.Name)
    }
    
    a.capabilities.Add(cap)
    a.version++
    
    a.RaiseDomainEvent(&CapabilityAddedEvent{
        AgentID:    a.id.String(),
        Capability: cap,
        AddedAt:    time.Now(),
    })
    
    return nil
}

// 业务规则：移除能力
func (a *Agent) RemoveCapability(capName string) error {
    if !a.capabilities.Contains(capName) {
        return fmt.Errorf("capability %s not found", capName)
    }
    
    a.capabilities.Remove(capName)
    a.version++
    
    a.RaiseDomainEvent(&CapabilityRemovedEvent{
        AgentID:    a.id.String(),
        Capability: capName,
        RemovedAt:  time.Now(),
    })
    
    return nil
}

// 业务规则：更新个性配置
func (a *Agent) UpdatePersonality(config *PersonalityConfig) error {
    if config == nil {
        return ErrInvalidConfig
    }
    
    old := a.personality
    a.personality = config
    a.version++
    
    a.RaiseDomainEvent(&PersonalityUpdatedEvent{
        AgentID:   a.id.String(),
        OldConfig: old,
        NewConfig: config,
        UpdatedAt: time.Now(),
    })
    
    return nil
}

// 查询方法
func (a *Agent) GetCapabilities() []*Capability {
    return a.capabilities.All()
}

func (a *Agent) HasCapability(name string) bool {
    return a.capabilities.Contains(name)
}

func (a *Agent) GetIdentity() *Identity {
    return a.identity
}

func (a *Agent) GetPersonality() *PersonalityConfig {
    return a.personality
}

// 事件相关
func (a *Agent) RaiseDomainEvent(event interface{}) {
    a.uncommittedEvents = append(a.uncommittedEvents, event)
}

func (a *Agent) GetUncommittedEvents() []interface{} {
    return a.uncommittedEvents
}

func (a *Agent) ClearUncommittedEvents() {
    a.uncommittedEvents = make([]interface{}, 0)
}

// 验证聚合根
func (a *Agent) Validate() error {
    if a.id == "" {
        return errors.New("agent id cannot be empty")
    }
    if a.userCode == "" {
        return errors.New("user code cannot be empty")
    }
    if a.identity == nil {
        return errors.New("agent identity cannot be nil")
    }
    return a.identity.Validate()
}

// 错误定义
var (
    ErrInvalidCapability = errors.New("invalid capability")
    ErrInvalidConfig     = errors.New("invalid config")
)
```

**测试**: `internal/domain/agent/entity_test.go`

```go
package agent_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/weibaohui/nanobot-go/internal/domain/agent"
)

func TestNewAgent(t *testing.T) {
    identity := agent.NewIdentity(
        "TestAgent",
        "A test AI agent",
        "assistant",
    )
    
    ag := agent.NewAgent("user-1", identity)
    
    assert.NotNil(t, ag)
    assert.Equal(t, "user-1", ag.UserCode())
    assert.NotEmpty(t, ag.ID().String())
    
    // 验证事件被发布
    events := ag.GetUncommittedEvents()
    assert.Equal(t, 1, len(events))
    assert.IsType(t, (*agent.AgentCreatedEvent)(nil), events[0])
}

func TestAgentAddCapability(t *testing.T) {
    identity := agent.NewIdentity("TestAgent", "Test", "assistant")
    ag := agent.NewAgent("user-1", identity)
    
    cap := &agent.Capability{
        Name:        "search",
        Type:        agent.ToolCapabilityType,
        Description: "Search capability",
    }
    
    err := ag.AddCapability(cap)
    assert.NoError(t, err)
    assert.True(t, ag.HasCapability("search"))
    
    // 验证版本增加
    assert.Equal(t, int64(2), ag.Version())
    
    // 验证事件
    events := ag.GetUncommittedEvents()
    // 两个事件：AgentCreatedEvent + CapabilityAddedEvent
    assert.Equal(t, 2, len(events))
}

func TestAgentAddDuplicateCapability(t *testing.T) {
    identity := agent.NewIdentity("TestAgent", "Test", "assistant")
    ag := agent.NewAgent("user-1", identity)
    
    cap := &agent.Capability{
        Name:        "search",
        Type:        agent.ToolCapabilityType,
        Description: "Search capability",
    }
    
    ag.AddCapability(cap)
    err := ag.AddCapability(cap)
    
    assert.Error(t, err)
    assert.Equal(t, int64(2), ag.Version()) // 版本未增加
}
```

### 2.2 ConversationSession 聚合根

**文件**: `internal/domain/session/conversation_session.go`

```go
package session

import (
    "errors"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "github.com/weibaohui/nanobot-go/internal/domain/message"
)

type SessionID string

func (id SessionID) String() string {
    return string(id)
}

func NewSessionID() SessionID {
    return SessionID(uuid.New().String())
}

// 会话状态枚举
type SessionState string

const (
    StateActive   SessionState = "active"
    StateArchived SessionState = "archived"
    StateFailed   SessionState = "failed"
)

// 聚合根：ConversationSession
type ConversationSession struct {
    // 标识
    id       SessionID
    userCode string
    agentID  string
    version  int64
    
    // 状态
    state SessionState
    
    // 消息历史（仅引用，实际存储在 Message 聚合中）
    messageIDs []string
    
    // 线程数据（虚拟文件系统状态）
    threadData *ThreadData
    
    // 工具调用状态
    pendingToolCalls map[string]*PendingToolCall
    
    // 元数据
    createdAt time.Time
    updatedAt time.Time
    
    // 事件
    uncommittedEvents []interface{}
}

// 构造函数
func NewConversationSession(userCode, agentID string) *ConversationSession {
    session := &ConversationSession{
        id:                NewSessionID(),
        userCode:          userCode,
        agentID:           agentID,
        version:           1,
        state:             StateActive,
        messageIDs:        make([]string, 0),
        threadData:        NewThreadData(userCode),
        pendingToolCalls: make(map[string]*PendingToolCall),
        createdAt:         time.Now(),
        updatedAt:         time.Now(),
        uncommittedEvents: make([]interface{}, 0),
    }
    
    session.RaiseDomainEvent(&SessionCreatedEvent{
        SessionID: session.id.String(),
        UserCode:  userCode,
        AgentID:   agentID,
        CreatedAt: session.createdAt,
    })
    
    return session
}

// 访问器
func (s *ConversationSession) ID() SessionID {
    return s.id
}

func (s *ConversationSession) UserCode() string {
    return s.userCode
}

func (s *ConversationSession) AgentID() string {
    return s.agentID
}

func (s *ConversationSession) State() SessionState {
    return s.state
}

func (s *ConversationSession) Version() int64 {
    return s.version
}

func (s *ConversationSession) MessageIDs() []string {
    return s.messageIDs
}

func (s *ConversationSession) ThreadData() *ThreadData {
    return s.threadData
}

// 业务规则：追加消息
func (s *ConversationSession) AppendMessageID(msgID string) error {
    if s.state != StateActive {
        return ErrSessionNotActive
    }
    
    if msgID == "" {
        return ErrInvalidMessageID
    }
    
    s.messageIDs = append(s.messageIDs, msgID)
    s.version++
    s.updatedAt = time.Now()
    
    s.RaiseDomainEvent(&MessageAppendedEvent{
        SessionID:  s.id.String(),
        MessageID:  msgID,
        AppendedAt: s.updatedAt,
    })
    
    return nil
}

// 业务规则：记录工具调用
func (s *ConversationSession) RegisterToolCall(call *PendingToolCall) error {
    if s.state != StateActive {
        return ErrSessionNotActive
    }
    
    if call.ID == "" || call.Name == "" {
        return ErrInvalidToolCall
    }
    
    s.pendingToolCalls[call.ID] = call
    s.version++
    
    s.RaiseDomainEvent(&ToolCallRegisteredEvent{
        SessionID: s.id.String(),
        ToolCall:  call,
        RegisteredAt: time.Now(),
    })
    
    return nil
}

// 业务规则：解决工具调用
func (s *ConversationSession) ResolveToolCall(toolCallID string, result interface{}, isError bool) error {
    if s.state != StateActive {
        return ErrSessionNotActive
    }
    
    call, ok := s.pendingToolCalls[toolCallID]
    if !ok {
        return fmt.Errorf("tool call %s not found", toolCallID)
    }
    
    call.Result = result
    call.Error = isError
    call.ResolvedAt = time.Now()
    s.version++
    
    s.RaiseDomainEvent(&ToolCallResolvedEvent{
        SessionID:   s.id.String(),
        ToolCallID:  toolCallID,
        Result:      result,
        IsError:     isError,
        ResolvedAt:  call.ResolvedAt,
    })
    
    return nil
}

// 业务规则：存档会话
func (s *ConversationSession) Archive() error {
    if s.state == StateArchived {
        return ErrSessionAlreadyArchived
    }
    
    s.state = StateArchived
    s.version++
    s.updatedAt = time.Now()
    
    s.RaiseDomainEvent(&SessionArchivedEvent{
        SessionID:  s.id.String(),
        ArchivedAt: s.updatedAt,
    })
    
    return nil
}

// 业务规则：关闭会话
func (s *ConversationSession) Close(reason string) error {
    if s.state != StateActive {
        return ErrSessionNotActive
    }
    
    s.state = StateFailed
    s.version++
    s.updatedAt = time.Now()
    
    s.RaiseDomainEvent(&SessionClosedEvent{
        SessionID: s.id.String(),
        Reason:    reason,
        ClosedAt:  s.updatedAt,
    })
    
    return nil
}

// 查询方法
func (s *ConversationSession) GetPendingToolCall(id string) *PendingToolCall {
    return s.pendingToolCalls[id]
}

func (s *ConversationSession) GetAllPendingToolCalls() []*PendingToolCall {
    result := make([]*PendingToolCall, 0, len(s.pendingToolCalls))
    for _, call := range s.pendingToolCalls {
        result = append(result, call)
    }
    return result
}

// 事件相关
func (s *ConversationSession) RaiseDomainEvent(event interface{}) {
    s.uncommittedEvents = append(s.uncommittedEvents, event)
}

func (s *ConversationSession) GetUncommittedEvents() []interface{} {
    return s.uncommittedEvents
}

func (s *ConversationSession) ClearUncommittedEvents() {
    s.uncommittedEvents = make([]interface{}, 0)
}

// 验证
func (s *ConversationSession) Validate() error {
    if s.id == "" {
        return errors.New("session id cannot be empty")
    }
    if s.userCode == "" {
        return errors.New("user code cannot be empty")
    }
    if s.agentID == "" {
        return errors.New("agent id cannot be empty")
    }
    if !isValidState(s.state) {
        return fmt.Errorf("invalid session state: %s", s.state)
    }
    return nil
}

func isValidState(state SessionState) bool {
    return state == StateActive || state == StateArchived || state == StateFailed
}

// 错误定义
var (
    ErrSessionNotActive      = errors.New("session is not active")
    ErrInvalidMessageID      = errors.New("invalid message id")
    ErrInvalidToolCall       = errors.New("invalid tool call")
    ErrSessionAlreadyArchived = errors.New("session already archived")
)

// 值对象：ThreadData
type ThreadData struct {
    userCode   string
    sessionID  string
    workspace  string
    uploads    string
    outputs    string
    metadata   map[string]interface{}
}

func NewThreadData(userCode string) *ThreadData {
    sessionID := uuid.New().String()
    return &ThreadData{
        userCode:  userCode,
        sessionID: sessionID,
        workspace: fmt.Sprintf("/mnt/user-data/workspace/%s", sessionID),
        uploads:   fmt.Sprintf("/mnt/user-data/uploads/%s", sessionID),
        outputs:   fmt.Sprintf("/mnt/user-data/outputs/%s", sessionID),
        metadata:  make(map[string]interface{}),
    }
}

func (td *ThreadData) GetWorkspacePath() string {
    return td.workspace
}

func (td *ThreadData) GetUploadsPath() string {
    return td.uploads
}

func (td *ThreadData) GetOutputsPath() string {
    return td.outputs
}

// 值对象：PendingToolCall
type PendingToolCall struct {
    ID         string
    Name       string
    Arguments  map[string]interface{}
    Result     interface{}
    Error      bool
    RegisteredAt time.Time
    ResolvedAt   time.Time
}
```

---

## 第 4 天：值对象与聚合设计

### 4.1 值对象定义

**文件**: `internal/domain/agent/value_objects.go`

```go
package agent

import (
    "errors"
    "fmt"
)

// 值对象：Identity
type Identity struct {
    name        string
    description string
    role        string
}

func NewIdentity(name, description, role string) *Identity {
    return &Identity{
        name:        name,
        description: description,
        role:        role,
    }
}

func (i *Identity) Name() string {
    return i.name
}

func (i *Identity) Description() string {
    return i.description
}

func (i *Identity) Role() string {
    return i.role
}

func (i *Identity) Validate() error {
    if i.name == "" {
        return errors.New("identity name cannot be empty")
    }
    return nil
}

// Equals 方法（值对象的关键）
func (i *Identity) Equals(other *Identity) bool {
    if other == nil {
        return false
    }
    return i.name == other.name &&
        i.description == other.description &&
        i.role == other.role
}

// 值对象：Capability
type CapabilityType string

const (
    ToolCapabilityType   CapabilityType = "tool"
    SkillCapabilityType  CapabilityType = "skill"
    ModelCapabilityType  CapabilityType = "model"
)

type Capability struct {
    Name        string
    Type        CapabilityType
    Description string
    Config      map[string]interface{}
}

func (c *Capability) Validate() error {
    if c.Name == "" {
        return errors.New("capability name cannot be empty")
    }
    if c.Type == "" {
        return fmt.Errorf("invalid capability type: %s", c.Type)
    }
    return nil
}

func (c *Capability) Equals(other *Capability) bool {
    if other == nil {
        return false
    }
    return c.Name == other.Name && c.Type == other.Type
}

// 值对象：CapabilitiesSet（集合值对象）
type CapabilitiesSet struct {
    items map[string]*Capability
}

func NewCapabilitiesSet() *CapabilitiesSet {
    return &CapabilitiesSet{
        items: make(map[string]*Capability),
    }
}

func (cs *CapabilitiesSet) Add(cap *Capability) {
    cs.items[cap.Name] = cap
}

func (cs *CapabilitiesSet) Remove(name string) {
    delete(cs.items, name)
}

func (cs *CapabilitiesSet) Contains(name string) bool {
    _, exists := cs.items[name]
    return exists
}

func (cs *CapabilitiesSet) All() []*Capability {
    result := make([]*Capability, 0, len(cs.items))
    for _, cap := range cs.items {
        result = append(result, cap)
    }
    return result
}

// 值对象：PersonalityConfig
type PersonalityConfig struct {
    ThinkingStyle  string
    ResponseStyle  string
    Tone           string
    MaxTokens      int
    Temperature    float64
    CustomSettings map[string]interface{}
}

func NewDefaultPersonalityConfig() *PersonalityConfig {
    return &PersonalityConfig{
        ThinkingStyle: "analytical",
        ResponseStyle: "concise",
        Tone:          "professional",
        MaxTokens:     2000,
        Temperature:   0.7,
        CustomSettings: make(map[string]interface{}),
    }
}

func (pc *PersonalityConfig) Validate() error {
    if pc.Temperature < 0 || pc.Temperature > 2 {
        return errors.New("temperature must be between 0 and 2")
    }
    if pc.MaxTokens <= 0 {
        return errors.New("max tokens must be positive")
    }
    return nil
}

func (pc *PersonalityConfig) Equals(other *PersonalityConfig) bool {
    if other == nil {
        return false
    }
    return pc.ThinkingStyle == other.ThinkingStyle &&
        pc.ResponseStyle == other.ResponseStyle &&
        pc.Tone == other.Tone &&
        pc.Temperature == other.Temperature
}
```

---

## 第 5 天：领域事件定义

**文件**: `internal/domain/events.go`

```go
package domain

import "time"

// DomainEvent 基础接口
type DomainEvent interface {
    EventType() string
    OccurredAt() time.Time
    AggregateID() string
    AggregateType() string
}

// 基础事件

type BaseDomainEvent struct {
    EventTypeStr   string
    OccurredAtTime time.Time
    AggregateIDVal string
    AggregateTypeVal string
}

func (e *BaseDomainEvent) EventType() string {
    return e.EventTypeStr
}

func (e *BaseDomainEvent) OccurredAt() time.Time {
    return e.OccurredAtTime
}

func (e *BaseDomainEvent) AggregateID() string {
    return e.AggregateIDVal
}

func (e *BaseDomainEvent) AggregateType() string {
    return e.AggregateTypeVal
}

// ========== Agent 事件 ==========

type AgentCreatedEvent struct {
    BaseDomainEvent
    UserCode string
    Identity interface{} // JSON
}

type CapabilityAddedEvent struct {
    BaseDomainEvent
    Capability interface{} // JSON
}

type CapabilityRemovedEvent struct {
    BaseDomainEvent
    Capability string
}

type PersonalityUpdatedEvent struct {
    BaseDomainEvent
    OldConfig interface{} // JSON
    NewConfig interface{} // JSON
}

// ========== Session 事件 ==========

type SessionCreatedEvent struct {
    BaseDomainEvent
    UserCode string
    AgentID  string
}

type MessageAppendedEvent struct {
    BaseDomainEvent
    MessageID string
    AppendedAt time.Time
}

type ToolCallRegisteredEvent struct {
    BaseDomainEvent
    ToolCall interface{} // JSON
    RegisteredAt time.Time
}

type ToolCallResolvedEvent struct {
    BaseDomainEvent
    ToolCallID string
    Result     interface{}
    IsError    bool
    ResolvedAt time.Time
}

type SessionArchivedEvent struct {
    BaseDomainEvent
    ArchivedAt time.Time
}

type SessionClosedEvent struct {
    BaseDomainEvent
    Reason   string
    ClosedAt time.Time
}

// ========== Message 事件 ==========

type ConversationRecordCreatedEvent struct {
    BaseDomainEvent
    SessionID string
    Role      string
    Content   interface{} // JSON
}
```

---

## 第 6 天-7 天：仓储接口与基础设施

### 6.1 仓储接口

**文件**: `internal/domain/agent/repository.go`

```go
package agent

import (
    "context"
)

type AgentRepository interface {
    // 基础CRUD
    FindByID(ctx context.Context, id AgentID) (*Agent, error)
    Save(ctx context.Context, agent *Agent) error
    Update(ctx context.Context, agent *Agent) error
    Delete(ctx context.Context, id AgentID) error
    
    // 查询
    FindByUserCode(ctx context.Context, userCode string) ([]*Agent, error)
    FindAll(ctx context.Context) ([]*Agent, error)
}
```

### 6.2 简单事件总线

**文件**: `internal/infrastructure/eventbus/event_bus.go`

```go
package eventbus

import (
    "context"
    "sync"
)

type EventHandler = func(context.Context, interface{}) error

type SimpleEventBus struct {
    subscribers map[string][]EventHandler
    mu          sync.RWMutex
}

func NewSimpleEventBus() *SimpleEventBus {
    return &SimpleEventBus{
        subscribers: make(map[string][]EventHandler),
    }
}

func (eb *SimpleEventBus) Subscribe(eventType string, handler EventHandler) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

func (eb *SimpleEventBus) Publish(ctx context.Context, event interface{}) error {
    eb.mu.RLock()
    handlers := eb.subscribers["*"] // 通用处理器
    eb.mu.RUnlock()
    
    for _, handler := range handlers {
        _ = handler(ctx, event) // 忽略错误以继续处理
    }
    
    return nil
}
```

### 6.3 依赖注入容器（第一版）

**文件**: `internal/di/container.go`

```go
package di

import (
    "github.com/weibaohui/nanobot-go/internal/infrastructure/eventbus"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type Container struct {
    Logger   *zap.Logger
    DB       *gorm.DB
    EventBus *eventbus.SimpleEventBus
}

func NewContainer(db *gorm.DB, logger *zap.Logger) *Container {
    return &Container{
        Logger:   logger,
        DB:       db,
        EventBus: eventbus.NewSimpleEventBus(),
    }
}
```

---

## 验收标准

### 代码质量

- ✅ 所有公共函数都有文档注释
- ✅ 单元测试覆盖率 ≥ 80%
- ✅ 代码通过 `go fmt`、`go vet`、`golangci-lint`
- ✅ 无 TODO 或 FIXME 未来处理

### 功能验收

- ✅ Agent 聚合根可正常创建、验证、发布事件
- ✅ ConversationSession 聚合根支持所有业务规则
- ✅ 所有值对象都是不可变的
- ✅ 仓储接口定义清晰，可被实现
- ✅ 事件总线实现可运行（基础版）

### 文档

- ✅ 添加 `internal/domain/README.md`，说明领域模型
- ✅ 添加 `internal/domain/agent/README.md`，说明 Agent 上下文
- ✅ 添加 `internal/domain/session/README.md`，说明 Session 上下文

---

## 检查清单

完成 Phase 1 之前，确保检查以下项目：

```
代码实现
□ Agent 聚合根完成
□ ConversationSession 聚合根完成
□ User 聚合根完成 (简化版)
□ LLMProvider 聚合根完成 (简化版)
□ 所有值对象实现
□ 领域事件定义

单元测试
□ Agent 聚合根测试
□ ConversationSession 聚合根测试
□ 事件发布测试
□ 值对象不可变性测试

集成
□ 仓储接口清晰
□ 事件总线可运行
□ DI 容器基础框架

文档
□ 领域模型设计文档
□ 各聚合根使用指南
□ API 文档

提交
□ 代码通过编译
□ 测试全部通过
□ 没有遗留的 TODO
□ 准备 PR 说明
```

---

## 提交 PR

完成以上所有工作后，提交 PR：

```bash
# 在 worktree 中
gh pr create \
  --title "feat: DDD Phase 1 - 核心领域模型实现" \
  --body "
## 概述

实现 DDD Phase 1 的所有核心聚合根和值对象。

## 变更

### 聚合根
- Agent 聚合根：管理 AI Agent 身份、能力、个性
- ConversationSession 聚合根：管理会话生命周期
- User 聚合根：多租户用户管理
- LLMProvider 聚合根：LLM 提供者配置

### 值对象
- Identity：Agent 身份
- Capability：Agent 能力
- PersonalityConfig：个性配置
- ThreadData：线程虚拟路径

### 领域事件
- AgentCreatedEvent、CapabilityAddedEvent 等
- SessionCreatedEvent、MessageAppendedEvent 等

### 基础设施
- SimpleEventBus：基础事件总线
- DI 容器框架

## 测试
- 单元测试覆盖率 82%
- 所有聚合根验证通过

## 下一步
- Phase 2：实施仓储与中间件

## 相关链接
- 设计文档：docs/design/030-DDD重构方案-完整计划.md
"
```

---

## 常见问题

### Q1：值对象是否应该实现 MarshalJSON?

**A**：是的。实现 `MarshalJSON` 和 `UnmarshalJSON` 以支持序列化。

```go
func (i *Identity) MarshalJSON() ([]byte, error) {
    type Alias Identity
    return json.Marshal(&struct {
        *Alias
    }{Alias: (*Alias)(i)})
}
```

### Q2：事务安全如何处理？

**A**：暂时不实现完全的事务性。留待 Phase 2 使用 `BeginTransaction` 模式。

### Q3：并发访问是否安全？

**A**：聚合根本身不需要并发保护（不可变；版本字段用于乐观锁）。并发安全在仓储层处理。

### Q4：如何处理对象克隆？

**A**：暂不需要。值对象通过创建新实例实现"克隆"。

---

## 后续支持

如有问题，查看以下资源：

- [DDD 完整方案](./030-DDD重构方案-完整计划.md)
- [DDD 结构分析](./DDD_STRUCTURE_ANALYSIS.md)
- [项目 AGENTS.md](../../AGENTS.md)
- Slack: `#backend` 频道

