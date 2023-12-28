package main

import "github.com/yum45f/did-dnssec/cmd"

func main() {
	cmd.Execute()
}

// func main() {
// 	f, err := os.Open("example/did.json")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()

// 	bytes, err := io.ReadAll(f)
// 	if err != nil {
// 		panic(err)
// 	}

// 	doc, err := core.NewDocumentTreeFromJSON(bytes)
// 	if err != nil {
// 		panic(err)
// 	}

// 	recursivePrintTree(doc, 0)

// 	fmt.Println("\n-----")

// 	for _, rr := range doc.RRs("yuma.space.") {
// 		fmt.Println(rr)
// 	}

// 	fmt.Println("\n-----")

// 	f, err = os.Create("out.txt")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()

// 	if err := doc.DumpRRs(f, "yuma.space."); err != nil {
// 		panic(err)
// 	} else {
// 		fmt.Println("Dumped to out.txt")
// 	}
// }

// func recursivePrintTree(tree *core.Node, depth int) {
// 	indent := getIndent(depth)
// 	if tree.Value.Type != core.ValTypeMap && tree.Value.Type != core.ValTypeArray {
// 		fmt.Printf("%s%s: %s (%s)\n", indent, tree.Key, tree.Value, tree.Value.Type.String())
// 	} else {
// 		fmt.Printf("%s%s: (%s)\n", indent, tree.Key, tree.Value.Type.String())
// 		for _, child := range *tree.Children {
// 			recursivePrintTree(&child, depth+1)
// 		}
// 	}
// }

// func getIndent(depth int) string {
// 	indent := ""
// 	for i := 0; i < depth; i++ {
// 		indent += "  "
// 	}

// 	return indent
// }
