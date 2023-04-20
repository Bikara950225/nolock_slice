package nolock_slice

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type sliceStruct[T any] struct {
	data []T
}

type meta struct {
	cap int64
	len int64
}

type NoLockSlice[T any] struct {
	metaData unsafe.Pointer
	slice    unsafe.Pointer
	l        sync.Mutex
}

func NewNoLockSlice[T any](len int64) *NoLockSlice[T] {
	slice := sliceStruct[T]{
		data: make([]T, len),
	}
	metaData := meta{cap: len, len: 0}

	return &NoLockSlice[T]{
		metaData: unsafe.Pointer(&metaData),
		slice:    unsafe.Pointer(&slice),
	}
}

func (s *NoLockSlice[T]) Append(item T) {
	for {
		metaData := (*meta)(atomic.LoadPointer(&s.metaData))
		if metaData.len+1 > metaData.cap {
			if s.lockAppend(item) {
				return
			}
			continue
		}

		newMeta := &meta{
			cap: metaData.cap, len: metaData.len + 1,
		}
		if atomic.CompareAndSwapPointer(&s.metaData, unsafe.Pointer(metaData), unsafe.Pointer(newMeta)) {
			slice := (*sliceStruct[T])(atomic.LoadPointer(&s.slice))
			slice.data[metaData.len] = item
			return
		}
	}
}

func (s *NoLockSlice[T]) Slice() []T {
	return (*sliceStruct[T])(atomic.LoadPointer(&s.slice)).data
}

func (s *NoLockSlice[T]) Len() int {
	return int((*meta)(atomic.LoadPointer(&s.metaData)).len)
}

func (s *NoLockSlice[T]) lockAppend(items ...T) bool {
	s.l.Lock()
	defer s.l.Unlock()

	metaData := (*meta)(atomic.LoadPointer(&s.metaData))
	if metaData.len+1 < metaData.cap {
		return false
	}

	slice := (*sliceStruct[T])(atomic.LoadPointer(&s.slice))

	// 模仿Golang的Append扩容规则
	newLen := int(metaData.len) + len(items)
	newCap := int(metaData.cap) * 2
	if newCap < newLen {
		newCap = newLen
	} else {
		if metaData.cap > 1024 {
			newCap = int(metaData.cap) + (int(metaData.cap) / 4)
		}
	}

	newSlice := sliceStruct[T]{
		data: make([]T, newCap),
	}
	copy(newSlice.data, slice.data)
	for i, item := range items {
		newSlice.data[int(metaData.len)+i] = item
	}

	atomic.StorePointer(&s.slice, unsafe.Pointer(&newSlice))
	atomic.StorePointer(&s.metaData, unsafe.Pointer(&meta{cap: int64(newCap), len: int64(newLen)}))

	return true
}
