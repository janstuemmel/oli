package main

import (
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	// "github.com/charmbracelet/glamour"
)

func main() {

	// o1, _ := glamour.Render("**hello", "dark")
	// o2, _ := glamour.Render("**", "dark")
	// fmt.Print(o1)
	// fmt.Print(o2)
	// return

	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	source := []byte(`# Title

Some **bold** text and a list:

- Item 1
- Item 2


	foo = 1
`)

	source = []byte("Here is a code block:\n\n```go\nfmt.Println(\"hello\")\n```")

	// Parse Markdown into an AST
	doc := md.Parser().Parse(text.NewReader(source))

	// curNode := doc.FirstChild().NextSibling()

	// fmt.Println(curNode)
	// fmt.Println(string(curNode.Text(source)))

	// Walk the AST
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {

		if entering {
			// fmt.Println(n.Kind(), n.ChildCount())

			if n.Type() == 1 {

				// fmt.Println(n.Lines())
				fmt.Println()
				fmt.Println(string(n.Text(source)))
				fmt.Println()
			}

			fmt.Println()
		}

		// if entering {
		// 	switch node := n.(type) {
		// 	case *ast.Heading:
		// 		fmt.Printf("Heading level %d\n", node.Lines().Value)
		// 	case *ast.Text:
		// 		fmt.Printf("Text: %q\n", node.Text(source))
		// 	case *ast.List:
		// 		fmt.Println("List")
		// 	case *ast.ListItem:
		// 		fmt.Println("List item")
		// 	case *ast.FencedCodeBlock:
		// 		fmt.Printf("code item %d", node.Language())
		// 	}
		// }

		return ast.WalkContinue, nil
	})

}
