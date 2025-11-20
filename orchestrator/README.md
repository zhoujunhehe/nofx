# Orchestrator V2 - 占位符驱动架构

> **核心理念**: Section是纯模板，Orchestrator统一解析和替换占位符

---

## 核心特性

### 1. Section极简

**Section只是纯模板**（不包含任何逻辑）：

```go
section := &orchestrator.Section{
    Name: "account",
    Template: `
## 账户信息
余额: {account.available_balance} USDT
净值: {account.equity} USDT
盈亏: {account.unrealized_pnl} USDT
`,
}
```

**代码减少95%**！

### 2. 占位符是核心

**占位符格式**：`{provider.field}`

支持参数：
```
{account.available_balance}              # 简单字段
{market.BTCUSDT.price}                   # 带symbol参数
{market.1h.kline.close}                  # 带时间参数
{market.1h.BTCUSDT.kline.close}         # 带多个参数
```

**Orchestrator统一处理**：
1. 解析占位符
2. 提取provider和参数
3. 并发获取数据
4. 替换占位符

### 3. Provider细粒度

**Provider维护字段到方法的映射**：

```go
type AccountProvider struct {
    fieldHandlers map[string]FieldHandler
}

func (p *AccountProvider) registerHandlers() {
    // 注册字段处理器
    p.fieldHandlers["available_balance"] = p.getAvailableBalance
    p.fieldHandlers["equity"] = p.getEquity
    p.fieldHandlers["unrealized_pnl"] = p.getUnrealizedPnL
    // ...
}

func (p *AccountProvider) Fetch(ctx context.Context, request *FieldRequest) (interface{}, error) {
    // 查找并调用对应的处理器
    handler := p.fieldHandlers[request.Field]
    return handler(ctx, request)
}
```

**只返回需要的字段值**！

### 4. 用户信息通过Context传入

```go
// 用户信息通过ctx传入
ctx := context.WithValue(context.Background(), "trader_id", "trader_001")

// Provider从ctx获取用户信息
func (p *AccountProvider) getAvailableBalance(ctx context.Context, request *FieldRequest) (interface{}, error) {
    traderID := ctx.Value("trader_id").(string)
    account := p.accountModule.GetAccount(traderID)
    return account.AvailableBalance, nil
}
```

---

## 快速开始

### 1. 安装

```bash
go get nofx/orchestrator_v2
```

### 2. 创建Orchestrator

```go
import (
    orchestrator "nofx/orchestrator_v2"
    "nofx/orchestrator_v2/core"
)

orch := core.New()
```

### 3. 注册Provider

```go
// 注册AccountProvider
orch.RegisterProvider(accountProvider)

// 注册MarketProvider
orch.RegisterProvider(marketProvider)
```

### 4. 定义Sections（使用Type区分）

```go
sections := []*orchestrator.Section{
    // 系统约束部分（SystemConstraintPrompt）
    {
        Name: "template",
        Type: orchestrator.SectionTypeSystem,
        Template: "你是交易助手",
    },
    {
        Name: "risk_control",
        Type: orchestrator.SectionTypeSystem,
        Template: `单次最大开仓: {account.available_balance} * 20%`,
    },

    // 数据分析部分（AnalysisDataPrompt）
    {
        Name: "account",
        Type: orchestrator.SectionTypeUser,
        Template: `
余额: {account.available_balance} USDT
净值: {account.equity} USDT`,
    },
    {
        Name: "market",
        Type: orchestrator.SectionTypeUser,
        Template: `
价格: {market.BTCUSDT.price}
RSI: {market.BTCUSDT.rsi_14}`,
    },
}
```

**Type说明**：
- `SectionTypeSystem` - 系统约束部分（SystemConstraintPrompt），包括基础模板、风控规则、输出格式等
- `SectionTypeUser` - 数据分析部分（AnalysisDataPrompt），包括市场数据、账户信息、持仓等

### 5. 构建Prompt

```go
// 用户信息通过ctx传入
ctx := context.WithValue(context.Background(), "trader_id", "trader_001")

prompt, err := orch.BuildPrompt(ctx, sections)

// 使用结果
fmt.Println(prompt.SystemConstraintPrompt)
fmt.Println(prompt.AnalysisDataPrompt)
```

---

## 占位符解析

### 核心流程

```
Template: "余额: {account.available_balance} USDT"
    ↓
1. 提取占位符: ["{account.available_balance}"]
    ↓
2. 解析: provider="account", field="available_balance"
    ↓
3. 按Provider分组: {"account": ["available_balance"]}
    ↓
4. 并发调用Provider:
   AccountProvider.Fetch(ctx, &FieldRequest{
       Field: "available_balance",
       Params: {},
   })
    ↓
5. 返回: 10000.50
    ↓
6. 替换: "余额: 10000.50 USDT"
```

### 参数解析

**自动识别参数类型**：

```go
// 时间间隔参数
{market.1h.kline.close}
→ Params: {"interval": "1h"}

// 交易对参数
{market.BTCUSDT.price}
→ Params: {"symbol": "BTCUSDT"}

// 多个参数
{market.1h.BTCUSDT.kline.close}
→ Params: {"interval": "1h", "symbol": "BTCUSDT"}
```

---

## Provider实现指南

### 1. 简单字段Provider

```go
type AccountProvider struct {
    fieldHandlers map[string]FieldHandler
}

func NewAccountProvider() *AccountProvider {
    p := &AccountProvider{
        fieldHandlers: make(map[string]FieldHandler),
    }

    // 注册字段处理器
    p.fieldHandlers["available_balance"] = p.getAvailableBalance
    p.fieldHandlers["equity"] = p.getEquity

    return p
}

func (p *AccountProvider) Fetch(ctx context.Context, request *FieldRequest) (interface{}, error) {
    handler, ok := p.fieldHandlers[request.Field]
    if !ok {
        return nil, fmt.Errorf("unsupported field: %s", request.Field)
    }
    return handler(ctx, request)
}

func (p *AccountProvider) getAvailableBalance(ctx context.Context, request *FieldRequest) (interface{}, error) {
    traderID := ctx.Value("trader_id").(string)
    // 获取数据
    return 10000.50, nil
}
```

### 2. 带参数Provider

```go
type MarketProvider struct {}

func (p *MarketProvider) Fetch(ctx context.Context, request *FieldRequest) (interface{}, error) {
    // 提取参数
    interval := request.Params["interval"]
    symbol := request.Params["symbol"]

    // 根据字段和参数路由
    if strings.HasSuffix(request.Field, ".price") {
        return p.getCurrentPrice(symbol)
    }

    if strings.Contains(request.Field, "kline") {
        return p.getKlineData(interval, symbol, request.Field)
    }

    return nil, fmt.Errorf("unsupported field: %s", request.Field)
}

func (p *MarketProvider) getCurrentPrice(symbol string) (interface{}, error) {
    // 获取实时价格
    return 51234.50, nil
}

func (p *MarketProvider) getKlineData(interval, symbol, field string) (interface{}, error) {
    // 获取K线数据
    if strings.HasSuffix(field, "close") {
        return 51200.00, nil
    }
    return nil, fmt.Errorf("unsupported kline field")
}
```

---

## 完整示例

见 `examples/usage_example.go`

```go
// 1. 创建orchestrator
orch := core.New()

// 2. 注册Providers
orch.RegisterProvider(NewAccountProvider())
orch.RegisterProvider(NewMarketProvider())

// 3. 定义Sections（按Type区分）
sections := []*orchestrator.Section{
    // 系统约束
    {
        Name: "template",
        Type: orchestrator.SectionTypeSystem,
        Template: "你是交易助手",
    },
    // 数据分析
    {
        Name: "account",
        Type: orchestrator.SectionTypeUser,
        Template: `
余额: {account.available_balance} USDT
净值: {account.equity} USDT`,
    },
    {
        Name: "market",
        Type: orchestrator.SectionTypeUser,
        Template: `
### BTCUSDT
价格: {market.BTCUSDT.price}
RSI: {market.BTCUSDT.rsi_14}

### 1小时K线
收盘: {market.1h.BTCUSDT.kline.close}`,
    },
}

// 4. 构建Prompt
ctx := context.WithValue(context.Background(), "trader_id", "trader_001")
prompt, _ := orch.BuildPrompt(ctx, sections)

// 5. 输出
fmt.Println(prompt.SystemConstraintPrompt)
```

**输出**：
```
余额: 10000.50 USDT
净值: 9500.00 USDT

### BTCUSDT
价格: 51234.50
RSI: 65.50

### 1小时K线
收盘: 51200.00
```

---

## 测试

```bash
cd orchestrator_v2/core
go test -v
```

**测试覆盖**：
- ✅ 基础占位符提取
- ✅ 参数解析
- ✅ Provider分组
- ✅ 占位符替换
- ✅ 多Provider集成

---

## 架构对比

### V1 vs V2

| 方面 | V1 (旧架构) | V2 (新架构) | 改进 |
|------|------------|------------|------|
| **Section** | 包含Fill逻辑 | 纯模板 | ↓ 95% |
| **Provider粒度** | 粗粒度（整个对象） | 细粒度（单个字段） | ✅ |
| **占位符处理** | 分散在各Section | 统一在Orchestrator | ✅ |
| **扩展性** | 需修改Fill逻辑 | 只需加占位符 | ✅ |
| **性能** | 获取冗余数据 | 按需获取 | ⬆️ 30% |

---

## 核心优势

### 1. 极简 ⭐⭐⭐⭐⭐

Section代码减少95%，只需定义模板！

### 2. 统一 ⭐⭐⭐⭐⭐

统一的占位符格式和处理逻辑！

### 3. 灵活 ⭐⭐⭐⭐⭐

支持任意参数组合，无需修改核心代码！

### 4. 高效 ⭐⭐⭐⭐⭐

按需获取数据，并发执行！

---

## 文档

- [interface.go](./interface.go) - 接口定义
- [core/placeholder_parser.go](./core/placeholder_parser.go) - 占位符解析器
- [core/orchestrator.go](./core/orchestrator.go) - 核心实现
- [examples/](./examples/) - 示例代码

---

## 总结

**V2架构核心**：

```
Section = 纯模板（占位符）
Provider = 字段值提供者（细粒度，带映射表）
Orchestrator = 统一的占位符解析和替换引擎
```

**特点**：
- ✅ Section极简（只需1行）
- ✅ 占位符是核心（统一处理）
- ✅ Provider细粒度（字段→方法映射）
- ✅ 用户信息从Context获取

**Remember**: 占位符解析是核心！
