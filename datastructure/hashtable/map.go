package hashtable

const KResizingWork uint64 = 128
const KMaxLoadFactor uint64 = 8

type HMap struct {
	t1          HTab
	t2          HTab
	resizingPos uint32
}

func (h *HMap) Lookup(key *HNode, cmp func(*HNode, *HNode) bool) *HNode {
	h.HelpResizing()
	from := h.t1.LookUp(key, cmp)
	if from == nil {
		from = h.t2.LookUp(key, cmp)
	}
	if from != nil {
		return *from
	}
	return nil
}

// HelpResizing 为避免服务器停滞时间过长，不一次性移动所有的东西，保留两个hashtable并逐渐在它们之间移动节点
func (h *HMap) HelpResizing() {
	if h.t2.tab == nil {
		return
	}

	nwork := uint64(0)
	for nwork < KResizingWork && h.t2.size > 0 {
		from := h.t2.tab[h.resizingPos]
		if *from == nil {
			h.resizingPos++
			continue
		}

		h.t1.Insert(h.t2.Detach(from))
		nwork++
	}

	if h.t2.size == 0 {
		h.t2 = HTab{}
	}
}

func (h *HMap) Insert(node *HNode) {
	if h.t1.tab == nil {
		h.t1 = *InitHashTable(4)
	}
	h.t1.Insert(node)

	if h.t2.tab == nil {
		loadFactor := h.t1.size / (h.t1.mask + 1)
		if loadFactor > KMaxLoadFactor {
			h.StartResizing()
		}
	}
	h.HelpResizing()
}

func (h *HMap) StartResizing() {
	if h.t2.tab != nil {
		panic("t2 not empty.")
	}
	h.t2 = h.t1
	h.t1 = *InitHashTable((h.t1.mask + 1) * 2)
	h.resizingPos = 0
}

func (h *HMap) Pop(key *HNode, cmp func(*HNode, *HNode) bool) *HNode {
	h.HelpResizing()

	from := h.t1.LookUp(key, cmp)
	if from != nil {
		return h.t1.Detach(from)
	}

	from = h.t2.LookUp(key, cmp)
	if from != nil {
		return h.t2.Detach(from)
	}

	return nil
}
