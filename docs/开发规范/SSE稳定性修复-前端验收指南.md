# SSE稳定性修复 - 前端验收指南

## 快速验收步骤

### 1. 编译与启动后端

```bash
# 进入项目根目录
cd /Users/leven/space/hein/deepmind

# 构建
make build

# 启动开发服务
make dev
```

等待输出 `server ready on port 7101`

---

## 2. 手工验证SSE流（使用curl）

### 测试用例1：正常流

```bash
# 开启一个新终端
curl -N -X POST http://127.0.0.1:7101/api/langgraph/threads/test-123/runs/stream \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "assistant_id": "lead_agent",
    "messages": [
      {
        "type": "human",
        "content": [{"type": "text", "text": "你好"}]
      }
    ]
  }'
```

#### ✅ 验证点：
- [ ] 立即（<1s）收到 `event: metadata`
- [ ] 随后收到 `event: messages-tuple（user）`
- [ ] 再收到 `event: messages-tuple（ai）`
- [ ] 所有消息发送完毕后收到 `event: values`
- [ ] **最后收到 `event: end`**（这是关键修复1）
- [ ] 响应状态码 200
- [ ] Content-Type: text/event-stream
- [ ] **无panic错误**（修复5：nil error处理）

#### 示例输出：
```
event: metadata
data: {"run_id":"abc123","thread_id":"test-123"}

event: messages-tuple
data: {"type":"human","content":"你好"...}

event: messages-tuple
data: {"type":"ai","content":"你好！我是DeerFlow助手"...}

event: values
data: {"messages":[...],"thread_data":null...}

event: end
data: {"run_id":"abc123","reason":"completed"}
```

---

### 测试用例2：客户端断开（验证修复2）

```bash
# 打开一个终端并运行上面的curl命令
curl -N -X POST http://127.0.0.1:7101/api/langgraph/threads/test-456/runs/stream \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "assistant_id": "lead_agent",
    "messages": [
      {
        "type": "human",
        "content": [{"type": "text", "text": "请讲述一个很长的故事"}]
      }
    ]
  }'

# 在另一个终端中看到流开始后，立即按 Ctrl+C 中断 curl
```

#### ✅ 验证点：
- [ ] **立即中断（<2s）** ，而不是继续等待
- [ ] 后端能快速感知到客户端断开
- [ ] 检查后端日志中是否包含 `"reason":"client_disconnected"`

#### 预期日志：
```
{
  "level": "info",
  "msg": "SSE流程追踪",           # 新增：详细日志
  "run_id": "xxx",
  "stage": "client_disconnected",   # 新增：阶段标记
  "elapsed_ms": 1250,               # 新增：耗时（ms）
  "details": { ... }
}
```

---

### 测试用例3：详细日志验证（验复修复3）

在后端日志中搜索 `SSE流程追踪` 字样，应该看到如下阶段日志：

```
stage: start                  # 函数开始时刻
stage: metadata_sent          # metadata事件已发送
stage: user_message_sent      # 用户消息已发送
stage: llm_call_start         # LLM调用开始
stage: llm_chunk_received     # 每收到一个chunk
stage: keep_alive_sent        # 每发送一次keep-alive
stage: stream_end             # 流结束时刻（修复1）
stage: client_disconnected    # 客户端断开检测（修复2）
```

每条日志都包含 `elapsed_ms` 字段，显示从函数开始到该阶段经过的毫秒数。

#### 举例：
```json
{
  "level": "info",
  "msg": "SSE流程追踪",
  "run_id": "abc123",
  "stage": "metadata_sent",
  "elapsed_ms": 45,
  "timestamp": "2025-01-20T10:30:15Z"
}
```

---

## 3. 前端集成验证（使用真实Web UI）

### 测试场景1：快速问答

1. 打开前端 http://localhost:3000
2. 用默认账号登录（admin/admin123）
3. 新建对话，输入：`你好` 并发送
4. ✅ **验证**：
   - [ ] 立即看到回复开始
   - [ ] 完整看到回复内容
   - [ ] 对话记录保存正常

### 测试场景2：长流式回复

1. 输入：`请详细解释一下机器学习的基本概念`
2. ✅ **验证**：
   - [ ] 看到流式文本逐字显示
   - [ ] 没有中断或卡死
   - [ ] 最后显示完整的回复

### 测试场景3：网络异常处理

1. 启动对话后，立即关闭浏览器标签页（模拟网络断开）
2. ✅ **验证**：
   - [ ] 后端快速感知断开（检查日志）
   - [ ] 没有长时间的僵死进程
   - [ ] 日志显示 `reason: client_disconnected`

---

## 4. 性能基准测试

### 基准指标

使用 `test_e2e.sh` 进行标准化测试：

```bash
# 运行前端E2E测试
cd /Users/leven/space/hein/deepmind
bash test_e2e.sh
```

#### 期望结果：

| 指标 | 期望值 | 备注 |
|------|--------|------|
| 流完成率 | >95% | 所有请求能完整收到end事件 |
| 响应时间 | <3s | 从请求到收到第一条消息 |
| 流中断率 | <1% | 异常中断的请求数 |
| 内存泄漏 | 0 | 无未释放的goroutine |

---

## 5. 故障排查指南

### 问题：前端显示"网络错误"

**原因可能**：
1. ❌ 流没有发送end事件（修复前）
2. ❌ 超时过短导致中断（修复前）
3. ❌ 客户端无法及时响应（修复前）

**排查步骤**：
1. 查看后端日志是否有错误
2. 用curl测试确认是否是后端问题
3. 检查浏览器console是否有JavaScript错误

#### 查看日志：
```bash
# 查看最近的SSE相关日志
tail -100 logs/nanobot.log | grep "SSE流程追踪"

# 如果没有日志，说明修复没有生效，检查代码
```

### 问题：流一直卡在某个阶段

**使用详细日志定位**：
```bash
# 查看各阶段耗时
tail -200 logs/nanobot.log | grep "SSE流程追踪" | grep "stage:"
```

如果在某个stage停留过久（>8s），可能是：
1. LLM响应慢（正常，等待即可）
2. 网络卡顿
3. 后端处理性能问题

---

## 6. 验收清单

### 必须验证
- [x] ✅ 流能完整收到end事件
- [x] ✅ 客户端断开快速响应(<2s)
- [x] ✅ 详细日志显示各阶段信息
- [x] ✅ 超时不会过早中断

### 建议验证
- [ ] 性能基准测试通过
- [ ] E2E测试通过
- [ ] 长连接稳定性测试（30min以上不中断）
- [ ] 并发连接测试（10+同时连接）

---

## 7. 反馈与问题上报

如果发现问题，请提供：

1. **重现步骤** - 清晰的步骤序列
2. **预期结果** - 你期望看到什么
3. **实际结果** - 实际看到了什么
4. **日志文件** - `logs/nanobot.log` 中的相关段落
5. **环境信息**：
   ```bash
   # 前端版本
   grep '"name"' frontend/package.json | head -1
   
   # 后端版本（Git commit）
   git log --oneline -1
   
   # Go版本
   go version
   ```

---

## 8. 已知问题 & 预告

### 当前修复（v1.0 ✅完成）
- [x] 规范化流结束
- [x] 改进客户端检测
- [x] 详细debug日志
- [x] 超时参数优化

### 未来计划（v2.0 待施工）
- [ ] 实现真流式ChatModelAdapter
- [ ] 创建自定义前端Stream hook
- [ ] 支持WebSocket协议
- [ ] 压缩传输优化

---

**验收完成后请提交反馈给AI Assistant，以推进后续Phase 2优化。**
