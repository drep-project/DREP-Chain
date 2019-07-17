package main

import "fmt"

type ICompair interface {
	 Compair(ICompair) bool
}

type Node struct {
	Val ICompair
	Left *Node
	Right *Node
	Parent *Node
}

func NewNode(val ICompair) *Node {
	return &Node{
		Val:val,
	}
}

func (node *Node)HasLeft() bool {
	return node.Left != nil
}

func (node *Node)HasRight() bool {
	return node.Right != nil
}

func (node *Node) IsLeft() bool {
	return node.Parent.Left == node
}

func (node *Node) IsRight() bool {
	return node.Parent.Right == node
}

func (node *Node) AddNode(newNode *Node) {
	if newNode.Val.Compair(node.Val){
		if node.HasRight() {
			node.Right.AddNode(newNode)
		}else{
			node.Right = newNode
			newNode.Parent = node
		}
	}else {
		if node.HasLeft() {
			node.Left.AddNode(newNode)
		}else{
			node.Left = newNode
			newNode.Parent = node
		}
	}
}

func (node *Node) DelNode(delNode *Node) {
	if delNode == node {
		if node.HasLeft(){
			if node.HasRight() {
				right := node.Right
				if node.IsLeft() {
					node.Parent.Left = right
				}else{
					node.Parent.Right = right
				}
				right.AddNode(node.Left)
				return
			}else{
				if node.IsLeft() {
					node.Parent.Left = node.Left
				}else{
					node.Parent.Right = node.Left
				}
				return
			}
		}else{
			if node.IsLeft() {
				node.Parent.Left = node.Right
			}else{
				node.Parent.Right = node.Right
			}
			return
		}
	}
	if delNode.Val.Compair(node.Val) {
      node.Right.DelNode(delNode)
	}else{
		node.Left.DelNode(delNode)
	}
}

func printAndGetChild(nodes []*Node){

	netChildren := []*Node{}
	for i:= 0;i<len(nodes);i++ {
		val := nodes[i].Val.(*IntVal)
		fmt.Print(*val)
		fmt.Print("|")
		if nodes[i].HasLeft() {
			netChildren = append(netChildren, nodes[i].Left)
		}
		if nodes[i].HasRight() {
			netChildren = append(netChildren, nodes[i].Right)
		}
	}
	fmt.Println("")
	if len(nodes) >0  {
		printAndGetChild(netChildren)
	}
}
func (node *Node) Print() {
	printAndGetChild([]*Node{node})
}

func ShellSort(data []int) {
	if len(data) < 2 {
		return
	}
	key := len(data)/2

	for key > 0 {
		for i := key;i < len(data);i++ {
			j := i
			for j>=key&&data[j]<data[j-key] {
				data[j],data[j-key] = data[j-key],data[j]
				j = j-key
			}
		}
		key = key/2
	}
}

type IntVal int
func NewIntVal(val int) *IntVal {
	return (*IntVal)(&val)
}
func (intVal *IntVal) Compair(anotherVal ICompair) bool {
	return *intVal > *anotherVal.(*IntVal)
}

func main() {
  rootNode :=NewNode(NewIntVal(10))
  eight := NewNode(NewIntVal(8))
  rootNode.AddNode(eight)
  rootNode.AddNode(NewNode(NewIntVal(9)))
  rootNode.AddNode(NewNode(NewIntVal(7)))
  rootNode.AddNode(NewNode(NewIntVal(11)))
  rootNode.Print()
	rootNode.DelNode(eight)
	rootNode.Print()
}