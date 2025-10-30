package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"nofx/config"
	"nofx/db"
	"nofx/manager"
	"strings"

	"github.com/gin-gonic/gin"
)

// Server HTTP API服务器
type Server struct {
	router        *gin.Engine
	traderManager *manager.TraderManager
	port          int
}

// NewServer 创建API服务器
func NewServer(traderManager *manager.TraderManager, port int) *Server {
	// 设置为Release模式（减少日志输出）
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 启用CORS
	router.Use(corsMiddleware())

	s := &Server{
		router:        router,
		traderManager: traderManager,
		port:          port,
	}

	// 设置路由
	s.setupRoutes()

	return s
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.GET("/health", s.handleHealth)

	// API路由组
	api := s.router.Group("/api")
	{
		// 竞赛总览
		api.GET("/competition", s.handleCompetition)

		// Trader列表
		api.GET("/traders", s.handleTraderList)

		// 指定trader的数据（使用query参数 ?trader_id=xxx）
		api.GET("/status", s.handleStatus)
		api.GET("/account", s.handleAccount)
		api.GET("/positions", s.handlePositions)
		api.GET("/decisions", s.handleDecisions)
		api.GET("/decisions/latest", s.handleLatestDecisions)
		api.GET("/statistics", s.handleStatistics)
		api.GET("/equity-history", s.handleEquityHistory)
		api.GET("/performance", s.handlePerformance)

		// 交易开关
		api.GET("/trading/enabled", s.handleGetTradingEnabled)
		api.POST("/trading/enabled", s.handleSetTradingEnabled)

		// 管理员：Trader CRUD
		api.GET("/admin/traders", s.handleAdminListTraders)
		api.POST("/admin/traders", s.handleAdminUpsertTrader)
		api.DELETE("/admin/traders", s.handleAdminDeleteTrader)
		api.POST("/admin/reload", s.handleAdminReload)
	}
}

// handleHealth 健康检查
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   c.Request.Context().Value("time"),
	})
}

// getTraderFromQuery 从query参数获取trader
func (s *Server) getTraderFromQuery(c *gin.Context) (*manager.TraderManager, string, error) {
	traderID := c.Query("trader_id")
	if traderID == "" {
		// 如果没有指定trader_id，返回第一个trader
		ids := s.traderManager.GetTraderIDs()
		if len(ids) == 0 {
			return nil, "", fmt.Errorf("没有可用的trader")
		}
		traderID = ids[0]
	}
	return s.traderManager, traderID, nil
}

// handleCompetition 竞赛总览（对比所有trader）
func (s *Server) handleCompetition(c *gin.Context) {
	comparison, err := s.traderManager.GetComparisonData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取对比数据失败: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, comparison)
}

// handleTraderList trader列表
func (s *Server) handleTraderList(c *gin.Context) {
	traders := s.traderManager.GetAllTraders()
	result := make([]map[string]interface{}, 0, len(traders))

	for _, t := range traders {
		result = append(result, map[string]interface{}{
			"trader_id":   t.GetID(),
			"trader_name": t.GetName(),
			"ai_model":    t.GetAIModel(),
		})
	}

	c.JSON(http.StatusOK, result)
}

// handleStatus 系统状态
func (s *Server) handleStatus(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	status := trader.GetStatus()
	c.JSON(http.StatusOK, status)
}

// handleAccount 账户信息
func (s *Server) handleAccount(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	log.Printf("📊 收到账户信息请求 [%s]", trader.GetName())
	account, err := trader.GetAccountInfo()
	if err != nil {
		log.Printf("❌ 获取账户信息失败 [%s]: %v", trader.GetName(), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取账户信息失败: %v", err),
		})
		return
	}

	log.Printf("✓ 返回账户信息 [%s]: 净值=%.2f, 可用=%.2f, 盈亏=%.2f (%.2f%%)",
		trader.GetName(),
		account["total_equity"],
		account["available_balance"],
		account["total_pnl"],
		account["total_pnl_pct"])
	c.JSON(http.StatusOK, account)
}

// handlePositions 持仓列表
func (s *Server) handlePositions(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	positions, err := trader.GetPositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取持仓列表失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, positions)
}

// handleDecisions 决策日志列表
func (s *Server) handleDecisions(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 获取所有历史决策记录（无限制）
	records, err := trader.GetDecisionLogger().GetLatestRecords(10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取决策日志失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, records)
}

// handleLatestDecisions 最新决策日志（最近5条，最新的在前）
func (s *Server) handleLatestDecisions(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	records, err := trader.GetDecisionLogger().GetLatestRecords(5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取决策日志失败: %v", err),
		})
		return
	}

	// 反转数组，让最新的在前面（用于列表显示）
	// GetLatestRecords返回的是从旧到新（用于图表），这里需要从新到旧
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}

	// 为每个决策记录添加AI模型信息
	aiModel := trader.GetAIModel()
	enhancedRecords := make([]map[string]interface{}, len(records))
	for i, record := range records {
		enhancedRecord := map[string]interface{}{
			"timestamp":       record.Timestamp,
			"cycle_number":    record.CycleNumber,
			"input_prompt":    record.InputPrompt,
			"cot_trace":       record.CoTTrace,
			"decision_json":   record.DecisionJSON,
			"account_state":   record.AccountState,
			"positions":       record.Positions,
			"candidate_coins": record.CandidateCoins,
			"decisions":       record.Decisions,
			"execution_log":   record.ExecutionLog,
			"success":         record.Success,
			"error_message":   record.ErrorMessage,
			"ai_model":        aiModel, // 添加AI模型信息
		}
		enhancedRecords[i] = enhancedRecord
	}

	c.JSON(http.StatusOK, enhancedRecords)
}

// handleStatistics 统计信息
func (s *Server) handleStatistics(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	stats, err := trader.GetDecisionLogger().GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取统计信息失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// handleEquityHistory 收益率历史数据
func (s *Server) handleEquityHistory(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 获取尽可能多的历史数据（几天的数据）
	// 每3分钟一个周期：10000条 = 约20天的数据
	records, err := trader.GetDecisionLogger().GetLatestRecords(10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取历史数据失败: %v", err),
		})
		return
	}

	// 构建收益率历史数据点
	type EquityPoint struct {
		Timestamp        string  `json:"timestamp"`
		TotalEquity      float64 `json:"total_equity"`      // 账户净值（wallet + unrealized）
		AvailableBalance float64 `json:"available_balance"` // 可用余额
		TotalPnL         float64 `json:"total_pnl"`         // 总盈亏（相对初始余额）
		TotalPnLPct      float64 `json:"total_pnl_pct"`     // 总盈亏百分比
		PositionCount    int     `json:"position_count"`    // 持仓数量
		MarginUsedPct    float64 `json:"margin_used_pct"`   // 保证金使用率
		CycleNumber      int     `json:"cycle_number"`
	}

	// 从AutoTrader获取初始余额（用于计算盈亏百分比）
	initialBalance := 0.0
	if status := trader.GetStatus(); status != nil {
		if ib, ok := status["initial_balance"].(float64); ok && ib > 0 {
			initialBalance = ib
		}
	}

	// 如果无法从status获取，且有历史记录，则从第一条记录获取
	if initialBalance == 0 && len(records) > 0 {
		// 第一条记录的equity作为初始余额
		initialBalance = records[0].AccountState.TotalBalance
	}

	// 如果还是无法获取，返回错误
	if initialBalance == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "无法获取初始余额",
		})
		return
	}

	var history []EquityPoint
	for _, record := range records {
		// TotalBalance字段实际存储的是TotalEquity
		totalEquity := record.AccountState.TotalBalance
		// TotalUnrealizedProfit字段实际存储的是TotalPnL（相对初始余额）
		totalPnL := record.AccountState.TotalUnrealizedProfit

		// 计算盈亏百分比
		totalPnLPct := 0.0
		if initialBalance > 0 {
			totalPnLPct = (totalPnL / initialBalance) * 100
		}

		history = append(history, EquityPoint{
			Timestamp:        record.Timestamp.Format("2006-01-02 15:04:05"),
			TotalEquity:      totalEquity,
			AvailableBalance: record.AccountState.AvailableBalance,
			TotalPnL:         totalPnL,
			TotalPnLPct:      totalPnLPct,
			PositionCount:    record.AccountState.PositionCount,
			MarginUsedPct:    record.AccountState.MarginUsedPct,
			CycleNumber:      record.CycleNumber,
		})
	}

	c.JSON(http.StatusOK, history)
}

// handlePerformance AI历史表现分析（用于展示AI学习和反思）
func (s *Server) handlePerformance(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 分析最近20个周期的交易表现
	performance, err := trader.GetDecisionLogger().AnalyzePerformance(20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("分析历史表现失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, performance)
}

// handleGetTradingEnabled 查询交易开关
func (s *Server) handleGetTradingEnabled(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	status := t.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"trader_id":       traderID,
		"trading_enabled": status["trading_enabled"],
	})
}

type setTradingReq struct {
	TraderID string `json:"trader_id"`
	Enabled  bool   `json:"enabled"`
}

// handleSetTradingEnabled 设置交易开关
func (s *Server) handleSetTradingEnabled(c *gin.Context) {
	var req setTradingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if req.TraderID == "" {
		// 允许不传则使用默认trader
		_, id, err := s.getTraderFromQuery(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.TraderID = id
	}
	if err := s.traderManager.SetTradingEnabled(req.TraderID, req.Enabled); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"trader_id": req.TraderID, "trading_enabled": req.Enabled})
}

// 管理端：列出DB中的traders
func (s *Server) handleAdminListTraders(c *gin.Context) {
	ctx := context.Background()
	list, err := db.ListTraders(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// 管理端：新增/更新 trader
func (s *Server) handleAdminUpsertTrader(c *gin.Context) {
	var body db.TraderDoc
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if body.TraderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trader_id 必填"})
		return
	}
	ctx := context.Background()
	if err := db.UpsertTrader(ctx, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// 管理端：删除 trader
func (s *Server) handleAdminDeleteTrader(c *gin.Context) {
	type req struct {
		TraderID string `json:"trader_id"`
	}
	var r req
	if err := c.ShouldBindJSON(&r); err != nil || r.TraderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trader_id 必填"})
		return
	}
	ctx := context.Background()
	if err := db.DeleteTrader(ctx, r.TraderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// 管理端：重载（从DB重新装载traders）
func (s *Server) handleAdminReload(c *gin.Context) {
	// 从MongoDB重新加载所有交易者
	ctx := context.Background()
	tradersFromDB, err := db.ListTraders(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("从数据库加载交易者失败: %v", err)})
		return
	}

	// 获取当前已加载的交易者
	currentTraders := s.traderManager.GetAllTraders()
	currentTraderIDs := make(map[string]bool)
	for _, trader := range currentTraders {
		currentTraderIDs[trader.GetID()] = true
	}

	// 创建数据库中的交易者ID映射
	dbTraderIDs := make(map[string]bool)
	for _, traderDoc := range tradersFromDB {
		dbTraderIDs[traderDoc.TraderID] = true
	}

	// 删除不在数据库中的交易者
	removedCount := 0
	for traderID := range currentTraderIDs {
		if !dbTraderIDs[traderID] {
			// 这个交易者在数据库中不存在，需要删除
			err := s.traderManager.RemoveTrader(traderID)
			if err != nil {
				log.Printf("❌ 删除交易者 %s 失败: %v", traderID, err)
				continue
			}
			removedCount++
		}
	}

	// 添加新的交易者
	addedCount := 0
	for _, traderDoc := range tradersFromDB {
		traderID := traderDoc.TraderID
		if !currentTraderIDs[traderID] {
			// 这是一个新的交易者，需要添加
			cfg := db.ToConfig(traderDoc)

			// 获取全局配置（从现有交易者中获取）
			var globalConfig *config.Config
			if len(currentTraders) > 0 {
				// 从现有交易者获取全局配置
				globalConfig = &config.Config{
					UseDefaultCoins:    true,                                                         // 默认值
					CoinPoolAPIURL:     "",                                                           // 从现有配置获取
					OITopAPIURL:        "",                                                           // 从现有配置获取
					MaxDailyLoss:       0.1,                                                          // 默认10%
					MaxDrawdown:        0.2,                                                          // 默认20%
					StopTradingMinutes: 60,                                                           // 默认60分钟
					Leverage:           config.LeverageConfig{BTCETHLeverage: 5, AltcoinLeverage: 5}, // 默认5倍杠杆
				}
			}

			err := s.traderManager.AddTrader(
				cfg,
				globalConfig.CoinPoolAPIURL,
				globalConfig.MaxDailyLoss,
				globalConfig.MaxDrawdown,
				globalConfig.StopTradingMinutes,
				globalConfig.Leverage,
			)
			if err != nil {
				log.Printf("❌ 添加新交易者 %s 失败: %v", traderID, err)
				continue
			}

			// 启动新交易者
			trader, err := s.traderManager.GetTrader(traderID)
			if err == nil {
				go trader.Run()
				log.Printf("✅ 已添加并启动新交易者: %s (%s)", traderDoc.Name, strings.ToUpper(traderDoc.AIModel))
				addedCount++
			}
		}
	}

	// 同步现有交易者的启停开关（根据DB）
	dbEnabled := make(map[string]bool)
	for _, td := range tradersFromDB {
		dbEnabled[td.TraderID] = td.Enabled
	}
	synced := 0
	for _, t := range s.traderManager.GetAllTraders() {
		id := t.GetID()
		if v, ok := dbEnabled[id]; ok {
			_ = s.traderManager.SetTradingEnabled(id, v)
			synced++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"message":       fmt.Sprintf("重载完成，新增 %d 个，删除 %d 个，同步开关 %d 个", addedCount, removedCount, synced),
		"added_count":   addedCount,
		"removed_count": removedCount,
		"synced_switch": synced,
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🌐 API服务器启动在 http://localhost%s", addr)
	log.Printf("📊 API文档:")
	log.Printf("  • GET  /api/competition      - 竞赛总览（对比所有trader）")
	log.Printf("  • GET  /api/traders          - Trader列表")
	log.Printf("  • GET  /api/status?trader_id=xxx     - 指定trader的系统状态")
	log.Printf("  • GET  /api/account?trader_id=xxx    - 指定trader的账户信息")
	log.Printf("  • GET  /api/positions?trader_id=xxx  - 指定trader的持仓列表")
	log.Printf("  • GET  /api/decisions?trader_id=xxx  - 指定trader的决策日志")
	log.Printf("  • GET  /api/decisions/latest?trader_id=xxx - 指定trader的最新决策")
	log.Printf("  • GET  /api/statistics?trader_id=xxx - 指定trader的统计信息")
	log.Printf("  • GET  /api/equity-history?trader_id=xxx - 指定trader的收益率历史数据")
	log.Printf("  • GET  /api/performance?trader_id=xxx - 指定trader的AI学习表现分析")
	log.Printf("  • GET  /health               - 健康检查")
	log.Println()

	return s.router.Run(addr)
}
