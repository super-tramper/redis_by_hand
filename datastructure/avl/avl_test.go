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
			from = &cur.left
		} else {
			from = &cur.right
		}
		if *from == nil {
			*from = &data.node
			data.node.parent = cur
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
			cur = cur.left
		} else {
			cur = cur.right
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
	if node.parent != parent {
		return fmt.Errorf("parent error")
	}
	err := avlVerify(node, node.left)
	if err != nil {
		return err
	}
	err = avlVerify(node, node.right)
	if err != nil {
		return err
	}

	if node.cnt != 1+Cnt(node.left)+Cnt(node.right) {
		return fmt.Errorf("cnt error")
	}

	l := Depth(node.left)
	r := Depth(node.right)

	if !(l == r || l+1 == r || l == r+1) {
		return fmt.Errorf("balance error")
	}
	if node.depth != 1+tools.Max(l, r) {
		return fmt.Errorf("depth error, expected %d, actual %d", node.depth, 1+tools.Max(l, r))
	}

	val := (*Data)(unsafe.Pointer(node)).val
	if node.left != nil {
		if node.left.parent != node {
			return fmt.Errorf("left children parent error")
		}
		if (*Data)(unsafe.Pointer(node.left)).val > val {
			return fmt.Errorf("left children value greater than parent value")
		}
	}
	if node.right != nil {
		if node.right.parent != node {
			return fmt.Errorf("right children parent error.")
		}
		if !((*Data)(unsafe.Pointer(node.right)).val >= val) {
			return fmt.Errorf(
				"right children value error, expected: %d, got: %d",
				val,
				(*Data)(unsafe.Pointer(node.right)).val)
		}
	}
	return nil
}

func extract(node *Node, extracted *[]uint32) {
	if node == nil {
		return
	}
	extract(node.left, extracted)
	*extracted = append(*extracted, (*Data)(unsafe.Pointer(node)).val)
	extract(node.right, extracted)
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
			fmt.Println(extracted)
			fmt.Println(*ref)
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
		err := containerVerify(&c, &ref)
		if err != nil {
			return err
		}

		add(&c, val)
		ref = append(ref, val)
		err = containerVerify(&c, &ref)
		if err != nil {
			return err
		}
	}
	return nil
}

func testInsertDup(sz uint32) {
	for val := uint32(0); val < sz; val++ {
		c := Container{}
		ref := make([]uint32, 0)
		for i := uint32(0); i < sz; i++ {
			add(&c, i)
			ref = append(ref, i)
		}
		err := containerVerify(&c, &ref)
		if err != nil {
			return
		}

		add(&c, val)
		ref = append(ref, val)
		err = containerVerify(&c, &ref)
		if err != nil {
			return
		}
	}
}

func testRemove(sz uint32) {
	for val := uint32(0); val < sz; val++ {
		c := Container{}
		ref := make([]uint32, 0)
		for i := uint32(0); i < sz; i++ {
			add(&c, i)
			ref = append(ref, i)
		}
		err := containerVerify(&c, &ref)
		if err != nil {
			return
		}

		add(&c, val)
		ref = append(ref, val)
		err = containerVerify(&c, &ref)
		if err != nil {
			return
		}

		if !del(&c, val) {
			panic("testRemove del error")
		}
		for i, v := range ref {
			if v == val {
				if i < len(ref)-1 {
					ref = append(ref[:i], ref[i+1:]...)
				} else {
					ref = ref[:i]
				}
			}
			err := containerVerify(&c, &ref)
			if err != nil {
				return
			}
		}
	}
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

//func TestRandomDeletion(t *testing.T) {
//	c := &Container{}
//	ref := &[]uint32{}
//	for i := uint32(0); i < 1000; i += 3 {
//		add(c, i)
//		*ref = append(*ref, i)
//	}
//
//	for i := uint32(0); i < 200; i += 3 {
//		val := uint32(rand.Int() % 1000)
//		fmt.Println("delete ", val)
//		it := findPosition(ref, val)
//		if it == -1 {
//			if del(c, val) {
//				t.Fatalf("TestRandomDeletion deleted nonexisting value: %d", val)
//			}
//		} else {
//			if !del(c, val) {
//				t.Fatalf("TestRandomDeletion delte failed: %d", val)
//			}
//			deletePosition(ref, it)
//		}
//
//		err := containerVerify(c, ref)
//		if err != nil {
//			t.Fatalf("TestRandomDeletion loop %d containerVerify error %v", i, err)
//		}
//	}
//}
