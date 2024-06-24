package cache

import (
	"testing"
	"time"
)

// метод чтобы получить все значения, но при этом не поднимать их в приоритете
// и чтобы они не удалялись в get
func getAllItems(c *cache) ([]interface{}, []interface{}) {
	keys := make([]interface{}, 0)
	values := make([]interface{}, 0)

	elem := c.list.Front()

	for elem != nil {
		next := elem.Next()
		keys = append(keys, elem.Value.(*item).key)
		values = append(values, elem.Value.(*item).value)
		elem = next
	}

	return keys, values
}

func TestCapacity(t *testing.T) {
	_, err := NewCache(0)
	if err.Error() != "invalid capacity" {
		t.Error("expected to return 'invalid capacity'")
	}

	_, err = NewCache(-5)
	if err.Error() != "invalid capacity" {
		t.Error("expected to return 'invalid capacity'")
	}

	cache, err := NewCache(5)
	if err != nil {
		t.Error(err)
	}
	defer cache.StopTTLRemoval()

	cap := cache.Cap()
	if cap != 5 {
		t.Errorf("expected capacity 5, but got %d", cap)
	}
}

func TestClear(t *testing.T) {
	cache, err := NewCache(2)
	if err != nil {
		t.Error(err)
	}
	defer cache.StopTTLRemoval()

	cache.Add(1, "hi")
	cache.Add(2, "there")

	cache.Clear()
	res, ok := cache.Get(1)
	if ok {
		t.Errorf("expected the value to be deleted, but got %s, %t", res, ok)
	}

	res, ok = cache.Get(2)
	if ok {
		t.Errorf("expected the value to be deleted, but got %s, %t", res, ok)
	}
}

func TestRemove(t *testing.T) {
	cache, err := NewCache(1)
	if err != nil {
		t.Error(err)
	}
	defer cache.StopTTLRemoval()

	cache.Add(1, "hi")

	cache.Remove(1)
	res, ok := cache.Get(1)
	if ok {
		t.Errorf("expected the value to be deleted, but got %s, %t", res, ok)
	}
}

func TestCache(t *testing.T) {
	cache, err := NewCache(2)
	if err != nil {
		t.Error(err)
	}
	defer cache.StopTTLRemoval()

	cache.Add(1, "hi")
	cache.Add(2, "there")

	res, ok := cache.Get(1)
	if res != "hi" || !ok {
		t.Errorf("expected 'hi', true but got %s, %t", res, ok)
	}

	res, ok = cache.Get(2)
	if res != "there" || !ok {
		t.Errorf("expected 'there', true but got %s, %t", res, ok)
	}

	res, ok = cache.Get(3)
	if ok {
		t.Errorf("expected no value, but got %s, %t", res, ok)
	}

	cache.Add(2, "here")
	res, ok = cache.Get(2)
	if res != "here" || !ok {
		t.Errorf("expected 'here', true but got %s, %t", res, ok)
	}
}

func TestLRUCacheLRU(t *testing.T) {
	cache, err := NewCache(3)
	if err != nil {
		t.Error(err)
	}
	defer cache.StopTTLRemoval()

	cache.Add("A", 1) //самый старый элемент -> должен быть вытеснен
	cache.Add("B", 2)
	cache.Add("C", 3)
	cache.Add("D", 4)

	_, ok := cache.Get("A")
	if ok {
		t.Error("expected that value to be repataced")
	}

	res, ok := cache.Get("D")
	if res != 4 || !ok {
		t.Errorf("expected 4, true but got %d, %t", res, ok)
	}
}

func TestLRUCacheGetLRU(t *testing.T) {
	cache, err := NewCache(3)
	if err != nil {
		t.Error(err)
	}
	defer cache.StopTTLRemoval()

	cache.Add("A", 1)
	cache.Add("B", 2)
	cache.Add("C", 3)

	//"A" перемещается вначало -> самый старый элемент - "B"
	res, ok := cache.Get("A")
	if res != 1 || !ok {
		t.Errorf("expected 1, true but got %d, %t", res, ok)
	}

	//"B" должен вытесниться
	cache.Add("D", 4)

	_, ok = cache.Get("B")
	if ok {
		t.Error("expected that value to be repataced")
	}

	res, ok = cache.Get("D")
	if res != 4 || !ok {
		t.Errorf("expected 4, true but got %d, %t", res, ok)
	}
}

func TestTTLRemoval(t *testing.T) {
	cache, err := NewCache(4)
	if err != nil {
		t.Error(err)
	}
	defer cache.StopTTLRemoval()

	cache.AddWithTTL(1, "A", 1*time.Millisecond)
	cache.AddWithTTL(2, "B", 1*time.Millisecond)
	cache.AddWithTTL(3, "C", 1*time.Millisecond)
	cache.AddWithTTL(4, "D", 1*time.Millisecond)

	keys, values := getAllItems(cache)
	if len(keys) != 4 || len(values) != 4 {
		t.Errorf("expected keys and values to have 4 elem, but got keys: %v values: %v", keys, values)
	}

	time.Sleep(2 * time.Second)

	keys, values = getAllItems(cache)
	if len(keys) != 0 || len(values) != 0 {
		t.Errorf("expected keys and values to have 0 elem, but got keys: %v values: %v", keys, values)

	}

}

func TestGetAfterTTL(t *testing.T) {
	cache, err := NewCache(3)
	if err != nil {
		t.Error(err)
	}
	defer cache.StopTTLRemoval()

	cache.Add(1, "some line")
	cache.AddWithTTL(2, "another line", 1*time.Millisecond)

	res, ok := cache.Get(2)
	if res != "another line" || !ok {
		t.Errorf("expected 'hi', true but got %s, %t", res, ok)
	}

	time.Sleep(2 * time.Millisecond)

	res, ok = cache.Get(2)
	if ok {
		t.Errorf("expected element to be removed by now, but got %s, %t", res, ok)
	}

}
