package main

import (
	"errors"
	"fmt"
	// "io"
	"os"
	"runtime"
)

// 这是链表调用节点所需的接口
type Node interface {
	// 当前节点设置上一个节点
	SetPrev(Node)
	// 获取当前节点的上一个节点
	GetPrev() Node
	// 当前节点设置下一个节点
	SetNext(Node)
	// 获取当前节点的下一个节点
	GetNext() Node
	// 当前节点进行非阻塞的运行
	Run()
	// 是否满足callback所需要的条件,
	// 如果是,就运行回调函数
	IsContinue() bool
	// 当前节点运行回调函数
	RunCallback()
}

// 理论上说,只要符合 接口Node 都能运行
// 这次进行的是io操作
// 将不同大小 需要的文件的文件路径 赋值给 当前节点的filePath
// 在运行Run函数的时候,根据文件路径打开文件,将文件资源赋值给 当前节点的file,
// 然后每次判断 当前节点的file 是否为空
// 如果不为空,每次读取5个字节,然后赋值给 节点的content
// 当Run函数出现 一些错误 或者 是文件读取完毕时的 io.EOF 时
// 停止Run函数运行,关闭文件资源,并且将 节点的content 和 错误 给回调函数处理
type node struct {
	file     *os.File
	filePath string
	content  []byte
	off      int64
	callback func([]byte, error)
	prev     Node
	next     Node
	err      error
}

// 节点初始化
func nodeNew() *node {
	return &node{
		file:     nil,
		filePath: "",
		content:  make([]byte, 5),
		off:      0,
		callback: nil,
		prev:     nil,
		next:     nil,
		err:      nil,
	}
}

// 这是一条循环链表
// 循环链表的实现函数

// Push 函数 参数: 实现了 Node接口 的类型 返回值:无
// 说明:将节点放进链表中,如果链表的长度为0时,则 链表的 head ,tail ,current 都会指向该节点
// 如果长度 > 0时,
// 第一步将 节点的prev 指向 链表的 tail,将 链表的tail的next 指向 节点
// 第二步将 节点赋值给 链表的tail
// 第三步将 链表的tail的next 指向 链表的head , 将 链表的head的prev 指向 链表的tail

// IsEnd 函数 参数:无 返回值:布尔
// 说明:当链表的长度为0时 或者 链表的current 为空 时,为结束链表运行

// Pop 函数 参数:无 返回值: 实现了 Node接口 的类型,error
// 说明:将 链表的当前节点 取出,并删除 在链表上的该节点
// 当 链表的当前节点 为 链表的head 时,链表的head的上一个节点不能为空
// 或者 链表的当前节点 为 链表的tail 时,链表的tail的下一个节点不能为空
// 否则强制将所有节点删除,返回当前的节点

// Next 函数 参数:无 返回值:无
// 说明: 链表的当前节点 指向 它的下一个节点

type Queue struct {
	// 指向第一个节点
	head Node
	// 指向最后一个节点
	tail Node
	// 指向当前节点
	current Node
	// 链表的长度
	length int
}

func QueueNew() *Queue {
	return &Queue{
		head:    nil,
		tail:    nil,
		current: nil,
		length:  0,
	}
}

func (this *node) SetPrev(n Node) {
	this.prev = n
}

func (this *node) GetPrev() Node {
	return this.prev
}

func (this *node) SetNext(n Node) {
	this.next = n
}

func (this *node) GetNext() Node {
	return this.next
}

func (this *node) SetCallback(function func([]byte, error)) {
	this.callback = function
}

func (this *node) RunCallback() {
	this.file.Close()
	this.callback(this.content, this.err)
}

func (this *node) SetFileName(filePath string) {
	this.filePath = filePath
}

func (this *node) Run() {
	if this.file == nil {
		this.file, this.err = os.Open(this.filePath)
	} else {
		NewContent := make([]byte, 5)
		OldContent := this.content
		OldLength := cap(OldContent)
		_, this.err = this.file.ReadAt(NewContent, this.off)
		this.content = make([]byte, OldLength+5)
		for i := 0; i < OldLength; i++ {
			this.content[i] = OldContent[i]
		}
		for i := 0; i < 5; i++ {
			this.content[OldLength+i] = NewContent[i]
		}
		this.off += 5
	}
}

func (this *node) IsContinue() bool {
	return this.err == nil
}

func (this *Queue) Push(n Node) {
	if this.length <= 0 {
		this.head = n
		this.current = this.head
		this.tail = this.head
	} else {
		n.SetPrev(this.tail)
		this.tail.SetNext(n)
		this.tail = n
		this.head.SetPrev(this.tail)
		this.tail.SetNext(this.head)
	}
	this.length++
}

func (this *Queue) IsEnd() bool {
	return this.current == nil || this.length == 0
}

func (this *Queue) Next() {
	if this.current != nil {
		if this.current.GetNext() != nil {
			this.current = this.current.GetNext()
		}
	}
}

func (this *Queue) Pop() (Node, error) {
	if this.length <= 0 {
		err := errors.New("Queue length is error")
		return nil, err
	} else {
		var current Node = nil
		if (this.current == this.head && this.head.GetPrev() == nil) || (this.current == this.tail && this.tail.GetNext() == nil) {
			current = this.current
			this.Clean()
		} else {
			prev, next := this.current.GetPrev(), this.current.GetNext()
			prev.SetNext(next)
			next.SetPrev(prev)
			this.current.SetNext(nil)
			this.current.SetPrev(nil)
			current = this.current
			this.current = prev
			this.length--
			runtime.GC()
		}

		return current, nil
	}
}

func (this *Queue) Clean() {
	this.head = nil
	this.current = nil
	this.tail = nil
	this.length = 0
	runtime.GC()
}

func main() {
	q := QueueNew()

	n1 := nodeNew()
	n1.filePath = "test.html"
	n1.SetCallback(func(content []byte, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(string(content))
	})

	n2 := nodeNew()
	n2.filePath = "index/index.html"
	n2.SetCallback(func(content []byte, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(string(content))
	})

	n3 := nodeNew()
	n3.filePath = "test.html"
	n3.SetCallback(func(content []byte, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(string(content))
	})

	n4 := nodeNew()
	n4.filePath = "index/index.html"
	n4.SetCallback(func(content []byte, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(string(content))
	})

	q.Push(n1)
	q.Push(n2)
	q.Push(n3)
	q.Push(n4)

	// fmt.Println(q.Pop())
	// fmt.Println(q.Pop())
	// fmt.Println(q.Pop())
	for q.IsEnd() == false {
		q.current.Run()
		if q.current.IsContinue() == false {
			data, err := q.Pop()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(0)
			}
			data.RunCallback()
		}
		q.Next()
	}
}
