package context

import (
	"fmt"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	ctx := New()
	if ctx == nil {
		t.Fatal("New() returned nil")
	}
	if ctx.Len() != 0 {
		t.Errorf("expected empty context, got %d items", ctx.Len())
	}
}

func TestPutGet(t *testing.T) {
	ctx := New()

	// 测试基本类型
	ctx.Put("string", "hello")
	ctx.Put("int", 42)
	ctx.Put("float", 3.14)
	ctx.Put("bool", true)

	// 测试获取
	tests := []struct {
		key      string
		expected interface{}
	}{
		{"string", "hello"},
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
	}

	for _, tt := range tests {
		val, ok := ctx.Get(tt.key)
		if !ok {
			t.Errorf("Get(%s) returned false", tt.key)
			continue
		}
		if val != tt.expected {
			t.Errorf("Get(%s) = %v, want %v", tt.key, val, tt.expected)
		}
	}
}

func TestGetNotFound(t *testing.T) {
	ctx := New()
	val, ok := ctx.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) returned true")
	}
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}

func TestDelete(t *testing.T) {
	ctx := New()
	ctx.Put("key", "value")

	if !ctx.Has("key") {
		t.Error("Has(key) returned false before delete")
	}

	ctx.Delete("key")

	if ctx.Has("key") {
		t.Error("Has(key) returned true after delete")
	}
}

func TestKeys(t *testing.T) {
	ctx := New()
	ctx.Put("a", 1)
	ctx.Put("b", 2)
	ctx.Put("c", 3)

	keys := ctx.Keys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}

	// 检查所有键都存在
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for _, expected := range []string{"a", "b", "c"} {
		if !keyMap[expected] {
			t.Errorf("expected key %s not found", expected)
		}
	}
}

func TestClone(t *testing.T) {
	ctx := New()
	ctx.Put("key1", "value1")
	ctx.Put("key2", 42)

	cloned := ctx.Clone()

	// 检查克隆的值
	val1, ok := cloned.Get("key1")
	if !ok || val1 != "value1" {
		t.Error("clone failed to copy key1")
	}

	val2, ok := cloned.Get("key2")
	if !ok || val2 != 42 {
		t.Error("clone failed to copy key2")
	}

	// 修改原始上下文不应影响克隆
	ctx.Put("key3", "new")
	_, ok = cloned.Get("key3")
	if ok {
		t.Error("changes to original affected clone")
	}

	// 修改克隆不应影响原始
	cloned.Put("key4", "clone_only")
	_, ok = ctx.Get("key4")
	if ok {
		t.Error("changes to clone affected original")
	}
}

func TestHas(t *testing.T) {
	ctx := New()
	ctx.Put("exists", true)

	if !ctx.Has("exists") {
		t.Error("Has(exists) returned false")
	}
	if ctx.Has("notexists") {
		t.Error("Has(notexists) returned true")
	}
}

func TestLen(t *testing.T) {
	ctx := New()
	if ctx.Len() != 0 {
		t.Errorf("expected 0, got %d", ctx.Len())
	}

	ctx.Put("a", 1)
	ctx.Put("b", 2)
	if ctx.Len() != 2 {
		t.Errorf("expected 2, got %d", ctx.Len())
	}

	ctx.Delete("a")
	if ctx.Len() != 1 {
		t.Errorf("expected 1, got %d", ctx.Len())
	}
}

func TestClear(t *testing.T) {
	ctx := New()
	ctx.Put("a", 1)
	ctx.Put("b", 2)
	ctx.Put("c", 3)

	ctx.Clear()

	if ctx.Len() != 0 {
		t.Errorf("expected 0 after clear, got %d", ctx.Len())
	}
	if ctx.Has("a") || ctx.Has("b") || ctx.Has("c") {
		t.Error("keys still exist after clear")
	}
}

func TestGetTyped(t *testing.T) {
	ctx := New()
	ctx.Put("string", "hello")

	val, err := ctx.GetTyped("string", nil)
	if err != nil {
		t.Errorf("GetTyped failed: %v", err)
	}
	if val != "hello" {
		t.Errorf("expected hello, got %v", val)
	}

	_, err = ctx.GetTyped("nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestGetT(t *testing.T) {
	ctx := New()
	ctx.Put("string", "hello")
	ctx.Put("int", 42)
	ctx.Put("float", 3.14)

	// 测试正确的类型
	str, err := GetT[string](ctx, "string")
	if err != nil {
		t.Errorf("GetT[string] failed: %v", err)
	}
	if str != "hello" {
		t.Errorf("expected hello, got %v", str)
	}

	num, err := GetT[int](ctx, "int")
	if err != nil {
		t.Errorf("GetT[int] failed: %v", err)
	}
	if num != 42 {
		t.Errorf("expected 42, got %v", num)
	}

	f, err := GetT[float64](ctx, "float")
	if err != nil {
		t.Errorf("GetT[float64] failed: %v", err)
	}
	if f != 3.14 {
		t.Errorf("expected 3.14, got %v", f)
	}

	// 测试错误的类型
	_, err = GetT[int](ctx, "string")
	if err == nil {
		t.Error("expected error for type mismatch")
	}

	// 测试不存在的键
	_, err = GetT[string](ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestMustGetT(t *testing.T) {
	ctx := New()
	ctx.Put("string", "hello")

	str := MustGetT[string](ctx, "string")
	if str != "hello" {
		t.Errorf("expected hello, got %v", str)
	}

	// 测试panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nonexistent key")
		}
	}()

	MustGetT[string](ctx, "nonexistent")
}

func TestConcurrentAccess(t *testing.T) {
	ctx := New()
	var wg sync.WaitGroup

	// 并发写入
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", n)
			ctx.Put(key, n)
		}(i)
	}

	// 并发读取
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", n)
			ctx.Get(key)
		}(i)
	}

	wg.Wait()

	if ctx.Len() != 100 {
		t.Errorf("expected 100 items, got %d", ctx.Len())
	}
}

// 基准测试
func BenchmarkPut(b *testing.B) {
	ctx := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Put("key", i)
	}
}

func BenchmarkGet(b *testing.B) {
	ctx := New()
	ctx.Put("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Get("key")
	}
}

func BenchmarkGetT(b *testing.B) {
	ctx := New()
	ctx.Put("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetT[string](ctx, "key")
	}
}

func BenchmarkClone(b *testing.B) {
	ctx := New()
	for i := 0; i < 100; i++ {
		ctx.Put(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Clone()
	}
}

func BenchmarkConcurrentPutGet(b *testing.B) {
	ctx := New()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				ctx.Put("key", i)
			} else {
				ctx.Get("key")
			}
			i++
		}
	})
}
