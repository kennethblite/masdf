package main

import (
	"container/heap"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
)

const nodesize = 4

var keymap = map[byte]int{
	32: 1, 115: 2, 100: 3, 102: 4}
var invkeymap = map[int]string{
	1: "s", 2: "d", 3: "f", 0: "<space>"}

type Node struct {
	Sum  float64
	Char string
	Next NodeTree
}

type NodeTree []*Node

func (a NodeTree) Len() int      { return len(a) }
func (a NodeTree) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a *NodeTree) Push(x interface{}) {
	node := x.(*Node)
	*a = append(*a, node)
}
func (a *NodeTree) Pop() interface{} {
	old := *a
	n := len(old)
	node := old[n-1]
	*a = old[0 : n-1]
	return node
}

func (a NodeTree) Less(i, j int) bool { return a[i].Sum > a[j].Sum }

func FrequencyFile(filename string) map[string]float64 {
	var objmap map[string]float64
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(b, &objmap)
	return objmap
}

var printmap bool
var shell bool
func init() {
	flag.BoolVar(&printmap, "p", false, "prints the key mapping tree")
	flag.BoolVar(&shell, "s", false, "begins a shell session with keyboard")
	flag.Parse()
}

func main() {
	objmap := FrequencyFile("res/non-numeral.json")
	//tree := maptoNodeArray(objmap)
	priority_branch := make(NodeTree, len(objmap))
	i := 0
	summed := 0.0
	for char, value := range objmap {
		priority_branch[i] = &Node{
			value,
			char,
			nil,
		}
		i++
		summed += value
	}
	for i := range priority_branch {
		priority_branch[i].Sum = priority_branch[i].Sum / summed
		//fmt.Println(priority_branch[i])
	}
	heap.Init(&priority_branch)
	sort.Sort(&priority_branch)
	for len(priority_branch) > nodesize {
		newnode := make([]*Node, 0, nodesize)
		size := 0.0
		for i := 0; i < nodesize; i++ {
			pop_node := priority_branch.Pop().(*Node)
			newnode = append(newnode, pop_node)
			size += pop_node.Sum
		}
		insert_node := &Node{size, "", newnode}
		heap.Push(&priority_branch, insert_node)
		sort.Sort(&priority_branch)
	}
	if printmap {
		printmapping(priority_branch, func(n Node) bool {
			return n.Char == ""
		}, "")
	}else if shell{
		reader := bufio.NewReader(os.Stdin)
		subprocess := exec.Command("bash")
		stdin, err := subprocess.StdinPipe()
		if err != nil{
			fmt.Println(err)
		}
		defer stdin.Close()
		subprocess.Stdout = os.Stdout
		subprocess.Stderr = os.Stderr
		if err = subprocess.Start(); err != nil{
			fmt.Println(err)
		}
		for {
		text, _ := reader.ReadBytes('\n')
		io.WriteString(stdin, string(text))
		}
		fmt.Println("END")
	}else{
		getChar(priority_branch)
	}

}

func printmapping(n NodeTree, f func(Node) bool, prefix string) {
	for i, v := range n {
		if f(*v) {
			printmapping(v.Next, f, prefix+invkeymap[i])
		} else {
			fmt.Println(v.Char + " :" + prefix + invkeymap[i])
		}
	}
}

func getChar(n NodeTree) {
	cur_node := n
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()
	var lastbyte byte = 0
	var b []byte = make([]byte, 1)
	for {
		os.Stdin.Read(b)
		//fmt.Println("I got the byte",b,"("+string(b)+")")
		if keymap[b[0]] != 0 {
			input := keymap[b[0]] - 1
			nodeindex := cur_node[input]
			//fmt.Println(nodeindex)
			if nodeindex.Char != "" {
				fmt.Print(nodeindex.Char)
				cur_node = n
			} else {
				cur_node = cur_node[input].Next
			}
		}
		if b[0] == 97 {
			if lastbyte == 97 {
				break
			} else {
				cur_node = n
			}
		}
		lastbyte = b[0]
	}

}
