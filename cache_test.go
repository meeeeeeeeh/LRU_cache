package cache

import "testing"

// go test -cover ./... - просмотр покрытия
// go test -coverprofile=coverage.out ./... - сохранение покрытия в файл
// go tool cover -html=coverage.out - вывод подробного отчета в html

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
	defer cache.Done()

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

	cache.Done()
}

func TestRemove(t *testing.T) {
	cache, err := NewCache(1)
	if err != nil {
		t.Error(err)
	}

	cache.Add(1, "hi")

	cache.Remove(1)
	res, ok := cache.Get(1)
	if ok {
		t.Errorf("expected the value to be deleted, but got %s, %t", res, ok)
	}
	cache.Done()
}

func TestCache(t *testing.T) {
	cache, err := NewCache(2)
	if err != nil {
		t.Error(err)
	}

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

	cache.Done()
}

func TestLRUCache1(t *testing.T) {
	cache, err := NewCache(3)
	if err != nil {
		t.Error(err)
	}
	defer cache.Done()

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

func TestLRUCache2(t *testing.T) {
	cache, err := NewCache(3)
	if err != nil {
		t.Error(err)
	}
	defer cache.Done()

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
