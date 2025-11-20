package market

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"nofx/logger"
)

type WSMonitor struct {
	wsClient       *WSClient
	combinedClient *CombinedStreamsClient
	symbols        []string
	alertsChan     chan Alert
	batchSize      int
	filterSymbols  sync.Map // 使用sync.Map来存储需要监控的币种和其状态
	FilterSymbol   []string //经过筛选的币种
	klineDataMap3m sync.Map // 添加缺失的字段
	klineDataMap4h sync.Map // 添加缺失的字段
}
type SymbolStats struct {
	LastActiveTime   time.Time
	AlertCount       int
	VolumeSpikeCount int
	LastAlertTime    time.Time
	Score            float64 // 综合评分
}

var WSMonitorCli *WSMonitor

func NewWSMonitor(batchSize int) *WSMonitor {
	WSMonitorCli = &WSMonitor{
		wsClient:       NewWSClient(),
		combinedClient: NewCombinedStreamsClient(batchSize),
		alertsChan:     make(chan Alert, 1000),
		batchSize:      batchSize,
	}
	return WSMonitorCli
}

func (m *WSMonitor) Initialize(coins []string) error {
	log.Println("初始化WebSocket监控器...")
	// 获取交易对信息
	apiClient := NewAPIClient()
	// 如果不指定交易对，则使用market市场的所有交易对币种
	if len(coins) == 0 {
		exchangeInfo, err := apiClient.GetExchangeInfo()
		if err != nil {
			return err
		}
		// 筛选永续合约交易对 --仅测试时使用
		//exchangeInfo.Symbols = exchangeInfo.Symbols[0:2]
		for _, symbol := range exchangeInfo.Symbols {
			if symbol.Status == "TRADING" && symbol.ContractType == "PERPETUAL" && strings.ToUpper(symbol.Symbol[len(symbol.Symbol)-4:]) == "USDT" {
				m.symbols = append(m.symbols, symbol.Symbol)
				m.filterSymbols.Store(symbol.Symbol, true)
			}
		}
	} else {
		m.symbols = coins
	}

	log.Printf("找到 %d 个交易对", len(m.symbols))

	return nil
}

func (m *WSMonitor) Start(coins []string) {
	logger.Info("启动WebSocket实时监控...")
	// 初始化交易对
	err := m.Initialize(coins)
	if err != nil {
		logger.Errorf("❌ 初始化币种失败: %v", err)
		return
	}

	err = m.combinedClient.Connect()
	if err != nil {
		logger.Errorf("❌ 批量订阅流失败: %v", err)
		return
	}
	// 订阅逻辑已迁移至 kline 服务；此处不再订阅 K 线
}

func (m *WSMonitor) subscribeAll() error {
	// K线订阅已移除
	logger.Info("跳过K线订阅（由 kline 服务负责）")
	return nil
}

func (m *WSMonitor) handleKlineData(symbol string, ch <-chan []byte, _time string) {
	for data := range ch {
		var klineData KlineWSData
		if err := json.Unmarshal(data, &klineData); err != nil {
			logger.Errorf("解析Kline数据失败: %v", err)
			continue
		}
		m.processKlineUpdate(symbol, klineData, _time)
	}
}

func (m *WSMonitor) getKlineDataMap(_time string) *sync.Map {
	var klineDataMap *sync.Map
	if _time == "3m" {
		klineDataMap = &m.klineDataMap3m
	} else if _time == "4h" {
		klineDataMap = &m.klineDataMap4h
	} else {
		klineDataMap = &sync.Map{}
	}
	return klineDataMap
}
func (m *WSMonitor) processKlineUpdate(symbol string, wsData KlineWSData, _time string) {
	// 转换WebSocket数据为Kline结构
	kline := Kline{
		OpenTime:  wsData.Kline.StartTime,
		CloseTime: wsData.Kline.CloseTime,
		Trades:    wsData.Kline.NumberOfTrades,
	}
	kline.Open, _ = parseFloat(wsData.Kline.OpenPrice)
	kline.High, _ = parseFloat(wsData.Kline.HighPrice)
	kline.Low, _ = parseFloat(wsData.Kline.LowPrice)
	kline.Close, _ = parseFloat(wsData.Kline.ClosePrice)
	kline.Volume, _ = parseFloat(wsData.Kline.Volume)
	kline.High, _ = parseFloat(wsData.Kline.HighPrice)
	kline.QuoteVolume, _ = parseFloat(wsData.Kline.QuoteVolume)
	kline.TakerBuyBaseVolume, _ = parseFloat(wsData.Kline.TakerBuyBaseVolume)
	kline.TakerBuyQuoteVolume, _ = parseFloat(wsData.Kline.TakerBuyQuoteVolume)
	// 更新K线数据
	var klineDataMap = m.getKlineDataMap(_time)
	value, exists := klineDataMap.Load(symbol)
	var klines []Kline
	if exists {
		klines = value.([]Kline)

		// 检查是否是新的K线
		if len(klines) > 0 && klines[len(klines)-1].OpenTime == kline.OpenTime {
			// 更新当前K线
			klines[len(klines)-1] = kline
		} else {
			// 添加新K线
			klines = append(klines, kline)

			// 保持数据长度
			if len(klines) > 100 {
				klines = klines[1:]
			}
		}
	} else {
		klines = []Kline{kline}
	}

	klineDataMap.Store(symbol, klines)
}

func (m *WSMonitor) GetCurrentKlines(symbol string, duration string) ([]Kline, error) {
	// 对每一个进来的symbol检测是否存在内类 是否的话就订阅它
	value, exists := m.getKlineDataMap(duration).Load(symbol)
	if !exists {
		// 如果Ws数据未初始化完成时,单独使用api获取 - 兼容性代码 (防止在未初始化完成是,已经有交易员运行)
		apiClient := NewAPIClient()
		klines, err := apiClient.GetKlines(symbol, duration, 100)
		if err != nil {
			return nil, fmt.Errorf("获取%v分钟K线失败: %v", duration, err)
		}

		// 动态缓存进缓存
		m.getKlineDataMap(duration).Store(strings.ToUpper(symbol), klines)

		// 订阅 WebSocket 流 (使用简化的订阅字符串)
		subStr := strings.ToLower(symbol) + "@kline_" + duration
		subErr := m.combinedClient.subscribeStreams([]string{subStr})
		logger.Infof("动态订阅流: %v", subStr)
		if subErr != nil {
			logger.Warnf("警告: 动态订阅%v分钟K线失败: %v (使用API数据)", duration, subErr)
		}

		// ✅ FIX: 返回深拷贝而非引用
		result := make([]Kline, len(klines))
		copy(result, klines)
		return result, nil
	}

	// ✅ FIX: 返回深拷贝而非引用，避免并发竞态条件
	klines := value.([]Kline)
	result := make([]Kline, len(klines))
	copy(result, klines)
	return result, nil
}

func (m *WSMonitor) Close() {
	m.wsClient.Close()
	close(m.alertsChan)
}
