package avl

import (
	"fmt"
	"math/rand"
	"redis_by_hand/tools"
	"sort"
	"testing"
	"unsafe"
)

type Data struct {
	node Node
	val  uint32
}

type Container struct {
	root *Node
}

func add(c *Container, val uint32) {
	data := &Data{}
	data.node.Init()
	data.val = val

	if c.root == nil {
		c.root = &data.node
		return
	}

	cur := c.root
	for {
		var from **Node
		if val < (*Data)(unsafe.Pointer(cur)).val {
			from = &cur.Left
		} else {
			from = &cur.Right
		}
		if *from == nil {
			*from = &data.node
			data.node.Parent = cur
			c.root = AVLFix(&data.node)
			break
		}
		cur = *from
	}
}

func del(c *Container, val uint32) bool {
	cur := c.root
	for cur != nil {
		nodeVal := (*Data)(unsafe.Pointer(cur)).val
		if val == nodeVal {
			break
		}
		if val < nodeVal {
			cur = cur.Left
		} else {
			cur = cur.Right
		}
	}
	if cur == nil {
		return false
	}
	c.root = AVLDel(cur)
	return true
}

func avlVerify(parent *Node, node *Node) error {
	if node == nil {
		return nil
	}
	if node.Parent != parent {
		return fmt.Errorf("parent error")
	}
	err := avlVerify(node, node.Left)
	if err != nil {
		return err
	}
	err = avlVerify(node, node.Right)
	if err != nil {
		return err
	}

	if node.Cnt != 1+Cnt(node.Left)+Cnt(node.Right) {
		return fmt.Errorf("cnt error")
	}

	l := Depth(node.Left)
	r := Depth(node.Right)

	if !(l == r || l+1 == r || l == r+1) {
		return fmt.Errorf("balance error")
	}
	if node.Depth != 1+tools.Max(l, r) {
		return fmt.Errorf("depth error, expected %d, actual %d", node.Depth, 1+tools.Max(l, r))
	}

	val := (*Data)(unsafe.Pointer(node)).val
	if node.Left != nil {
		if node.Left.Parent != node {
			return fmt.Errorf("left children parent error")
		}
		if (*Data)(unsafe.Pointer(node.Left)).val > val {
			return fmt.Errorf("left children value greater than parent value")
		}
	}
	if node.Right != nil {
		if node.Right.Parent != node {
			return fmt.Errorf("right children parent error")
		}
		if (*Data)(unsafe.Pointer(node.Right)).val < val {
			return fmt.Errorf(
				"right children value error, expected: %d, got: %d",
				val,
				(*Data)(unsafe.Pointer(node.Right)).val)
		}
	}
	return nil
}

func extract(node *Node, extracted *[]uint32) {
	if node == nil {
		return
	}
	extract(node.Left, extracted)
	*extracted = append(*extracted, (*Data)(unsafe.Pointer(node)).val)
	extract(node.Right, extracted)
}

func dispose(c *Container) {
	for c.root != nil {
		//node := c.root
		c.root = AVLDel(c.root)
	}
}

func containerVerify(c *Container, ref *[]uint32) error {
	err := avlVerify(nil, c.root)
	if err != nil {
		return err
	}
	if Cnt(c.root) != uint32(len(*ref)) {
		return fmt.Errorf("container cnt error, expected: %d, got: %d", len(*ref), Cnt(c.root))
	}
	extracted := make([]uint32, 0)
	extract(c.root, &extracted)
	if len(extracted) != len(*ref) {
		return fmt.Errorf("extracted length error, expected: %d, got: %d", len(*ref), len(extracted))
	}
	for i := 0; i < len(*ref); i++ {
		if extracted[i] != (*ref)[i] {
			return fmt.Errorf("extracted content error, index: %d", i)
		}
	}
	return nil
}

func testInsert(sz uint32) error {
	for val := uint32(0); val < sz; val++ {
		c := Container{}
		ref := make([]uint32, 0)
		for i := uint32(0); i < sz; i++ {
			if i == val {
				continue
			}
			add(&c, i)
			ref = append(ref, i)
		}
		sort.Slice(ref, func(i, j int) bool {
			return ref[i] < ref[j]
		})
		err := containerVerify(&c, &ref)
		if err != nil {
			return err
		}

		add(&c, val)
		ref = append(ref, val)
		sort.Slice(ref, func(i, j int) bool {
			return ref[i] < ref[j]
		})
		err = containerVerify(&c, &ref)
		if err != nil {
			return err
		}
		dispose(&c)
	}
	return nil
}

func testInsertDup(sz uint32) error {
	for val := uint32(0); val < sz; val++ {
		c := Container{}
		ref := make([]uint32, 0)
		for i := uint32(0); i < sz; i++ {
			add(&c, i)
			ref = append(ref, i)
		}
		sort.Slice(ref, func(i, j int) bool {
			return ref[i] < ref[j]
		})
		err := containerVerify(&c, &ref)
		if err != nil {
			return err
		}

		add(&c, val)
		ref = append(ref, val)
		sort.Slice(ref, func(i, j int) bool {
			return ref[i] < ref[j]
		})
		err = containerVerify(&c, &ref)
		if err != nil {
			return err
		}
		dispose(&c)
	}
	return nil
}

func testRemove(sz uint32) error {
	for val := uint32(0); val < sz; val++ {
		c := Container{}
		ref := make([]uint32, 0)
		for i := uint32(0); i < sz; i++ {
			add(&c, i)
			ref = append(ref, i)
		}
		sort.Slice(ref, func(i, j int) bool {
			return ref[i] < ref[j]
		})
		err := containerVerify(&c, &ref)
		if err != nil {
			return fmt.Errorf("append verify error: %v", err)
		}

		if !del(&c, val) {
			return fmt.Errorf("testRemove del error")
		}
		for i, v := range ref {
			if v == val {
				if i < len(ref)-1 {
					ref = append(ref[:i], ref[i+1:]...)
				} else {
					ref = ref[:i]
				}
			}
		}
		err = containerVerify(&c, &ref)
		if err != nil {
			return fmt.Errorf("del verify error: %v", err)
		}
		dispose(&c)
	}
	return nil
}

func TestEmptyAVL(t *testing.T) {
	c := &Container{}
	ref := &[]uint32{}
	err := containerVerify(c, ref)
	if err != nil {
		t.Fatalf("TestEmptyAVL failed: %v", err)
	}
}

func TestDelete(t *testing.T) {
	c := &Container{}
	ref := &[]uint32{123}
	add(c, 123)
	err := containerVerify(c, ref)
	if err != nil {
		t.Fatalf("TestDelete containerVerify1 123 failed: %v", err)
	}
	if del(c, 124) {
		t.Fatalf("TestDelete del error1 124")
	}
	if !del(c, 123) {
		t.Fatalf("TestDelete del error2 123")
	}
	result := &[]uint32{}
	err = containerVerify(c, result)
	if err != nil {
		t.Fatalf("TestDelete containerVerify2 error %v", err)
	}
}

func findPosition(ref *[]uint32, val uint32) int {
	for i, v := range *ref {
		if v == val {
			return i
		}
	}
	return -1
}

func deletePosition(s *[]uint32, p int) {
	if len(*s)-1 == p {
		*s = (*s)[:p]
	} else {
		*s = append((*s)[:p], (*s)[p+1:]...)
	}

}

func TestSequentialInsertion(t *testing.T) {
	c := &Container{}
	ref := &[]uint32{}
	for i := uint32(0); i < 1000; i += 3 {
		add(c, i)
		*ref = append(*ref, i)
		err := containerVerify(c, ref)
		if err != nil {
			t.Fatalf("TestSequentialInsertion loop %d containerVerify error %v", i, err)
		}
	}
}

func TestRandomInsertion(t *testing.T) {
	c := &Container{}
	ref := &[]uint32{}

	for i := uint32(0); i < 100; i++ {
		val := uint32(rand.Int() % 1000)
		add(c, val)
		*ref = append(*ref, val)
		sort.Slice(*ref, func(i, j int) bool {
			return (*ref)[i] < (*ref)[j]
		})

		err := containerVerify(c, ref)
		if err != nil {
			t.Fatalf("TestRandomInsertion loop %d containerVerify error %v", i, err)
		}
	}
}

func TestRandomDeletion(t *testing.T) {
	c := &Container{}
	ref := &[]uint32{}
	for i := uint32(0); i < 100; i++ {
		val := uint32(rand.Int() % 1000)
		add(c, val)
		*ref = append(*ref, val)
	}
	sort.Slice(*ref, func(i, j int) bool {
		return (*ref)[i] < (*ref)[j]
	})

	for i := uint32(0); i < 200; i += 3 {
		val := uint32(rand.Int() % 1000)
		it := findPosition(ref, val)
		if it == -1 {
			if del(c, val) {
				t.Fatalf("TestRandomDeletion deleted nonexisting value: %d", val)
			}
		} else {
			if !del(c, val) {
				t.Fatalf("TestRandomDeletion delte failed: %d", val)
			}
			deletePosition(ref, it)
		}

		err := containerVerify(c, ref)
		if err != nil {
			t.Fatalf("TestRandomDeletion loop %d containerVerify error %v", i, err)
		}
	}
}

func TestInsertionDeletion(t *testing.T) {
	for i := uint32(0); i < 200; i++ {
		err := testInsert(i)
		if err != nil {
			t.Fatalf("TestInsertionDeletion testInsert failed: %v", err)
		}
		err = testInsertDup(i)
		if err != nil {
			t.Fatalf("TestInsertionDeletion testInsertDup failed: %v", err)
		}
		err = testRemove(i)
		if err != nil {
			t.Fatalf("TestInsertionDeletion testRemove failed: %v", err)
		}
	}
}

func TestAVLOffset(t *testing.T) {
	for i := uint32(1); i < 500; i++ {
		err := testAVLOffset(i)
		if err != nil {
			t.Fatalf("TestAVLOffset failed: %v", err)
		}
	}
}

func testAVLOffset(sz uint32) error {
	c := &Container{}
	for i := uint32(0); i < sz; i++ {
		add(c, i)
	}

	min := c.root
	for min != nil && min.Left != nil {
		min = min.Left
	}
	for i := uint32(0); i < sz; i++ {
		node := AVLOffset(min, int64(i))
		if (*Data)(unsafe.Pointer(node)).val != i {
			return fmt.Errorf("val not equals")
		}

		for j := uint32(0); j < sz; j++ {
			offset := int64(j) - int64(i)
			n2 := AVLOffset(node, offset)
			if (*Data)(unsafe.Pointer(n2)).val != j {
				return fmt.Errorf("val not equals")
			}
		}
		if AVLOffset(node, -(int64(i))-1) != nil {
			return fmt.Errorf("offset left %d error", -(int64(i))-1)
		}
		if AVLOffset(node, int64(sz-i)) != nil {
			return fmt.Errorf("offset right %d error", int64(sz-i))
		}
	}
	return nil
}
