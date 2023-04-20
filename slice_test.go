package nolock_slice

import (
	"sync"
	"testing"
)

func checkSlice(src []int) (int, bool) {
	srcMap := make(map[int]int, len(src))
	for i, item := range src {
		srcMap[item]++
		if srcMap[item] > 1 {
			return i, false
		}
	}
	return -1, true
}

func TestNoLockSlice_Append(t *testing.T) {
	t.Run("多协程并发, 不超出cap", func(t *testing.T) {
		count := 10000
		wg := sync.WaitGroup{}
		wg.Add(count)

		nls := NewNoLockSlice[int](int64(count))
		for i := 0; i < count; i++ {
			go func(ii int) {
				defer wg.Done()

				nls.Append(ii)
			}(i)
		}
		wg.Wait()

		if nls.Len() != count {
			t.Errorf("nls.Len() = %d, want %d", nls.Len(), count)
			return
		}
		if _, ok := checkSlice(nls.Slice()); !ok {
			t.Errorf("Append() check error, slice存在重复数据")
			return
		}
	})

	t.Run("多协程并发, 超出cap", func(t *testing.T) {
		count := 10000
		wg := sync.WaitGroup{}
		wg.Add(count)

		nls := NewNoLockSlice[int](int64(10))
		for i := 0; i < count; i++ {
			go func(ii int) {
				defer wg.Done()

				nls.Append(ii)
			}(i)
		}
		wg.Wait()

		if nls.Len() != count {
			t.Errorf("nls.Len() = %d, want %d", nls.Len(), count)
			return
		}
		if _, ok := checkSlice(nls.Slice()); !ok {
			t.Errorf("Append() check error, slice存在重复数据")
			return
		}
	})
}
