package mapper

import (
	"fmt"
	"os"
	"testing"
	"time"

	"nofx/proxy/provider"
	"nofx/proxy/storage"
)

// TestUserIPMapperBasic 测试基本功能
func TestUserIPMapperBasic(t *testing.T) {
	// 使用临时文件
	tempFile := "/tmp/test_mapper_basic.json"
	defer func() {
		os.Remove(tempFile)
		os.Remove(tempFile + ".tmp")
	}()

	stor := storage.NewFileStorage(storage.FileStorageConfig{
		FilePath:      tempFile,
		FlushInterval: 100 * time.Millisecond, // 缩短刷新间隔便于测试
	})

	m := NewUserIPMapper(stor)

	// 设置3个IP
	ips := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.3"},
	}
	m.UpdateAvailableIPs(ips)

	// 分配10个用户
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("user_%d", i)
		ip, err := m.GetIPForUser(userID)
		if err != nil {
			t.Fatalf("分配IP失败: %v", err)
		}
		t.Logf("用户 %s -> IP %s", userID, ip)
	}

	// 验证负载均衡
	stats := m.GetAllMappings()
	for ip, users := range stats {
		t.Logf("IP %s: %d 用户", ip, len(users))
	}

	// 每个IP应该分配到大约相同数量的用户（3-4个）
	for ip, users := range stats {
		if len(users) < 2 || len(users) > 5 {
			t.Errorf("IP %s 的用户数不均衡: %d", ip, len(users))
		}
	}

	// 关闭storage，等待异步刷新完成
	stor.Close()
}

// TestUserIPMapperStability 测试老用户绑定的稳定性
func TestUserIPMapperStability(t *testing.T) {
	tempFile := "/tmp/test_mapper_stability.json"
	defer os.Remove(tempFile)

	stor := storage.NewFileStorage(storage.FileStorageConfig{FilePath: tempFile})
	defer stor.Close()

	m := NewUserIPMapper(stor)

	// 初始3个IP
	ips1 := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.3"},
	}
	m.UpdateAvailableIPs(ips1)

	// 分配10个用户
	initialMappings := make(map[string]string)
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("user_%d", i)
		ip, _ := m.GetIPForUser(userID)
		initialMappings[userID] = ip
	}

	t.Logf("=== 初始分配 (3个IP) ===")
	for userID, ip := range initialMappings {
		t.Logf("%s -> %s", userID, ip)
	}

	// 扩容到6个IP
	ips2 := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.3"},
		{IP: "192.168.1.4"}, // 新增
		{IP: "192.168.1.5"}, // 新增
		{IP: "192.168.1.6"}, // 新增
	}
	m.UpdateAvailableIPs(ips2)

	t.Logf("=== 扩容后 (6个IP) ===")

	// 验证老用户的映射不变
	changedCount := 0
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("user_%d", i)
		ip, _ := m.GetIPForUser(userID)

		if ip != initialMappings[userID] {
			changedCount++
			t.Errorf("❌ 老用户映射发生变化: %s, 原IP=%s, 现IP=%s",
				userID, initialMappings[userID], ip)
		}
	}

	if changedCount == 0 {
		t.Logf("✅ 所有老用户的IP绑定保持不变")
	} else {
		t.Fatalf("扩容导致 %d 个老用户映射变化", changedCount)
	}

	// 分配10个新用户
	t.Logf("=== 分配新用户 ===")
	for i := 10; i < 20; i++ {
		userID := fmt.Sprintf("user_%d", i)
		ip, _ := m.GetIPForUser(userID)
		t.Logf("新用户 %s -> IP %s", userID, ip)
	}

	// 查看最终分布
	t.Logf("=== 最终分布 (20用户, 6IP) ===")
	allMappings := m.GetAllMappings()
	for ip, users := range allMappings {
		t.Logf("IP %s: %d 用户", ip, len(users))
	}
}

// TestUserIPMapperPersistence 测试持久化
func TestUserIPMapperPersistence(t *testing.T) {
	tempFile := "/tmp/test_mapper_persist.json"
	defer os.Remove(tempFile)

	// 第一次：创建映射
	{
		stor := storage.NewFileStorage(storage.FileStorageConfig{FilePath: tempFile})
		m := NewUserIPMapper(stor)
		ips := []provider.ProxyIP{
			{IP: "192.168.1.1"},
			{IP: "192.168.1.2"},
		}
		m.UpdateAvailableIPs(ips)

		for i := 0; i < 5; i++ {
			userID := fmt.Sprintf("user_%d", i)
			ip, _ := m.GetIPForUser(userID)
			t.Logf("初次分配: %s -> %s", userID, ip)
		}

		// 手动保存
		if err := stor.Flush(); err != nil {
			t.Fatalf("保存失败: %v", err)
		}
		stor.Close()
	}

	// 第二次：从文件加载
	{
		stor := storage.NewFileStorage(storage.FileStorageConfig{FilePath: tempFile})
		defer stor.Close()
		m := NewUserIPMapper(stor)
		ips := []provider.ProxyIP{
			{IP: "192.168.1.1"},
			{IP: "192.168.1.2"},
		}
		m.UpdateAvailableIPs(ips)

		t.Logf("=== 从文件加载后 ===")

		// 验证老用户映射一致
		for i := 0; i < 5; i++ {
			userID := fmt.Sprintf("user_%d", i)
			ip, _ := m.GetIPForUser(userID)
			t.Logf("加载后: %s -> %s", userID, ip)
		}

		// 统计信息
		totalUsers, totalIPs, avg := m.GetStats()
		t.Logf("统计: %d 用户, %d IP, 平均 %.2f 用户/IP", totalUsers, totalIPs, avg)

		if totalUsers != 5 {
			t.Errorf("期望5个用户，实际 %d", totalUsers)
		}
	}
}

// TestGetUsersForIP 测试查询功能
func TestGetUsersForIP(t *testing.T) {
	m := NewUserIPMapper(nil) // 不使用持久化

	ips := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
	}
	m.UpdateAvailableIPs(ips)

	// 分配用户
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("user_%d", i)
		m.GetIPForUser(userID)
	}

	// 查询每个IP的用户
	t.Logf("=== IP绑定的用户列表 ===")
	for _, ip := range ips {
		users := m.GetUsersForIP(ip.IP)
		t.Logf("IP %s: %v", ip.IP, users)

		if len(users) == 0 {
			t.Errorf("IP %s 没有绑定用户", ip.IP)
		}
	}

	// 查询不存在的IP
	users := m.GetUsersForIP("192.168.1.999")
	if len(users) != 0 {
		t.Errorf("不存在的IP应该返回空列表")
	}
}

// TestRemoveUserMapping 测试删除映射
func TestRemoveUserMapping(t *testing.T) {
	m := NewUserIPMapper(nil)

	ips := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
	}
	m.UpdateAvailableIPs(ips)

	// 分配用户
	m.GetIPForUser("user_1")
	ip1, _ := m.GetIPForUser("user_1")
	t.Logf("user_1 绑定到: %s", ip1)

	// 删除映射
	err := m.RemoveUserMapping("user_1")
	if err != nil {
		t.Fatalf("删除映射失败: %v", err)
	}

	// 再次获取IP（应该分配新IP）
	ip2, _ := m.GetIPForUser("user_1")
	t.Logf("删除后重新分配: %s", ip2)

	// 可能分配到不同的IP（取决于负载均衡）
	totalUsers, _, _ := m.GetStats()
	if totalUsers != 1 {
		t.Errorf("期望1个用户，实际 %d", totalUsers)
	}
}

// TestLoadBalancing 测试负载均衡
func TestLoadBalancing(t *testing.T) {
	m := NewUserIPMapper(nil)

	ips := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.3"},
		{IP: "192.168.1.4"},
		{IP: "192.168.1.5"},
	}
	m.UpdateAvailableIPs(ips)

	// 分配100个用户
	for i := 0; i < 100; i++ {
		userID := fmt.Sprintf("user_%d", i)
		m.GetIPForUser(userID)
	}

	// 检查分布
	t.Logf("=== 100个用户在5个IP上的分布 ===")
	allMappings := m.GetAllMappings()

	minUsers := 100
	maxUsers := 0

	for ip, users := range allMappings {
		count := len(users)
		t.Logf("IP %s: %d 用户 (%.0f%%)", ip, count, float64(count)/100*100)

		if count < minUsers {
			minUsers = count
		}
		if count > maxUsers {
			maxUsers = count
		}
	}

	// 验证负载均衡效果（最大差异不超过2个用户）
	diff := maxUsers - minUsers
	t.Logf("最大用户数: %d, 最小用户数: %d, 差异: %d", maxUsers, minUsers, diff)

	if diff > 2 {
		t.Errorf("负载不够均衡，差异: %d (期望 ≤2)", diff)
	}
}

// TestGetCandidateIPsDegradation 测试所有IP被排除时的降级策略
func TestGetCandidateIPsDegradation(t *testing.T) {
	m := NewUserIPMapper(nil)

	ips := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
	}
	m.UpdateAvailableIPs(ips)

	// 测试：排除所有IP时，应返回候选IP（降级策略）
	excludeAll := []string{"192.168.1.1", "192.168.1.2"}
	candidates := m.getCandidateIPs(excludeAll)

	// 降级策略：即使所有IP都被排除，也应该返回候选IP
	if len(candidates) == 0 {
		t.Fatal("降级策略失败：所有IP被排除时应返回候选IP")
	}

	t.Logf("✅ 降级策略测试通过：所有IP被排除时返回了 %d 个候选IP", len(candidates))

	// 测试：部分排除时，应返回未排除的IP
	excludePartial := []string{"192.168.1.1"}
	candidates2 := m.getCandidateIPs(excludePartial)

	if len(candidates2) != 1 {
		t.Errorf("部分排除时应返回1个IP，实际: %d", len(candidates2))
	}

	if candidates2[0] != "192.168.1.2" {
		t.Errorf("应返回未排除的IP (192.168.1.2)，实际: %s", candidates2[0])
	}

	// 测试：不排除时，应返回所有IP
	candidates3 := m.getCandidateIPs(nil)
	if len(candidates3) != 2 {
		t.Errorf("不排除时应返回2个IP，实际: %d", len(candidates3))
	}
}
