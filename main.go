package main

import (
	"context"
	"fmt"
	"log"
	"nofx/api"
	"nofx/config"
	"nofx/db"
	"nofx/manager"
	"nofx/pool"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║    🏆 AI模型交易竞赛系统 - Qwen vs DeepSeek               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 优先从MongoDB加载traders；若失败或为空，再加载配置文件
	_ = dbInit()

	var cfg *config.Config
	// 默认配置文件
	configFile := "config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	tradersFromDB, dbErr := loadTradersFromDB()
	if dbErr == nil && len(tradersFromDB) > 0 {
		cfg = &config.Config{Traders: tradersFromDB}
		// 其余全局项仍从文件读取（若存在）
		if fileCfg, err := config.LoadConfig(configFile); err == nil {
			cfg.UseDefaultCoins = fileCfg.UseDefaultCoins
			cfg.CoinPoolAPIURL = fileCfg.CoinPoolAPIURL
			cfg.OITopAPIURL = fileCfg.OITopAPIURL
			cfg.APIServerPort = fileCfg.APIServerPort
			cfg.MaxDailyLoss = fileCfg.MaxDailyLoss
			cfg.MaxDrawdown = fileCfg.MaxDrawdown
			cfg.StopTradingMinutes = fileCfg.StopTradingMinutes
			cfg.Leverage = fileCfg.Leverage
		} else {
			// 设置默认端口
			if cfg.APIServerPort == 0 {
				cfg.APIServerPort = 8080
			}
		}
		log.Printf("📋 从MongoDB加载traders: %d 个", len(cfg.Traders))
	} else {
		log.Printf("📋 加载配置文件: %s", configFile)
		var err error
		cfg, err = config.LoadConfig(configFile)
		if err != nil {
			log.Fatalf("❌ 加载配置失败: %v", err)
		}
	}

	log.Printf("✓ 配置加载成功，共%d个trader参赛", len(cfg.Traders))
	fmt.Println()

	// 设置是否使用默认主流币种
	pool.SetUseDefaultCoins(cfg.UseDefaultCoins)
	if cfg.UseDefaultCoins {
		log.Printf("✓ 已启用默认主流币种列表（BTC、ETH、SOL、BNB、XRP、DOGE、ADA、HYPE）")
	}

	// 设置币种池API URL
	if cfg.CoinPoolAPIURL != "" {
		pool.SetCoinPoolAPI(cfg.CoinPoolAPIURL)
		log.Printf("✓ 已配置AI500币种池API")
	}
	if cfg.OITopAPIURL != "" {
		pool.SetOITopAPI(cfg.OITopAPIURL)
		log.Printf("✓ 已配置OI Top API")
	}

	// 创建TraderManager
	traderManager := manager.NewTraderManager()

	// 添加所有trader
	for i, traderCfg := range cfg.Traders {
		log.Printf("📦 [%d/%d] 初始化 %s (%s模型)...",
			i+1, len(cfg.Traders), traderCfg.Name, strings.ToUpper(traderCfg.AIModel))

		err := traderManager.AddTrader(
			traderCfg,
			cfg.CoinPoolAPIURL,
			cfg.MaxDailyLoss,
			cfg.MaxDrawdown,
			cfg.StopTradingMinutes,
			cfg.Leverage, // 传递杠杆配置
		)
		if err != nil {
			log.Fatalf("❌ 初始化trader失败: %v", err)
		}
	}

	fmt.Println()
	fmt.Println("🏁 竞赛参赛者:")
	for _, traderCfg := range cfg.Traders {
		fmt.Printf("  • %s (%s) - 初始资金: %.0f USDT\n",
			traderCfg.Name, strings.ToUpper(traderCfg.AIModel), traderCfg.InitialBalance)
	}

	fmt.Println()
	fmt.Println("🤖 AI全权决策模式:")
	fmt.Printf("  • AI将自主决定每笔交易的杠杆倍数（山寨币最高%d倍，BTC/ETH最高%d倍）\n",
		cfg.Leverage.AltcoinLeverage, cfg.Leverage.BTCETHLeverage)
	fmt.Println("  • AI将自主决定每笔交易的仓位大小")
	fmt.Println("  • AI将自主设置止损和止盈价格")
	fmt.Println("  • AI将基于市场数据、技术指标、账户状态做出全面分析")
	fmt.Println()
	fmt.Println("⚠️  风险提示: AI自动交易有风险，建议小额资金测试！")
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止运行")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// 创建并启动API服务器
	apiServer := api.NewServer(traderManager, cfg.APIServerPort)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Printf("❌ API服务器错误: %v", err)
		}
	}()

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 启动所有trader
	traderManager.StartAll()

	// 等待退出信号
	<-sigChan
	fmt.Println()
	fmt.Println()
	log.Println("📛 收到退出信号，正在停止所有trader...")
	traderManager.StopAll()

	fmt.Println()
	fmt.Println("👋 感谢使用AI交易竞赛系统！")
}

func dbInit() error {
	ctx := context.Background()
	_, err := db.Connect(ctx)
	if err != nil {
		log.Printf("⚠️  MongoDB 未连接: %v (将回退到文件配置)", err)
	} else {
		log.Printf("✓ MongoDB 已准备就绪")
	}
	return err
}

func loadTradersFromDB() ([]config.TraderConfig, error) {
	ctx := context.Background()
	list, err := db.ListTraders(ctx)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	res := make([]config.TraderConfig, 0, len(list))
	for _, d := range list {
		res = append(res, db.ToConfig(d))
	}
	return res, nil
}
