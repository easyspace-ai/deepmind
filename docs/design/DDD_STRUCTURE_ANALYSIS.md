# DDD (Domain-Driven Design) Structure Analysis - nanobot-go

**Analysis Date:** 2026-03-25

---

## Executive Summary

The nanobot-go project demonstrates a **multi-domain architecture** with clear separation of concerns. It uses Go best practices with GORM for persistence, interfaces for dependency injection, and asynchronous messaging via a centralized MessageBus. However, there are several areas where stronger DDD patterns (aggregates, value objects, domain events) could be applied.

---

## 1. Core Bounded Contexts (Domains)

The system is organized into the following **bounded contexts**:

### 1.1 Agent Orchestration Domain
**Purpose:** Manage AI agent configuration, personality, capabilities, and execution  
**Location:** `internal/service/agent/`, `internal/models/agent.go`, `pkg/agent/`

**Key Responsibilities:**
- Agent lifecycle (create, update, delete)
- Configuration management (model, temperature, tokens, etc.)
- Capability management (skills, tools, MCP bindings)
- Default agent selection
- Multi-tenant agent isolation

**Related Files:**
- [internal/models/agent.go](internal/models/agent.go) - Agent entity/aggregate
- [internal/repository/agent.go](internal/repository/agent.go) - Agent persistence interface
- [internal/service/agent/types.go](internal/service/agent/types.go) - Agent service interface
- [pkg/agent/master_agent.go](pkg/agent/master_agent.go) - Master agent orchestration

**Aggregate Root:** `Agent`

---

### 1.2 Channel/Communication Domain  
**Purpose:** Handle multi-channel communication (Feishu, WebSocket, DingTalk, Matrix)  
**Location:** `internal/service/channel.go`, `internal/models/channel.go`, `pkg/channels/`

**Key Responsibilities:**
- Channel configuration and management
- User/Agent channel bindings
- Channel-specific payloads (Feishu, WebSocket, DingTalk configs)
- Channel activation/deactivation
- Multi-tenant channel isolation

**Related Files:**
- [internal/models/channel.go](internal/models/channel.go) - Channel entity with type hierarchy
- [internal/repository/channel.go](internal/repository/channel.go) - Channel persistence
- [internal/service/channel.go](internal/service/channel.go) - Channel service
- [pkg/channels/base.go](pkg/channels/base.go) - Channel interface & manager

**Aggregate Root:** `Channel`

**Value Objects:** `ChannelType`, `FeishuChannelConfig`, `DingTalkChannelConfig`, `MatrixChannelConfig`, `WebSocketChannelConfig`

---

### 1.3 Session Management Domain
**Purpose:** Track active conversation sessions and lifecycle  
**Location:** `internal/service/session.go`, `internal/models/session.go`, `pkg/session/`

**Key Responsibilities:**
- Session creation and lifecycle
- User-Agent-Channel session association
- Session context management (cancellation, metadata)
- Last activity tracking
- Memory-based runtime vs. persistent storage separation

**Related Files:**
- [internal/models/session.go](internal/models/session.go) - Session database model
- [internal/repository/session.go](internal/repository/session.go) - Session persistence interface
- [internal/service/session.go](internal/service/session.go) - Session service
- [pkg/session/manager.go](pkg/session/manager.go) - Runtime session manager

**Aggregate Root:** `Session`

**Key Pattern:** Separation of concerns - persistent Session model vs. runtime Session with context management

---

### 1.4 Conversation/Message Domain
**Purpose:** Store and query conversation history with audit trail (trace/span)  
**Location:** `internal/service/conversation/`, `internal/models/conversation_record.go`

**Key Responsibilities:**
- Conversation record persistence
- Message history tracking (role, content, tokens)
- Distributed tracing (TraceID, SpanID)
- Token usage analytics
- Multi-tenant conversation isolation
- Query and pagination

**Related Files:**
- [internal/models/conversation_record.go](internal/models/conversation_record.go) - Conversation event model
- [internal/repository/conversation_record.go](internal/repository/conversation_record.go) - Persistence interface
- [internal/service/conversation/types.go](internal/service/conversation/types.go) - Conversation service interface

**Aggregate Root:** `ConversationRecord` (immutable append-only)

**Value Objects:** `ConversationDTO`, `TokenUsageDTO`, `QueryOptions`

**Pattern:** Event sourcing-like - records are immutable append-only events with distributed trace context

---

### 1.5 User/Authentication Domain
**Purpose:** User management and multi-tenancy support  
**Location:** `internal/service/user.go`, `internal/models/user.go`

**Key Responsibilities:**
- User registration and authentication
- User code generation
- Multi-tenant user isolation via UserCode
- Password management (hashed)
- User activation/deactivation

**Related Files:**
- [internal/models/user.go](internal/models/user.go) - User entity
- [internal/repository/user.go](internal/repository/user.go) - User persistence
- [internal/service/user.go](internal/service/user.go) - User service

**Aggregate Root:** `User`

---

### 1.6 LLM Provider/Model Configuration Domain
**Purpose:** Manage LLM API credentials and supported models  
**Location:** `internal/service/provider.go`, `internal/models/llm_provider.go`

**Key Responsibilities:**
- API key management (encrypted storage)
- Supported models list management
- Embedding model configuration
- Provider prioritization
- Provider status tracking

**Related Files:**
- [internal/models/llm_provider.go](internal/models/llm_provider.go) - LLM provider config
- [internal/repository/provider.go](internal/repository/provider.go) - Provider persistence
- [internal/service/provider.go](internal/service/provider.go) - Provider service

**Aggregate Root:** `LLMProvider`

**Value Objects:** `ModelInfo`, `EmbeddingModelInfo`, `EmbeddingModelConfig`

---

### 1.7 MCP (Model Context Protocol) Server Domain
**Purpose:** External tool/capability integration via MCP servers  
**Location:** `internal/service/mcp/`, `internal/models/mcp_*.go`

**Key Responsibilities:**
- MCP server configuration (stdio, http, sse transports)
- Tool discovery and storage
- Agent-MCP bindings
- Environment variable management
- MCP server status tracking
- Tool capability filtering

**Related Files:**
- [internal/models/mcp_server.go](internal/models/mcp_server.go) - MCP server config
- [internal/models/mcp_tool.go](internal/models/mcp_tool.go) - Tool tracking
- [internal/models/agent_mcp_binding.go](internal/models/agent_mcp_binding.go) - Agent-MCP association
- [internal/repository/mcp_server.go](internal/repository/mcp_server.go) - MCP persistence
- [internal/service/mcp/types.go](internal/service/mcp/types.go) - MCP service

**Aggregate Root:** `MCPServer` (with `MCPToolModel` as aggregate entity)

**Sub-Aggregate:** `AgentMCPBinding`

**Value Objects:** `MCPTool`, config structs

---

## 2. Entity and Aggregate Types

### Primary Aggregates (Aggregate Roots)

| Aggregate | File | Description | Tenant Scope |
|-----------|------|-------------|---------|
| `Agent` | [internal/models/agent.go](internal/models/agent.go) | AI agent with capabilities & config | UserCode |
| `Channel` | [internal/models/channel.go](internal/models/channel.go) | Communication channel with provider config | UserCode |
| `Session` | [internal/models/session.go](internal/models/session.go) | Active conversation session | UserCode + AgentCode + ChannelCode |
| `ConversationRecord` | [internal/models/conversation_record.go](internal/models/conversation_record.go) | Immutable message event | TraceID (distributed trace) |
| `User` | [internal/models/user.go](internal/models/user.go) | System user with auth | Self (id) |
| `LLMProvider` | [internal/models/llm_provider.go](internal/models/llm_provider.go) | LLM API configuration | UserCode |
| `MCPServer` | [internal/models/mcp_server.go](internal/models/mcp_server.go) | External tool provider | Global (shared) |

### Child Entities/Value Objects with Aggregates

| Entity | Parent | File | Purpose |
|--------|--------|------|---------|
| `AgentMCPBinding` | Agent | [internal/models/agent_mcp_binding.go](internal/models/agent_mcp_binding.go) | Agent-specific MCP config |
| `MCPToolModel` | MCPServer | [internal/models/mcp_tool.go](internal/models/mcp_tool.go) | Tool from MCP server |
| `CronJob` | Agent | [internal/models/cron_job.go](internal/models/cron_job.go) | Scheduled agent execution |

---

## 3. Domain Value Objects

Value objects are **immutable, equality-based objects with no identity**. The system uses several:

### 3.1 Channel Configuration Value Objects

```go
// FeishuChannelConfig - Feishu-specific configuration
type FeishuChannelConfig struct {
    AppID             string
    AppSecret         string
    EncryptKey        string
    VerificationToken string
}

// DingTalkChannelConfig - DingTalk-specific configuration
type DingTalkChannelConfig struct {
    ClientID     string
    ClientSecret string
}

// MatrixChannelConfig - Matrix-specific configuration
type MatrixChannelConfig struct {
    Homeserver string
    UserID     string
    Token      string
}

// WebSocketChannelConfig - WebSocket configuration
type WebSocketChannelConfig struct {
    Addr string
    Path string
}
```
**File:** [internal/models/channel.go](internal/models/channel.go)

**Pattern:** Stored as JSON in `Channel.Config` field; deserialized on use

---

### 3.2 LLM Provider Model Value Objects

```go
// ModelInfo - Represents an available model
type ModelInfo struct {
    ID        string
    Name      string
    MaxTokens int
}

// EmbeddingModelInfo - Embedding model specification
type EmbeddingModelInfo struct {
    ID         string
    Dimensions int
}

// EmbeddingModelConfig - Runtime embedding config
type EmbeddingModelConfig struct {
    APIKey     string
    BaseURL    string
    Model      string
    Dimensions int
}
```
**File:** [internal/models/llm_provider.go](internal/models/llm_provider.go)

---

### 3.3 Conversation Analysis Value Objects

```go
// ConversationDTO - Data transfer object (not strictly a value object, but used as immutable)
type ConversationDTO struct {
    ID           uint
    TraceID      string
    SpanID       string
    EventType    string
    Timestamp    time.Time
    Role         string
    Content      string
    TokenUsage   *TokenUsageDTO
}

// TokenUsageDTO - Token consumption metrics
type TokenUsageDTO struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
    ReasoningTokens  int
    CachedTokens     int
}

// QueryOptions - Query parameters (typically not persisted)
type QueryOptions struct {
    OrderBy string
    Order   string
    Limit   int
    Offset  int
    Roles   []string
}
```
**File:** [internal/service/conversation/types.go](internal/service/conversation/types.go)

---

### 3.4 Channel Types

```go
type ChannelType string

const (
    ChannelTypeFeishu    ChannelType = "feishu"
    ChannelTypeWebSocket ChannelType = "websocket"
)
```
**File:** [internal/models/channel.go](internal/models/channel.go)

---

### 3.5 Workflow/Task Value Objects

```go
type WorkflowType string
const (
    WorkflowTypeSequential WorkflowType = "sequential"
    WorkflowTypeParallel   WorkflowType = "parallel"
    WorkflowTypeTask       WorkflowType = "task"
)

type SubTaskType string
const (
    SubTaskTypeAnalyze SubTaskType = "analyze"
    SubTaskTypeCreate  SubTaskType = "create"
    SubTaskTypeModify  SubTaskType = "modify"
    SubTaskTypeDelete  SubTaskType = "delete"
    SubTaskTypeTest    SubTaskType = "test"
    SubTaskTypeVerify  SubTaskType = "verify"
)

type Complexity string
const (
    ComplexitySimple  Complexity = "simple"
    ComplexityMedium  Complexity = "medium"
    ComplexityComplex Complexity = "complex"
)
```
**File:** [pkg/planner/model/workflow.go](pkg/planner/model/workflow.go), [pkg/planner/model/task.go](pkg/planner/model/task.go), [pkg/planner/model/intent.go](pkg/planner/model/intent.go)

---

## 4. Service Layer Patterns

The project uses **well-defined service interfaces** with dependency injection. Services follow a **repository pattern** for persistence abstraction.

### 4.1 Service Pattern Structure

**General pattern:**
```go
// Public interface
type Service interface {
    CreateXxx(...) (*Model, error)
    GetXxx(...) (*Model, error)
    UpdateXxx(...) error
    DeleteXxx(...) error
    ListXxx(...) ([]Model, error)
}

// Private implementation
type service struct {
    xxxRepository Repository
    // other dependencies
}

// Factory
func NewService(repo Repository, deps ...interface{}) Service {
    return &service{xxxRepository: repo}
}
```

### 4.2 Service Implementations

| Service | File | Responsibilities | Dependencies |
|---------|------|------------------|--------------|
| `AgentService` | [internal/service/agent/types.go](internal/service/agent/types.go) | Agent CRUD + config + capabilities | AgentRepository, CodeService |
| `ChannelService` | [internal/service/channel.go](internal/service/channel.go) | Channel CRUD + config management | ChannelRepository, CodeService |
| `SessionService` | [internal/service/session.go](internal/service/session.go) | Session CRUD + activity tracking | SessionRepository |
| `ConversationService` | [internal/service/conversation/types.go](internal/service/conversation/types.go) | Message history + analytics | ConversationRepository |
| `UserService` | [internal/service/user.go](internal/service/user.go) | User CRUD + auth | UserRepository |
| `ProviderService` | [internal/service/provider.go](internal/service/provider.go) | LLM provider config + model mgmt | ProviderRepository |
| `MCPService` | [internal/service/mcp/types.go](internal/service/mcp/types.go) | MCP server config + tool mgmt | MCPServerRepository |
| `SkillService` | [internal/service/skill/service.go](internal/service/skill/service.go) | Skill catalog management | SkillRepository |
| `CronJobService` | [internal/service/cron.go](internal/service/cron.go) | Scheduled job management | CronJobRepository |
| `EmbeddingService` | [internal/service/embedding/config.go](internal/service/embedding/config.go) | Embedding model config | EmbeddingRepository |

### 4.3 Cross-Service Coordination

**Code Lookup Service Pattern:**
```go
type CodeLookupService interface {
    GetUserByCode(code string) (*models.User, error)
    GetChannelByCode(code string) (*models.Channel, error)
    GetAgentByCode(code string) (*models.Agent, error)
}
```
**Purpose:** Provides fast lookup by business "code" fields (not database IDs)

**File:** [internal/api/handler.go](internal/api/handler.go)

---

## 5. Repository/Persistence Interfaces

The system follows the **Repository Pattern** for data access abstraction.

### 5.1 Repository Interfaces

Located in: [internal/repository/](internal/repository/)

| Repository | File | Methods | Aggregate |
|------------|------|---------|-----------|
| `AgentRepository` | [agent.go](internal/repository/agent.go) | Create, GetByID, GetByAgentCode, GetByUserCode, Update, Delete, GetWithChannels | Agent |
| `ChannelRepository` | [channel.go](internal/repository/channel.go) | Create, GetByID, GetByCode, GetByUserCode, Update, Delete | Channel |
| `SessionRepository` | [session.go](internal/repository/session.go) | Create, GetByID, GetBySessionKey, GetByChannelCode, UpdateLastActive | Session |
| `ConversationRecordRepository` | [conversation_record.go](internal/repository/conversation_record.go) | Create, CreateBatch, GetByTraceID, ListBySessionKey | ConversationRecord |
| `UserRepository` | [user.go](internal/repository/user.go) | Create, GetByID, GetByUsername, Update, Delete | User |
| `ProviderRepository` | [provider.go](internal/repository/provider.go) | Create, GetByID, GetByUserCode, Update, Delete | LLMProvider |
| `MCPServerRepository` | [mcp_server.go](internal/repository/mcp_server.go) | Create, GetByID, GetByCode, List, Update, Delete | MCPServer |
| `MCPToolRepository` | [mcp_tool.go](internal/repository/mcp_tool.go) | Create, GetByMCPServerID, Update, Delete | MCPToolModel |
| `AgentMCPBindingRepository` | [agent_mcp_binding.go](internal/repository/agent_mcp_binding.go) | Create, GetByAgentID, Update, Delete | AgentMCPBinding |

### 5.2 Example Repository Interface

```go
// AgentRepository - Agent persistence interface
type AgentRepository interface {
    Create(agent *models.Agent) error
    GetByID(id uint) (*models.Agent, error)
    GetByAgentCode(code string) (*models.Agent, error)
    GetByAgentCodes(codes []string) ([]*models.Agent, error)  // Batch query
    GetByUserCode(userCode string) ([]models.Agent, error)    // Tenant isolation
    GetDefaultByUserCode(userCode string) (*models.Agent, error)
    Update(agent *models.Agent) error
    Delete(id uint) error
    GetWithChannels(id uint) (*models.Agent, error)           // Eager loading
    CheckAgentCodeExists(code string) (bool, error)
    ListAll() ([]models.Agent, error)
}
```

### 5.3 Implementation Pattern

All repositories use **GORM** for ORM with consistent error handling:

```go
type agentRepository struct {
    db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) AgentRepository {
    return &agentRepository{db: db}
}

func (r *agentRepository) Create(agent *models.Agent) error {
    if err := r.db.Create(agent).Error; err != nil {
        return fmt.Errorf("failed to create agent: %w", err)
    }
    return nil
}
```

---

## 6. Cross-Cutting Concerns

### 6.1 Message Bus / Event Publishing

**Pattern:** Asynchronous pub-sub messaging for decoupling

**File:** [pkg/bus/queue.go](pkg/bus/queue.go)

**Components:**
```go
type MessageBus struct {
    inbound             chan *InboundMessage
    outbound            chan *OutboundMessage
    stream              chan *StreamChunk
    outboundSubscribers map[string][]OutboundCallback
    streamSubscribers   map[string][]StreamCallback
    taskEventSubscribers []TaskEventCallback
}

// Publishing patterns
func (b *MessageBus) PublishInbound(msg *InboundMessage)
func (b *MessageBus) PublishOutbound(msg *OutboundMessage)
func (b *MessageBus) PublishStream(chunk *StreamChunk)
func (b *MessageBus) PublishTaskEvent(eventType string, payload map[string]any)
```

**Usage:**
- Channels → MessageBus.PublishInbound → Agent processing
- Agent processing → MessageBus.PublishOutbound → Channels (responses)
- Real-time updates → MessageBus.PublishStream → WebSocket/SSE clients
- Task lifecycle events → MessageBus.PublishTaskEvent → Task tracking UI

---

### 6.2 Authentication & Authorization Middleware

**Pattern:** JWT-based authentication with context propagation

**File:** [internal/api/middleware.go](internal/api/middleware.go)

**Middleware:**
```go
// AuthMiddleware - JWT token validation
func AuthMiddleware() gin.HandlerFunc {
    // Extract & validate Bearer token
    // Store userID, username in context
}

// CorsMiddleware - Cross-origin resource sharing
func CorsMiddleware() gin.HandlerFunc {
    // Allow cross-origin requests
    // Set appropriate CORS headers
}
```

---

### 6.3 Multi-Tenancy Pattern

**Key Implementation:**
- Every major entity includes `UserCode` field for tenant isolation
- Database queries filter by `UserCode`
- JWT token contains `userID` → resolved to `UserCode` at handler level

**Entities with tenant isolation:**
- Agent (`user_code` index)
- Channel (`user_code` index)
- Session (`user_code` index)
- LLMProvider (`user_code` index)
- ConversationRecord (`user_code` index)

---

### 6.4 Session/Context Management

**Pattern:** Context-based cancellation for long-running tasks

**File:** [pkg/session/manager.go](pkg/session/manager.go)

```go
type Session struct {
    Key       string
    CreatedAt time.Time
    UpdatedAt time.Time
    cancel    context.CancelFunc  // Runtime cancellation
    ctx       context.Context     // Execution context
}

// Separation: persistent data in database vs. runtime state
```

**Two-tier session model:**
1. **Database Session** (`internal/models/session.go`) - persistent metadata
2. **Runtime Session** (`pkg/session/manager.go`) - in-memory execution context

---

### 6.5 Distributed Tracing

**Pattern:** TraceID + SpanID for correlating distributed actions

**File:** [internal/models/conversation_record.go](internal/models/conversation_record.go)

```go
type ConversationRecord struct {
    TraceID      string    // Request correlation ID
    SpanID       string    // Operation within trace
    ParentSpanID string    // Parent operation
    // ... other fields
}
```

---

### 6.6 Logging

**Pattern:** Structured logging with zap logger

**Usage Pattern:**
```go
logger.Info("message", zap.String("key", "value"))
logger.Error("error occurred", zap.Error(err))
```

---

## 7. Anti-Patterns & Concerns

### ⚠️ 7.1 Anemic Model Anti-Pattern (Moderate Risk)

**Issue:** Entities lack business logic; mostly data containers

**Examples:**
```go
// Agent is mostly properties, minimal behavior
type Agent struct {
    ID              uint
    AgentCode       string
    Name            string
    SkillsList      string    // Stored as JSON string
    ToolsList       string    // Stored as JSON string
    // ... 20+ more primitive fields
}
```

**Getter methods exist but limited domain logic:**
```go
func (a *Agent) GetAvailableSkills() []string {
    // Just unmarshals JSON; no validation or domain rules
}
```

**Impact:** Makes it hard to enforce domain rules (invariants). Business logic ends up in services.

**Recommendation:** 
- Add factory methods for safe construction
- Encapsulate related fields (e.g., skill management)
- Move validation to aggregate

---

### ⚠️ 7.2 JSON Serialization for Complex Values (Low-Medium Risk)

**Issue:** Nested configurations stored as JSON strings in database

**Pattern:**
```go
type Channel struct {
    Config string `json:"config"` // Entire config as JSON string
}

type LLMProvider struct {
    SupportedModels string `json:"supported_models"`     // JSON array of ModelInfo
    EmbeddingModels string `json:"embedding_models"`    // JSON array of EmbeddingModelInfo
}
```

**Implications:**
- No database-level validation of structure
- Deserialization errors not caught until runtime
- Difficult to query on nested properties
- Prone to silent data corruption

**Recommendation:**
- Create separate tables for complex nested data (ModelInfo, ConfigEntry)
- Use foreign keys and relational structure
- Keep only simple scalar values as JSON

---

### ⚠️ 7.3 Missing Domain Events (Medium Risk)

**Issue:** No explicit domain events for important state changes

**Current State:**
- When an Agent is updated, no `AgentUpdated` event is raised
- When Session is created, services don't publish a `SessionCreated` event
- Changes aren't captured for audit trail, webhooks, or cross-domain synchronization

**Pattern Gap:** Event-driven architecture is partially implemented (MessageBus exists) but not used for domain layer

**Recommendation:**
- Define domain events (AgentConfigured, ChannelActivated, SessionStarted, etc.)
- Publish events from aggregate roots
- Subscribe services to react to events

---

### ⚠️ 7.4 Thin Repository Pattern (Low Risk)

**Issue:** Repositories only wrap GORM calls; lack complex queries

**Current State:**
```go
func (r *agentRepository) GetByAgentCode(code string) (*models.Agent, error) {
    var agent models.Agent
    if err := r.db.Where("agent_code = ?", code).First(&agent).Error; err != nil {
        return nil, err
    }
    return &agent, nil
}
```

**Gap:** No repositories for complex queries across aggregates (e.g., "find all sessions for an agent in the last hour")

**Recommendation:**
- Add **Query Objects** for complex filtering
- Create **Specification** pattern for reusable query conditions
- Consider **Query Service** separate from repository

---

### ⚠️ 7.5 No Explicit Boundary Specification (Medium Risk)

**Issue:** Limits of bounded contexts not explicitly documented

**Current State:**
- Message passing is clear (pub-sub via MessageBus)
- Domain models don't explicitly mark which fields are "internal" vs. "exposed"
- DTOs used for external APIs, but no schema or documentation

**Recommendation:**
- Create context map document (which domains interact, when, how)
- Use Tags/comments to mark API boundaries
- Define UnboundedContext explicitly (MCP servers are shared, not multi-tenant)

---

### ⚠️ 7.6 Blob Anti-Pattern in Message Fields (Low Risk)

**Issue:** Conversation record `Content` field is unstructured text/blob

```go
type ConversationRecord struct {
    Content string `json:"content"` // Could be text, JSON, markdown, HTML
}
```

**Impact:** 
- Can't query by subcontent
- Different message types (user input, tool result, error) mixed without structure

**Recommendation:**
- Create Message Value Object with typed payloads
- Use discriminated union / type field
- Enables richer message analytics

---

### ⚠️ 7.7 Aggregate Boundary Incertitude (Low-Medium Risk)

**Issue:** Unclear which entities belong to which aggregate

**Examples:**
- Is `AgentMCPBinding` part of `Agent` aggregate or a separate aggregate?
- Is `MCPToolModel` part of `MCPServer` aggregate or standalone?
- What's the consistency boundary for "enable/disable tool for agent"?

**Current State:** Relationships exist but aggregate boundaries not explicitly defined

**Recommendation:**
- Document aggregate boundaries (Agent owns MCPBinding vs. AgentMCPBinding is its own AR)
- Define consistency guarantees per aggregate
- Specify transaction boundaries

---

## 8. Summary Table: DDD Structure

| Aspect | Status | Assessment | Files |
|--------|--------|------------|-------|
| **Bounded Contexts** | ✅ Well-defined | 7 clear domains | service/, models/ |
| **Aggregates** | ⚠️ Implicit | Entities exist but boundaries fuzzy | models/*.go |
| **Value Objects** | ✅ Good | Rich value object hierarchy | models/, service/types |
| **Repositories** | ✅ Good | Clean interface-based pattern | repository/*.go |
| **Services** | ✅ Good | Clear service interfaces | service/ |
| **Domain Events** | ❌ Missing | Pub-sub exists but not domain events | bus/queue.go |
| **Middleware/CC** | ✅ Good | Auth, CORS, multi-tenancy | api/middleware.go |
| **Message Bus** | ✅ Good | Inbound/outbound/stream channels | pkg/bus/ |
| **Session Mgmt** | ✅ Good | Context-based cancellation | pkg/session/ |
| **Multi-Tenancy** | ✅ Good | UserCode isolation pattern | all models |
| **Distributed Tracing** | ✅ Good | TraceID/SpanID correlation | conversation_record |

---

## 9. Recommendations

### Priority 1: High Impact
1. **Add Domain Events** - Publish `AgentConfigured`, `ChannelActivated`, `SessionStarted` events for audit and sync
2. **Document Aggregate Boundaries** - Explicitly specify which entities belong to which aggregate
3. **Extract Blob Content** - Structure `ConversationRecord.Content` with typed message payloads

### Priority 2: Medium Impact
4. **Rich Aggregates** - Add domain logic and invariant checking to aggregate roots
5. **Query Objects** - Implement specification pattern for complex queries
6. **Configuration Tables** - Replace JSON strings with relational tables (ModelInfo, ConfigEntry)

### Priority 3: Nice to Have
7. **Factory Methods** - Add bundle construction logic to aggregates
8. **Anti-Corruption Layer** - Explicit adapters for external system integration
9. **Ubiquitous Language Documentation** - Glossary of domain terms

---

## File Structure Map

```
nanobot-go/
├── internal/
│   ├── api/                           # HTTP handlers & middleware
│   │   ├── handler.go                # Main handler with service deps
│   │   ├── middleware.go             # Auth, CORS middleware
│   │   ├── *_handler.go              # Domain-specific handlers
│   │   └── providers.go              # Service provider setup
│   ├── models/                        # Domain entities (aggregates)
│   │   ├── agent.go                  # Agent aggregate
│   │   ├── channel.go                # Channel aggregate + value objects
│   │   ├── session.go                # Session aggregate
│   │   ├── conversation_record.go    # Conversation aggregate
│   │   ├── user.go                   # User aggregate
│   │   ├── llm_provider.go           # LLM provider aggregate
│   │   ├── mcp_server.go             # MCP server aggregate
│   │   ├── mcp_tool.go               # MCP tool entity
│   │   └── agent_mcp_binding.go      # Agent-MCP binding entity
│   ├── repository/                    # Repository pattern interfaces
│   │   ├── agent.go                  # AgentRepository
│   │   ├── channel.go                # ChannelRepository
│   │   ├── session.go                # SessionRepository
│   │   ├── conversation_record.go    # ConversationRecordRepository
│   │   ├── user.go                   # UserRepository
│   │   ├── provider.go               # ProviderRepository
│   │   └── *.go                      # Other repositories
│   ├── service/                       # Service layer (use cases)
│   │   ├── agent/                    # Agent service context
│   │   │   └── types.go
│   │   ├── channel.go                # Channel service
│   │   ├── session.go                # Session service
│   │   ├── user.go                   # User service
│   │   ├── provider.go               # LLM provider service
│   │   ├── mcp/                      # MCP service context
│   │   ├── conversation/             # Conversation service context
│   │   └── *.go                      # Other services
│   ├── app/
│   │   └── config.go                 # Default configuration
│   └── database/                     # Database initialization
├── pkg/
│   ├── agent/                        # Agent orchestration (domain code)
│   │   ├── master_agent.go
│   │   ├── memory/
│   │   ├── task/
│   │   ├── state/
│   │   └── ...
│   ├── bus/                          # Message bus (cross-cutting)
│   │   ├── queue.go                  # MessageBus implementation
│   │   └── events.go                 # Event types
│   ├── channels/                     # Channel adapters
│   │   ├── base.go                   # Channel interface & manager
│   │   ├── feishu/
│   │   └── websocket/
│   ├── session/                      # Session management
│   │   └── manager.go
│   ├── planner/                      # Planner/orchestration domain
│   │   └── model/
│   └── ...
└── docs/
    └── design/
        └── DDD_STRUCTURE_ANALYSIS.md (this file)
```

---

**Document Version:** 1.0  
**Last Updated:** 2026-03-25  
**Author:** DDD Analysis Agent
