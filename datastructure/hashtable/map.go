package hashtable

const KResizingWork uint64 = 128
const KMaxLoadFactor uint64 = 8

type HMap struct {
	T1          HTab
	T2          HTab
	resizingPos uint32
}

func (h *HMap) Lookup(key *HNode, cmp func(*HNode, *HNode) bool) *HNode {
	h.HelpResizing()
	from := h.T1.LookUp(key, cmp)
	if from == nil {
		from = h.T2.LookUp(key, cmp)
	}
	if from != nil {
		return *from
	}
	return nil
}

// HelpResizing 为避免服务器停滞时间过长，不一次性移动所有的东西，保留两个hashtable并逐渐在它们之间移动节点
func (h *HMap) HelpResizing() {
	if h.T2.tab == nil {
		return
	}

	nwork := uint64(0)
	for nwork < KResizingWork && h.T2.size > 0 {
		from := h.T2.tab[h.resizingPos]
		if from == nil {
			h.resizingPos++
			continue
		}

		h.T1.Insert(h.T2.Detach(&from))
		nwork++
	}

	if h.T2.size == 0 {
		h.T2 = HTab{}
	}
}

func (h *HMap) Insert(node *HNode) {
	if h.T1.tab == nil {
		h.T1 = *InitHashTable(4)
	}
	h.T1.Insert(node)

	if h.T2.tab == nil {
		loadFactor := h.T1.size / (h.T1.mask + 1)
		if loadFactor > KMaxLoadFactor {
			h.StartResizing()
		}
	}
	h.HelpResizing()
}

func (h *HMap) StartResizing() {
	if h.T2.tab != nil {
		panic("T2 not empty.")
	}
	h.T2 = h.T1
	h.T1 = *InitHashTable((h.T1.mask + 1) * 2)
	h.resizingPos = 0
}

func (h *HMap) Pop(key *HNode, cmp func(*HNode, *HNode) bool) *HNode {
	h.HelpResizing()

	from := h.T1.LookUp(key, cmp)
	if from != nil {
		return h.T1.Detach(from)
	}

	from = h.T2.LookUp(key, cmp)
	if from != nil {
		return h.T2.Detach(from)
	}

	return nil
}

func (h *HMap) Size() uint64 {
	return h.T1.size + h.T2.size
}
