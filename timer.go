package main

import (
	"redis_by_hand/datastructure/hashtable"
	"redis_by_hand/datastructure/heap"
	"time"
	"unsafe"
)

// NextTimerMs 下一计时器到期毫秒数
func NextTimerMs() uint32 {
	now := uint64(timeStamp())
	next := ^uint64(0)

	// 取堆顶元素时间
	if !heap.Empty(&g_data.heap) && g_data.heap[0].Val < next {
		next = g_data.heap[0].Val
	}

	// 无计时器
	if next == ^uint64(0) {
		return 10000
	}

	if next <= now {
		return 0
	}

	return uint32(next-now) / 1000
}

func ProcessTimers() {
	// 为poll解析添加额外10000ns
	now := uint64(timeStamp()) + 10000

	const kMaxWorks = 2000
	nWorks := int32(0)
	for !heap.Empty(&g_data.heap) && g_data.heap[0].Val < now {
		ent := (*Entry)(EntAddrByHeapIdx(g_data.heap[0].Ref))
		node := g_data.db.Pop(&ent.Node, hashtable.HNodeSame)
		if node == &ent.Node {
			panic("same node")
		}
		ent.Del()
		if nWorks >= kMaxWorks {
			nWorks++
			break
		}
		nWorks++
	}
}

func timeStamp() int64 {
	now := time.Now()
	return now.UnixMicro()
}

func EntAddrByHeapIdx(idx *int32) unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(idx)) - unsafe.Sizeof(hashtable.HNode{}))
}
