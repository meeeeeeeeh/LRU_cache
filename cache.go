package cache

//в структуре элемента хранится:
//key - ключ чтобы можно было при удаление элемета взять последний с конца списка, удалить его из листа и сразу удалить его из мапы
// если бы его не было в структуре, то пришлось бы перебирать всю мапу в поиске нужного значения
//ttl - время жизни элемента (используется только для добавленных с дедлайном, во всех остальных случаях - zeroValue)

//в конструкторе структуры запускается горутина, которая каждую минуту просматривает все элементы
//и если в них есть с истеченным дедлайном, то они удаляются

//так же все элементы могут быть удалены по алгоритму lru если capacity переполнится
//так как элементы и с дедлайном и без хранятся в одной мапе и у них одна структура, что упрощает работу с ними

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type item struct {
	key   interface{}
	value interface{}
	ttl   time.Time
}

type cache struct {
	capacity int
	items    map[interface{}]*list.Element
	list     *list.List
	mu       sync.Mutex
	done     chan struct{}
}

type ICache interface {
	Cap() int
	Clear()
	Add(key, value interface{})
	AddWithTTL(key, value interface{}, ttl time.Duration)
	Get(key interface{}) (value interface{}, ok bool)
	Remove(key interface{})
}

func NewCache(cap int) (*cache, error) {
	if cap <= 0 {
		err := errors.New("invalid capacity")
		return nil, err
	}
	cache := &cache{
		capacity: cap,
		items:    make(map[interface{}]*list.Element),
		list:     list.New(),
		done:     make(chan struct{}),
	}
	go cache.deleteByTTL()
	return cache, nil
}

func (c *cache) deleteByTTL() {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			elem := c.list.Front()
			for elem != nil {
				next := elem.Next()
				if elem.Value.(*item).ttl.Before(time.Now()) {
					c.mu.Unlock()
					c.Remove(elem.Value.(*item).key)
					c.mu.Lock()
				}
				elem = next
			}
			c.mu.Unlock()
		case <-c.done:
			ticker.Stop()
			return
		}
	}
}

func (c *cache) Done() {
	c.done <- struct{}{}
}

func (c *cache) AddWithTTL(key, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if ok {
		if elem.Value.(*item).value != value {
			elem.Value.(*item).value = value
		}
		elem.Value.(*item).ttl = time.Now().Add(ttl)
		c.list.MoveToFront(elem)
	} else {
		if c.list.Len() >= c.capacity {
			c.deleteCacheItem()
		}
		newItem := &item{key: key, value: value, ttl: time.Now().Add(ttl)}
		item := c.list.PushFront(newItem)
		c.items[key] = item
	}
}

func (c *cache) Cap() int {
	return c.capacity
}

func (c *cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[interface{}]*list.Element)
	c.list.Init()
}

func (c *cache) Add(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if ok {
		if elem.Value.(*item).value != value {
			elem.Value.(*item).value = value
		}
		elem.Value.(*item).ttl = time.Time{}
		c.list.MoveToFront(elem)
	} else {
		if c.list.Len() >= c.capacity {
			c.deleteCacheItem()
		}
		newItem := &item{key: key, value: value}
		item := c.list.PushFront(newItem)
		c.items[key] = item
	}
}

func (c *cache) deleteCacheItem() {
	elem := c.list.Back()
	if elem != nil {
		c.list.Remove(elem)
		delete(c.items, elem.Value.(*item).key)
	}
}

func (c *cache) Get(key interface{}) (value interface{}, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, false
	}
	if elem.Value.(*item).ttl.Before(time.Now()) {
		c.Remove(elem.Value.(*item).key)
		return nil, false
	}
	c.list.MoveToFront(elem)
	return elem.Value.(*item).value, true
}

func (c *cache) Remove(key interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if ok {
		c.list.Remove(elem)
		delete(c.items, elem.Value.(*item).key)
	}
}
