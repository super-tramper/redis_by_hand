package hashtable

type HNode struct {
	Next  *HNode
	HCode uint64
}

type HTab struct {
	tab  []**HNode
	mask uint64 // hashcode长度最大为64位
	size uint64
}

func InitHashTable(n uint64) *HTab {
	if n < 0 || (n-1)&n != 0 {
		panic("illegal table size.")
	}
	h := &HTab{}
	h.tab = make([]**HNode, n)
	h.mask = n - 1
	return h
}

func (h *HTab) Insert(node *HNode) {
	pos := node.HCode & h.mask
	next := (**HNode)(h.tab[pos])
	node.Next = *next
	h.tab[pos] = &node
	h.size++
}

func (h *HTab) LookUp(key *HNode, cmp func(*HNode, *HNode) bool) **HNode {
	if h.tab == nil {
		return nil
	}

	pos := key.HCode & h.mask
	from := h.tab[pos]

	for from != nil {
		if cmp(*from, key) {
			return from
		}
		from = &(*from).Next
	}

	return nil
}

// Detach 从单链表中删除一个节点
func (h *HTab) Detach(from **HNode) *HNode {
	node := *from
	*from = (*from).Next
	h.size--
	return node
}
