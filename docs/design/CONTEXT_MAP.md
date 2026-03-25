# Bounded Context Map - nanobot-go

## Context Relationships Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          HTTP API Layer (Gin)                               │
│  [ AuthMiddleware ] → [ Handler ] → [CorsMiddleware]                         │
└────────────────┬────────────────┬────────────────┬──────────────────────────┘
                 │                │                │
         ┌───────┴──────┐  ┌──────┴────────┐  ┌──┴────────────────┐
         │              │  │               │  │                   │
    ┌────▼──────┐  ┌────▼──┴──────┐  ┌─────▼──┴─────┐      ┌──────▼──────┐
    │   USER     │  │   AGENT      │  │   CHANNEL    │      │     SESSION  │
    │  DOMAIN    │  │   DOMAIN     │  │   DOMAIN     │      │    DOMAIN    │
    │            │  │              │  │              │      │              │
    │ ┌────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │      │ ┌──────────┐ │
    │ │ User   │ │  │ │Agent     │ │  │ │Channel   │ │      │ │Session   │ │
    │ │        │ │  │ │(AR)      │ │  │ │  (AR)    │ │      │ │  (AR)    │ │
    │ │+code   │ │  │ │+skills   │ │  │ │+type     │ │      │ │+key      │ │
    │ │+username│ │  │ │+tools    │ │  │ │+config   │ │      │ │+context  │ │
    │ │+auth   │ │  │ │+mcp      │ │  │ │+agent    │ │      │ │+user     │ │
    │ └────────┘ │  │ └──────────┘ │  │ └──────────┘ │      │ └──────────┘ │
    │            │  │   ┌────────┐  │  │              │      │              │
    │Repository│ │  │ ┌─┤AgentMCP│  │  │ Repository   │      │ Repository  │
    │ ┌──────┐ │  │ │ │Binding  │  │  │ ┌──────────┐  │      │ ┌────────┐ │
    │ │User  │ │  │ │ └────────┘  │  │ │Channel   │  │      │ │Session │ │
    │ │Repo  │ │  │ │              │  │ │Repo      │  │      │ │Repo    │ │
    │ └──────┘ │  │ │ Repository  │  │ │└──────────┘  │      │ └────────┘ │
    │ Service │ │  │ │ ┌────────┐ │  │ │ Service      │      │  Service   │
    │┌──────┐ │  │ │ │Agent   │ │  │ │┌──────────┐   │      │┌────────┐  │
    ││User  │ │  │ │ │Repo   │ │  │ ││Channel  │   │      ││Session │  │
    ││Svc   │ │  │ │ └────────┘ │  │ ││Svc      │   │      ││Svc     │  │
    │└──────┘ │  │ │            │  │ ││         │   │      │└────────┘  │
    └────────┘ │  │ │ Service  │ │  │ │└──────────┘   │      └──────────────┘
               │  │ │┌────────┐ │  │ │ Lookup:       │
               │  │ ││Agent   │ │  │ │ By Channel    │     ┌──────────────────┐
               │  │ ││Svc     │ │  │ │ Code         │ ←───→ CONVERSATION      │
               │  │ └────────┘ │  │ │ │ By User      │     │     DOMAIN       │
               │  │            │  │ │ │ Code        │     │                  │
               │  │ CodeService│  │ │ └──────────────┘   │ ┌──────────────┐  │
               │  │ - Generate  │ │ │                     │ │Conversation │  │
               │  │   AgentCode │ │ │                     │ │Record (AR)  │  │
               │  │ - Generate  │ │ │                     │ │+traceID     │  │
               │  │   ChannelCode
               │  └──────────────┘ │ │  ┌────────────────┐│ │+spanID    │  │
               │                    │ │  │ Pub/Sub Events│ │ │+content   │  │
               │                    │ │  │                │ │ │+role      │  │
               └────────────────────┘ │  │ Repository     │ │ └──────────┘  │
                                     │  │ ┌────────────┐  │ │ Repository   │
                                     │  │ │Conversation│  │ │ ┌──────────┐ │
                                     │  │ │Rec Repo    │  │ │ │Conversat │ │
                                     │  │ └────────────┘  │ │ │ionRecRepo│ │
                                     │ │ Service        │ │ │ └──────────┘ │
                                     │ │ ┌────────────┐  │ │ Service       │
                                     │ │ │Conversation│  │ │ ┌──────────┐ │
                                     │ │ │Svc         │  │ │ │Conversat │ │
                                     │ │ └────────────┘  │ │ │ionSvc    │ │
                                     └─┴─────────────────┘ │ └──────────┘ │
                                                            └──────────────┘
```

---

## Message Bus Integration (Cross-Cutting)

```
                    ┌─────────────────────────────┐
                    │      MESSAGE BUS            │
                    │  (pkg/bus/queue.go)         │
                    │                             │
                    │  ┌───────────────────────┐  │
                    │  │ Inbound Channel       │  │
                    │  │ (Device → Agent)      │  │
                    │  └───────────┬───────────┘  │
                    │              │              │
   ┌────────────┐   │   ┌────────────────────┐  │
   │  CHANNELS  │───┤   │ Outbound Channel   │  │
   │            │   │   │ (Agent → Device)   │  │
   │ ┌────────┐ │   │   └────────────┬──────┘  │
   │ │Feishu  │ │   │                │         │
   │ │WebSocket│───┤   ┌────────────────────┐  │
   │ │Matrix   │ │   │ Stream Channel       │  │
   │ │DingTalk │ │   │ (Real-time updates)  │  │
   │ └────────┘ │   │ └────────────┬──────┘  │
   └────────────┘   │              │         │
                    │   ┌────────────────────┐│
                    │   │ Task Events        ││
                    │   │ (Lifecycle hooks)  ││
                    │   └────────────────────┘│
                    │                         │
                    └────┬──────────────┬─────┘
                         │              │
        ┌────────────────┴──┐  ┌───────┴──────────────┐
        │                   │  │                      │
        ▼                   ▼  ▼                      ▼
    ┌─────────┐         ┌──────────────┐    ┌───────────────┐
    │AGENT    │         │ UI/WebSocket │    │ Task Manager  │
    │CORE     │         │ (SSE/Live)   │    │ (Background)  │
    │         │         │              │    │               │
    │ Eino    │         │ Dashboard    │    │ Cron Jobs     │
    │ Langgraph         │ Status       │    │ Task Queue    │
    └─────────┘         └──────────────┘    └───────────────┘
```

---

## LLM Provider & MCP Server Domain

```
┌──────────────────────────────────────────────────────────────────────┐
│                      LLM PROVIDER DOMAIN                             │
│                                                                      │
│  ┌──────────────────────┐                                           │
│  │ LLMProvider (AR)     │                                           │
│  │ +apiKey              │  SUPPORTS:                                │
│  │ +defaultModel        │  ┌─────────────────────┐                 │
│  │ +priority            │  │ ModelInfo (VO)      │                 │
│  │ +embeddingModels     │  │ +id                 │                 │
│  │                      │  │ +name               │                 │
│  │ - GetModels()        │  │ +maxTokens          │                 │
│  │ - AddModel()         │  └─────────────────────┘                 │
│  │ - SetDefault()       │                                           │
│  └──────────────────────┘  ┌─────────────────────┐                 │
│                            │EmbeddingModelInfo   │                 │
│  Repository Pattern        │ +id                 │                 │
│  ┌──────────────────────┐  │ +name               │                 │
│  │ ProviderRepository   │  │ +dimensions         │                 │
│  └──────────────────────┘  └─────────────────────┘                 │
│                                                                     │
│  Service Interface:                                                │
│  ┌──────────────────────┐                                          │
│  │ ProviderService      │                                          │
│  │ - CreateProvider()   │                                          │
│  │ - GetModels()        │                                          │
│  │ - RefreshModels()    │                                          │
│  └──────────────────────┘                                          │
└──────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                    MCP SERVER DOMAIN (Shared)                       │
│                                                                     │
│  ┌──────────────────────────────────────────┐                      │
│  │ MCPServer (AR)                           │                      │
│  │ +code                                    │                      │
│  │ +name                                    │                      │
│  │ +transportType (stdio/http/sse)          │                      │
│  │ +status (active/inactive/error)          │                      │
│  │ +capabilities                            │                      │
│  │                                          │                      │
│  │ - GetTools()                             │                      │
│  │ - IsToolEnabled()                        │                      │
│  └──────────────────────────────────────────┘                      │
│                           │                                        │
│                           ├─ CONTAINS ─┐                          │
│                           │             │                          │
│         ┌─────────────────▼──────────────▼──────────────┐          │
│         │ MCPToolModel (Child Entity)                 │          │
│         │ +id                                        │          │
│         │ +name                                      │          │
│         │ +description                              │          │
│         │ +inputSchema                              │          │
│         └─────────────────┬──────────────────────────┘          │
│                           │                                     │
│  Repository Pattern       │                                     │
│  ┌──────────────────┐    ├─ ASSOCIATES ────┐                  │
│  │ MCPServerRepo    │    │                 │                  │
│  ├─────────────────┐    │              ┌──▼──────────────────┐
│  │ MCPToolRepo     │    │              │ AgentMCPBinding    │
│  └──────────────────┘    │              │ (Cross-Domain)    │
│                          │              │ +agentID          │
│  Service Interface       │              │ +mcpServerID      │
│  ┌──────────────────┐    │              │ +enabledTools    │
│  │ MCPService       │    │              │ +isActive        │
│  │ - CreateServer() │    │              └──────────────────┘
│  │ - LoadTools()    │    │                   △
│  │ - GetTools()     │    └───────────────────┘
│  │ - UpdateTools()  │         JOINS
│  └──────────────────┘
└─────────────────────────────────────────────────────────────────────┘
```

---

## Agent Aggregate with MCP Bindings

```
┌────────────────────────────────────────────────────────┐
│              AGENT DOMAIN AGGREGATE                    │
│                                                        │
│  ┌────────────────────────────────────────────────┐   │
│  │ AGENT (Aggregate Root Entity)                  │   │
│  │ {AgentID, AgentCode, UserCode}                 │   │
│  │                                                │   │
│  │ Configuration:                                │   │
│  │ +name, +description                           │   │
│  │ +model, +maxTokens, +temperature              │   │
│  │ +maxIterations                                │   │
│  │                                                │   │
│  │ Markdown Content:                             │   │
│  │ +identityContent (IDENTITY.md)                │   │
│  │ +soulContent (SOUL.md)                        │   │
│  │ +agentsContent (AGENTS.md)                    │   │
│  │ +userContent (USER.md)                        │   │
│  │ +toolsContent (TOOLS.md)                      │   │
│  │                                                │   │
│  │ Capabilities:                                 │   │
│  │ +skillsList (JSON)                            │   │
│  │ +toolsList (JSON)                             │   │
│  │ +mcpList (JSON)                               │   │
│  │                                                │   │
│  │ + GetAvailableSkills()                        │   │
│  │ + GetAvailableTools()                         │   │
│  │ + GetAvailableMCPs()                          │   │
│  └────────────────────────────────────────────────┘   │
│                       │                                │
│                       ├─ CONTAINS (0..N) ────────┐    │
│                       │                          │    │
│        ┌──────────────▼──────────────────────────▼──┐ │
│        │ AgentMCPBinding (Child Entity)            │ │
│        │ {AgentID, MCPServerID}                    │ │
│        │                                           │ │
│        │ +enabledTools (JSON or null)              │ │
│        │ +isActive (boolean)                       │ │
│        │ +autoLoad (boolean)                       │ │
│        │                                           │ │
│        │ + GetEnabledTools()                       │ │
│        │ + IsToolEnabled(toolName)                 │ │
│        └───────────────────────────────────────────┘ │
│                                                      │
└──────────────────────────────────────────────────────┘
```

---

## Session Context & Conversation Record Flow

```
SESSION LIFECYCLE:

CREATE SESSION (internal/models/session.go)
├─ UserCode
├─ AgentCode
├─ ChannelCode
├─ SessionKey (unique)
├─ ExternalID (from channel)
├─ LastActiveAt (timestamp)
└─ Metadata (JSON)
        │
        ├──────────────┬─ PERSISTENT (Database)
        │              │
        │              └─ SessionRepository
        │                 └─ Queries: by key, by channel, by user
        │
        ├──────────────┬─ RUNTIME (In-Memory)
        │              │
        │              └─ pkg/session/Manager
        │                 ├─ Holds context.Context
        │                 ├─ Holds context.CancelFunc
        │                 └─ For cancellation/timeout

DURING SESSION:

ConversationRecord Events (pub on MessageBus)
└─ TraceID (request correlation)
   ├─ SpanID (specific operation)
   ├─ ParentSpanID (nesting)
   ├─ Role (user, assistant, system, tool)
   ├─ Content (message)
   ├─ EventType (message, tool_call, function_result, error)
   ├─ Token Usage
   │  ├─ PromptTokens
   │  ├─ CompletionTokens
   │  ├─ ReasoningTokens
   │  └─ CachedTokens
   └─ Stored in ConversationRecord table
      └─ Queries: by traceID, by sessionKey, by time range

ANALYTICS:

ConversationService (Statistics)
└─ TokenStats (total, daily trends)
├─ AgentDistribution (by agentCode)
├─ ChannelDistribution (by channelCode)
├─ RoleDistribution (user vs. assistant)
└─ SessionStats (active sessions, duration)
```

---

## Tenant Isolation Pattern

```
┌──────────────────────────────────────────────────────────┐
│            MULTI-TENANT ISOLATION via UserCode          │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │ User "alice" (UserCode = "u_abc123")              │ │
│  │                                                    │ │
│  │ Permissions:                                      │ │
│  │ ├─ Can see only agents where user_code=u_abc123  │ │
│  │ ├─ Can create channels for own agents            │ │
│  │ ├─ Sessions linked to own agents only            │ │
│  │ └─ Conversations visible only to own sessions    │ │
│  └────────────────────────────────────────────────────┘ │
│                                                          │
│  Entities with UserCode Index:                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │ Agent      (index: user_code)  ◄─┐               │ │
│  │ Channel    (index: user_code)  ◄─┼─ Owned by     │ │
│  │ Session    (index: user_code)  ◄─┼─ User        │ │
│  │ LLMProvider (index: user_code) ◄─┤  & Multi-   │ │
│  │ Conversation (index: user_code)◄─┘  Tenant    │ │
│  └────────────────────────────────────────────────────┘ │
│                                                          │
│  Global (Not Tenant-scoped):                            │
│  ┌────────────────────────────────────────────────────┐ │
│  │ MCPServer ◄─ Shared across all users             │ │
│  │           └─ But Agent-MCP binding still scoped  │ │
│  └────────────────────────────────────────────────────┘ │
│                                                          │
│  Authentication Flow:                                   │
│  ┌────────────────────────────────────────────────────┐ │
│  │ JWT Token                                         │ │
│  │ {userID: 42}                                      │ │
│  │     └──────────────┐                             │ │
│  │                    ▼                             │ │
│  │              Handler resolves                    │ │
│  │              userID → UserCode                   │ │
│  │                    ▼                             │ │
│  │          Query filter added:                     │ │
│  │          WHERE user_code = ?                     │ │
│  └────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

---

## Cross-Domain Communication

```
                       MESSAGE BUS
                   (pkg/bus/queue.go)
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
    ┌────────┐        ┌─────────┐        ┌──────────┐
    │INBOUND │        │OUTBOUND │        │  STREAM  │
    │MESSAGE │        │MESSAGE  │        │  CHUNK   │
    └────────┘        └─────────┘        └──────────┘
        │                  │                  │
        │ Device Input     │ Agent Response   │ Real-time
        │ from Channel     │ to Channel       │ Updates
        │                  │                  │
        ▼                  ▼                  ▼
    ┌────────────┐   ┌───────────┐    ┌──────────────┐
    │ AGENT CORE │   │ CHANNEL   │    │    UI Layer  │
    │ (Process)  │   │ ADAPTER   │    │ (WebSocket)  │
    │            │   │           │    │              │
    │ - Route    │   │ - Format  │    │ - Renders    │
    │ - Execute  │   │ - Send    │    │   Status     │
    │ - Reason   │───▶ to device │──▶ - Shows       │
    │            │   └───────────┘    │   progress   │
    └────────────┘                    └──────────────┘
        │
        │ Task Events
        ▼
    ┌──────────────┐
    │ TASK MANAGER │
    │              │
    │ - Publishing │
    │   task state │
    │ - Lifecycle  │
    │   hooks      │
    └──────────────┘
```

---

## Deployment/Structural Concerns

```
┌─────────────────────────────────────────────────────────────┐
│                     ANTI-PATTERNS / RISKS                   │
│─────────────────────────────────────────────────────────────│
│                                                              │
│  1. ANEMIC MODEL (Low Encapsulation)               [⚠️ MED]  │
│     ├─ Entities with mostly data, minimal logic             │
│     ├─ Business rules in Service instead of Aggregate       │
│     └─ Fix: Add factory methods, validation to roots        │
│                                                              │
│  2. JSON BLOB (Weak Typing)                        [⚠️ LOW]  │
│     ├─ Complex configs stored as JSON strings               │
│     ├─ No DB-level validation of structure                  │
│     └─ Fix: Create separate tables with FKs                 │
│                                                              │
│  3. MISSING DOMAIN EVENTS                          [⚠️ MED]  │
│     ├─ Pub-sub exists but not for domain events             │
│     ├─ No change audit trail at domain layer                │
│     └─ Fix: AgentConfigured, SessionStarted events          │
│                                                              │
│  4. IMPLICIT AGGREGATE BOUNDARIES                  [⚠️ LOW]  │
│     ├─ Unclear which entities belong to which aggregate     │
│     ├─ Is AgentMCPBinding part of Agent or separate?        │
│     └─ Fix: Explicit documentation + aggregate markers      │
│                                                              │
│  5. BLOB MESSAGE CONTENT                           [⚠️ LOW]  │
│     ├─ ConversationRecord.Content is untyped string         │
│     ├─ Different message types mixed without structure      │
│     └─ Fix: Create typed Message Value Object               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Recommended Evolution Path

```
Phase 1: SHORT TERM (Foundation)
├─ [ ] Document aggregate boundaries explicitly
├─ [ ] Add domain event publishing patterns
└─ [ ] Create bounded context context map (this doc)

Phase 2: MEDIUM TERM (Strengthening)
├─ [ ] Add value object factories with validation
├─ [ ] Extract JSON blobs to relational tables
├─ [ ] Implement specification pattern for queries
└─ [ ] Add domain events to all state changes

Phase 3: LONG TERM (Sophistication)
├─ [ ] Implement CQRS for complex queries
├─ [ ] Event sourcing for audit trail
├─ [ ] Anti-corruption layers for external systems
└─ [ ] Ubiquitous language glossary & patterns
```

---

**Context Map Version:** 1.0  
**Generated:** 2026-03-25
