# DDD Architecture Quick Reference - nanobot-go

## Core Domains at a Glance

| Domain | Root Aggregate | Repository | Service | TenantScoped | Key Entities |
|--------|----------------|------------|---------|-------------|-------------|
| **Agent** | `Agent` | `AgentRepository` | `AgentService` | UserCode | AgentMCPBinding |
| **Channel** | `Channel` | `ChannelRepository` | `ChannelService` | UserCode | [FeishuConfig, WebSocketConfig, ...] |
| **Session** | `Session` | `SessionRepository` | `SessionService` | UserCode | RuntimeSession (pkg/session) |
| **Conversation** | `ConversationRecord` | `ConversationRepository` | `ConversationService` | TraceID | TokenUsageDTO |
| **User** | `User` | `UserRepository` | `UserService` | Self | - |
| **LLMProvider** | `LLMProvider` | `ProviderRepository` | `ProviderService` | UserCode | ModelInfo, EmbeddingModelInfo |
| **MCPServer** | `MCPServer` | `MCPServerRepository` | `MCPService` | Global | MCPToolModel, AgentMCPBinding |

---

## File Location Cheat Sheet

### Models (Entities & Value Objects)
- `internal/models/agent.go` - Agent aggregate
- `internal/models/channel.go` - Channel aggregate + configs
- `internal/models/session.go` - Session aggregate
- `internal/models/conversation_record.go` - Immutable message events
- `internal/models/user.go` - User aggregate
- `internal/models/llm_provider.go` - LLM provider config
- `internal/models/mcp_server.go` - MCP server config

### Repositories
- `internal/repository/agent.go` - Agent persistence interface
- `internal/repository/channel.go` - Channel persistence
- `internal/repository/session.go` - Session persistence
- `internal/repository/conversation_record.go` - Conversation persistence
- `internal/repository/user.go` - User persistence
- `internal/repository/provider.go` - LLM provider persistence
- `internal/repository/mcp_server.go` - MCP server persistence

### Services
- `internal/service/agent/types.go` - Agent service (CRUD + config)
- `internal/service/channel.go` - Channel service
- `internal/service/session.go` - Session service
- `internal/service/conversation/types.go` - Conversation service
- `internal/service/user.go` - User service
- `internal/service/provider.go` - LLM provider service
- `internal/service/mcp/types.go` - MCP service (server + tool mgmt)

### API & HTTP
- `internal/api/handler.go` - Main handler dispatcher
- `internal/api/middleware.go` - Auth & CORS middleware
- `internal/api/*_handler.go` - Domain-specific handlers
- `internal/api/providers.go` - Service dependency injection

### Cross-Cutting (Package-level)
- `pkg/bus/queue.go` - Message bus (inbound, outbound, stream, events)
- `pkg/session/manager.go` - Runtime session management
- `pkg/channels/base.go` - Channel adapter interface & manager
- `pkg/agent/master_agent.go` - Agent orchestration

---

## Key Patterns Used

### 1. Repository Pattern
```go
// Interface-based abstraction for data access
type AgentRepository interface {
    Create(...) error
    GetByID(...) (*Agent, error)
    // ... more methods
}

// Implementation with GORM
type agentRepository struct {
    db *gorm.DB
}
```
**Why:** Decouples business logic from persistence details

---

### 2. Service Interface + Implementation
```go
type Service interface {
    CreateAgent(...) (*Agent, error)
    GetAgent(...) (*Agent, error)
    // ... more methods
}

type service struct {
    agentRepository AgentRepository
    codeService CodeService
}
```
**Why:** Clear contracts, dependency injection, easy to mock

---

### 3. Value Objects as Configuration
```go
// Stored as JSON in database, typed in code
type FeishuChannelConfig struct {
    AppID      string
    AppSecret  string
    EncryptKey string
}

// Unmarshalled on use
config := UnmarshalFeishuConfig(channel.Config)
```
**Why:** Type safety while maintaining schema flexibility

---

### 4. Message Bus for Async Communication
```go
// Decouples channels from agent core
bus.PublishInbound(&InboundMessage{...})    // Device → Agent
bus.PublishOutbound(&OutboundMessage{...})  // Agent → Device
bus.PublishStream(&StreamChunk{...})        // Agent → UI
```
**Why:** Loose coupling, allows multiple subscribers

---

### 5. Multi-Tenancy via UserCode
```go
// Every major aggregate includes UserCode
type Agent struct {
    ID       uint
    UserCode string  // Tenant key
    // ... fields
}

// Queries automatically filtered
agents := repo.GetByUserCode(userCode)
```
**Why:** Simple, efficient tenant isolation

---

### 6. Distributed Tracing Context
```go
// Every conversation tied to TraceID
type ConversationRecord struct {
    TraceID      string    // Request correlation
    SpanID       string    // Operation within trace
    ParentSpanID string    // Nesting
}
```
**Why:** Audit trail, debugging, performance analysis

---

### 7. Middleware-based Authentication
```go
// JWT auth middleware adds userID to context
authorized := router.Group("/api/v1")
authorized.Use(AuthMiddleware())

// Resolved to UserCode in handlers
userCode := getLookupService().GetUserCode(userID)
```
**Why:** Centralized auth, applied to all protected routes

---

## Common Query Patterns

### Get Agent by Business Code (not ID)
```go
agent, err := agentService.GetAgentByCode(agentCode)
// Looks up by AgentCode, not numeric ID
```

### Get User's Agents
```go
agents, err := agentService.GetUserAgents(userCode)
// Tenant-scoped query
```

### Get Conversation History
```go
result, err := conversationService.ListBySessionKey(ctx, sessionKey, page, pageSize)
// Paginated history with token counts
```

### Get MCP Tools for Agent
```go
binding, err := mcpService.GetAgentBinding(agentID, mcpServerID)
enabled := binding.IsToolEnabled(toolName)
// Checks if tool enabled in agent-specific binding
```

---

## Anti-Pattern Warnings

### ⚠️ Don't Add Business Logic Directly to Handlers
```go
// ❌ WRONG: Business logic in handler
func (h *Handler) handleAgent(c *gin.Context) {
    // Database query
    // Validation
    // Transformation
    c.JSON(200, result)
}

// ✅ RIGHT: Delegate to service
func (h *Handler) handleAgent(c *gin.Context) {
    result, err := h.agentService.GetAgent(id)
    c.JSON(200, result)
}
```

### ⚠️ Don't Bypass Repositories Directly
```go
// ❌ WRONG: Direct GORM access in service
func (s *service) GetAgent(agentCode string) error {
    s.db.Where("agent_code = ?", agentCode).First(&agent)
}

// ✅ RIGHT: Use repository interface
func (s *service) GetAgent(agentCode string) error {
    agent, err := s.agentRepo.GetByAgentCode(agentCode)
}
```

### ⚠️ Don't Mix Tenant Contexts
```go
// ❌ WRONG: Forget to check tenant
agents := agentRepo.GetAll()  // Anyone can see all agents

// ✅ RIGHT: Always scope by tenant
agents := agentRepo.GetByUserCode(userCode)  // Tenant-isolated
```

### ⚠️ Don't Store Complex Logic as JSON
```go
// ❌ WRONG: Unmarshalling business logic in multiple places
if skills, err := unmarshalSkills(agent.SkillsList); err != nil {
    // Handle error in multiple services
}

// ✅ RIGHT: Aggregate method handles
skills := agent.GetAvailableSkills()  // Single source of truth
```

---

## Testing Recommendations

### Unit Test Repository
```go
// Mock database behavior
type mockAgentRepo struct {
    agents map[string]*models.Agent
}

func (m *mockAgentRepo) GetByCode(code string) (*models.Agent, error) {
    // Return test data
}

// Test service with mock repo
service := NewService(mockAgentRepo)
result, err := service.GetAgent(code)
```

### Unit Test Service
```go
// Test business logic with mocked dependencies
func TestAgentService_CreateAgent(t *testing.T) {
    mockRepo := &mockAgentRepository{}
    mockCodeService := &mockCodeService{}
    
    service := NewService(mockRepo, mockCodeService)
    agent, err := service.CreateAgent(req)
    
    assert.NoError(t, err)
    assert.Equal(t, "expected_code", agent.AgentCode)
}
```

### Integration Test Handler
```go
// Test full HTTP flow with test database
func TestAgentHandler_CreateAgent(t *testing.T) {
    db := setupTestDB()
    defer db.Close()
    
    handler := NewHandler(agentService, ...)
    router.POST("/agents", handler.createAgent)
    
    resp := performRequest(router, "POST", "/agents", payload)
    assert.Equal(t, 201, resp.Code)
}
```

---

## Performance Considerations

### 1. Batch Queries
```go
// Use batch methods when available
agents := agentRepo.GetByAgentCodes([]string{...})
// Avoid N+1 queries
```

### 2. Eager Loading
```go
// Load related entities upfront
agent := agentRepo.GetWithChannels(agentID)
// vs. separate queries for bindings
```

### 3. Index Strategy
```
// Indexed fields for common queries:
- Agent.agent_code
- Agent.user_code
- Channel.channel_code
- Session.session_key
- ConversationRecord.trace_id (for distributed tracing)
```

### 4. Message Bus Buffer
```go
// Stream channel has larger buffer (1000) vs. others (100)
stream: make(chan *StreamChunk, 1000)
// Prevents real-time update loss during spikes
```

---

## Bounded Context Interaction Summary

```
User Domain
└─ Creates/manages Users
   └─ Produces: UserCode

    ↓ (userCode provided to other contexts)

Agent Domain
├─ Manages Agent lifetimes
├─ Consumes: UserCode (for scoping)
├─ Produces: AgentCode
└─ Interacts with:
   ├─ MCP Domain (via AgentMCPBinding)
   ├─ Channel Domain (Agent is assigned to Channels)
   └─ Session Domain (Session tied to Agent)

    ↓ (agentCode in session)

Channel Domain
├─ Manages Channel lifetimes
├─ Consumes: UserCode, AgentCode
├─ Produces: ChannelCode
└─ Interacts with:
   ├─ Session Domain (creates Sessions)
   └─ Message Bus (publishes user inputs)

    ↓ (input from Channel)

Session Domain
├─ Manages active conversation
├─ Stores runtime context
├─ Consumes: UserCode, AgentCode, ChannelCode
└─ Interacts with:
   ├─ Agent Domain (executes Agent)
   └─ Conversation Domain (records history)

    ↓ (execution happens)

Agent Orchestration Domain
├─ Runs agent logic (Eino/Langgraph)
├─ Invokes tools (from MCP or skills)
├─ Produces: responses & events
└─ Publishes:
   ├─ Outbound messages (via MessageBus → Channel)
   ├─ Stream updates (via MessageBus → UI)
   └─ Conversation records (via MessageBus → Conversation Domain)

    ↓ (async pub-sub)

Conversation Domain
├─ Persists message history
├─ Provides analytics
├─ Consumes: TraceID, conversational events
└─ Serves:
   ├─ History queries
   ├─ Token usage stats
   └─ Session playbacks
```

---

## Next Steps for DDD Implementation

1. **Define Ubiquitous Language** - Glossary of domain terms (already started in AGENTS.md, CLAUDE.md)
2. **Explicit Aggregate Boundaries** - Mark which entities belong together (anti-corruption layer tags)
3. **Domain Events** - Add AgentConfigured, ChannelActivated, etc. events
4. **Factory Methods** - Safe construction of aggregates with validation
5. **Value Object Factories** - Typed config creation (FeishuChannelConfig.New())
6. **Specification Pattern** - Complex queries as reusable specifications
7. **Event Sourcing** - For audit trail and replay capabilities

---

**Last Updated:** 2026-03-25  
**Quick Reference V1.0**
