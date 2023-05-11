package heap

type HeapItem struct {
	Val uint64
	Ref *int32
}

type Heap []HeapItem // 小顶堆

func HeapParent(i int32) int32 {
	return (i+1)/2 - 1
}

func HeapLeft(i int32) int32 {
	return i*2 + 1
}

func HeapRight(i int32) int32 {
	return i*2 + 2
}

// HeapUp 向上冒泡
func (h Heap) HeapUp(p int32) {
	t := h[p]
	for p > 0 && h[HeapParent(p)].Val > t.Val {
		h[p] = h[HeapParent(p)]
		*h[p].Ref = p
		p = HeapParent(p)
	}
	h[p] = t
	*h[p].Ref = p
}

// HeapDown 入堆
func (h Heap) HeapDown(p int32, l int32) {
	t := h[p]
	for {
		// 找出堆中Val最小的元素
		le := HeapLeft(p)
		ri := HeapRight(p)
		mp := int32(-1)
		mv := t.Val
		if le < l && h[le].Val < mv {
			mp = le
			mv = h[le].Val
		}
		if ri < l && h[ri].Val < mv {
			mp = ri
		}
		if mp == int32(-1) {
			break
		}
		// 将p处节点同最小值交换
		h[p] = h[mp]
		*h[p].Ref = p
		p = mp
	}
	h[p] = t
	*h[p].Ref = p
}

// HeapUpdate 更新一个位置
func (h Heap) HeapUpdate(p int32, l int32) {
	if p > 0 && h[HeapParent(p)].Val > h[p].Val {
		h.HeapUp(p)
	} else {
		h.HeapDown(p, l)
	}
}

func Empty(heap *Heap) bool {
	return heap == nil || len(*heap) == 0
}
