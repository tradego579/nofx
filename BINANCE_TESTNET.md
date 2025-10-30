# 币安测试网使用指南

本指南说明如何在 NOFX 中启用和使用币安测试网（模拟盘）进行安全测试。

## 🎯 什么是测试网？

币安测试网（Binance Futures Testnet）是币安官方提供的模拟交易环境：
- ✅ **虚拟资金测试** - 使用模拟资金进行交易
- ✅ **真实市场数据** - 使用真实的市场价格和数据
- ✅ **零风险** - 所有交易都是虚拟的，不会损失真实资金
- ✅ **完整功能** - 支持所有交易功能和 API
- ✅ **免费使用** - 无需真实入金

## 📝 配置步骤

### 1. 获取测试网 API 密钥

#### 访问币安测试网
打开浏览器访问：**https://demo.binance.com/**

#### 如何通过合约模拟交易进行测试？
1. 登录您的[币安合约模拟交易](https://demo.binance.com/cn/futures) 账户。如果您没有模拟交易账户，请点击【创建】。
2. 点击右上角的账户图标进入【API 管理】页面，或点击[此处](https://demo.binance.com/cn/my/settings/api-management)访问【API 管理】页面。选择【创建 API】，并为您的 API 密钥输入名称。

#### 获取虚拟资金
- 测试网默认提供 **5000 USDT** 虚拟资金
- 如果需要更多，可以联系测试网管理员或在 Discord 申请

### 2. 配置 NOFX

#### 复制配置模板
```bash
cp config.json.example config.json
```

#### 编辑配置文件
打开 `config.json`，添加或修改币安测试网配置：

```json
{
  "traders": [
    {
      "id": "my_testnet_trader",
      "name": "My Testnet Trader",
      "ai_model": "deepseek",
      "exchange": "binance",
      "binance_api_key": "你的测试网API_KEY",
      "binance_secret_key": "你的测试网SECRET_KEY",
      "binance_testnet": true,
      "deepseek_key": "你的DeepSeek_API_KEY",
      "initial_balance": 10000,
      "scan_interval_minutes": 3
    }
  ],
  "leverage": {
    "btc_eth_leverage": 5,
    "altcoin_leverage": 5
  },
  "api_server_port": 8080
}
```

**关键配置说明：**
- `binance_testnet: true` - **启用测试网模式**
- `exchange: "binance"` - 使用币安交易所
- `initial_balance: 10000` - 设置为测试网虚拟资金数量

### 3. 启动 NOFX

#### 使用 Docker 启动（推荐）
```bash
./start.sh start --build
```

#### 手动启动
```bash
# 启动后端
go build -o nofx
./nofx

# 新开终端，启动前端
cd web
npm run dev
```

#### 验证测试网连接
启动后，检查日志输出，应该看到：
```
ℹ️  trader[0]: 使用币安测试网 (https://testnet.binancefuture.com) - 仅用于测试，不会产生真实交易
🔗 使用币安测试网: https://testnet.binancefuture.com
```

**✅ 如果看到以上信息，说明已成功连接到测试网！**

## 🔍 测试网 vs 主网对比

| 特性 | 测试网 | 主网 |
|------|--------|------|
| **资金** | 虚拟资金（免费获得） | 真实资金 |
| **交易** | 模拟交易，无真实盈亏 | 真实交易，有真实盈亏 |
| **风险** | 零风险 | 有风险 |
| **API 端点** | testnet.binancefuture.com | fapi.binance.com |
| **市场数据** | 真实市场价格 | 真实市场价格 |
| **手续费** | 虚拟费用 | 真实费用 |
| **适用场景** | 测试、调试、学习 | 正式交易 |

## 📊 测试功能清单

测试网支持以下所有功能：

- [x] **账户查询** - 查看余额、持仓
- [x] **开仓/平仓** - 多空双向交易
- [x] **杠杆交易** - 1x-125x 杠杆
- [x] **止损止盈** - 保护性订单
- [x] **市价单/限价单** - 各种订单类型
- [x] **技术指标** - RSI, MACD, EMA 等
- [x] **AI 决策** - 完整交易逻辑
- [x] **实时数据** - K线数据更新

## ⚠️ 注意事项

### 1. 测试网限制
- 测试网可能偶有不稳定（比主网稍差）
- 部分高级功能可能未完全开放
- 资金是虚拟的，无法提取

### 2. API 密钥管理
- 测试网 API 密钥与主网**完全独立**
- 不要在主网使用测试网密钥
- 测试完成后可以删除测试网 API 密钥

### 3. 配置检查
**启用测试网时：**
```bash
curl http://localhost:8080/health
```

检查配置是否正确：
```bash
curl http://localhost:8080/api/config | jq '.'
```

### 4. 日志监控
查看测试网连接日志：
```bash
./start.sh logs backend | grep -i "testnet\|币安"
```

## 🚀 完整测试流程

### 步骤 1：基本连接测试
1. 启动 NOFX
2. 检查日志中是否显示 "测试网" 字样
3. 访问 http://localhost:3000
4. 确认能正常加载仪表板

### 步骤 2：AI 决策测试
1. 等待第一个 AI 决策周期（3-5 分钟）
2. 查看决策日志：`cat decision_logs/my_testnet_trader/decisions.json | jq`
3. 确认 AI 能正常分析市场数据

### 步骤 3：交易功能测试
1. 观察 AI 是否会做出开仓决策
2. 检查测试网账户是否有虚拟持仓
3. 查看 Web 界面是否显示持仓信息
4. 等待平仓，验证整个交易流程

### 步骤 4：性能对比测试
如果你有主网账户，可以：
1. 配置两个 trader（一个测试网，一个主网）
2. 使用相同的 AI 配置
3. 对比两者的决策和表现
4. 分析差异并优化策略

## 🔧 常见问题

### Q1: 启动时报错 "invalid API key"
**A1**: 检查测试网 API Key 是否正确，注意：
- API Key 和 Secret Key 不要弄混
- 不要有多余的空格或换行
- 确认 API 密钥有 Futures 权限

### Q2: 测试网余额显示为 0
**A2**:
- 确认是在 testnet.binancefuture.com 创建的 API 密钥
- 检查 API 权限是否包含 "Enable Futures"
- 重启 NOFX 系统

### Q3: 无法连接到测试网
**A3**:
- 检查网络连接
- 确认币安测试网服务正常：https://status.binance.vision/
- 检查防火墙是否阻止了 testnet.binancefuture.com

### Q4: 测试网和主网 API 密钥可以同时使用吗？
**A4**: 不可以。测试网和主网的 API 密钥是独立的，需要分别配置。在 `config.json` 中：
- 测试网 trader：`"binance_testnet": true`
- 主网 trader：`"binance_testnet": false`

### Q5: 如何从测试网切换到主网？
**A5**: 只需修改 `config.json`：
```json
{
  "binance_testnet": false,  // 改为 false
  "binance_api_key": "主网API_KEY",
  "binance_secret_key": "主网SECRET_KEY"
}
```
重启 NOFX 即可使用主网进行真实交易。

## 📚 相关链接

- **币安测试网**: https://testnet.binance.vision
- **测试网 API 文档**: https://binance-docs.github.io/apidocs/testnet/futures/en/
- **NOFX 项目**: https://github.com/tinkle-community/nofx
- **测试网状态**: https://status.binance.vision/
- **开发者社区**: https://t.me/nofx_dev_community

## 💡 最佳实践

1. **始终先在测试网测试** - 所有新策略都应该先在测试网验证
2. **小资金开始** - 即使在主网，也建议从小资金开始
3. **记录测试结果** - 保存测试网的决策日志用于分析
4. **逐步增加复杂度** - 先测试基本功能，再测试高级特性
5. **对比主网表现** - 定期对比测试网和主网的表现差异

---

**🎉 现在你可以安全地在测试网上测试 NOFX 系统了！**

记住：**测试网 = 零风险的学习和测试环境**

如果遇到问题，欢迎在 [Telegram 社区](https://t.me/nofx_dev_community) 提问！