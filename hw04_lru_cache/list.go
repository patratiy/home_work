package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	length int
	front  *ListItem
	back   *ListItem
}

func NewList() List {
	return new(list)
}

func (lst *list) Len() int {
	return lst.length
}

func (lst *list) Front() *ListItem {
	return lst.front
}

func (lst *list) Back() *ListItem {
	return lst.back
}

func (lst *list) PushFront(v interface{}) *ListItem {
	item := &ListItem{Value: v, Next: lst.front}
	if lst.front != nil {
		lst.front.Prev = item
	} else {
		lst.back = item
	}
	lst.front = item
	lst.length++
	return item
}

func (lst *list) PushBack(val interface{}) *ListItem {
	item := &ListItem{Value: val, Prev: lst.back}
	if lst.back != nil {
		lst.back.Next = item
	} else {
		lst.front = item
	}
	lst.back = item
	lst.length++
	return item
}

func (lst *list) Remove(itm *ListItem) {
	if itm.Prev != nil {
		itm.Prev.Next = itm.Next
	} else {
		lst.front = itm.Next
	}
	if itm.Next != nil {
		itm.Next.Prev = itm.Prev
	} else {
		lst.back = itm.Prev
	}
	itm.Next = nil
	itm.Prev = nil
	lst.length--
}

func (lst *list) MoveToFront(itm *ListItem) {
	if itm == lst.front {
		return
	}
	if itm.Prev != nil {
		itm.Prev.Next = itm.Next
	} else {
		lst.front = itm.Next
	}
	if itm.Next != nil {
		itm.Next.Prev = itm.Prev
	} else {
		lst.back = itm.Prev
	}
	itm.Prev = nil
	itm.Next = lst.front
	lst.front.Prev = itm
	lst.front = itm
}
