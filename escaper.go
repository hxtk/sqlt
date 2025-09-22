package sqlt

import (
	"errors"
	"fmt"
	"slices"
	"text/template"
	"text/template/parse"
)

// ErrUnexpectedNode indicates that the escaper failed to escape the template.
//
// This error should never occur, and if it does, it indicates a bug in this
// package.
var ErrUnexpectedNode = errors.New("unexpected node while escaping template")

const escapeFuncName = "_sqlt_escapeSql"

func escapeNode(t *template.Template, node parse.Node, safe []string) error {
	switch n := node.(type) {
	case *parse.ActionNode:
		return escapeAction(n, safe)
	case *parse.BreakNode, *parse.CommentNode, *parse.ContinueNode, *parse.TextNode:
		return nil

	case *parse.IfNode:
		return escapeBranch(t, &n.BranchNode, safe)
	case *parse.RangeNode:
		return escapeBranch(t, &n.BranchNode, safe)
	case *parse.WithNode:
		return escapeBranch(t, &n.BranchNode, safe)

	case *parse.ListNode:
		return escapeList(t, n, safe)

	case *parse.TemplateNode:
		return escapeTemplate(t, n, safe)
	default:
		return fmt.Errorf("%w: %s", ErrUnexpectedNode, node.String())
	}
}

func escapeAction(n *parse.ActionNode, safe []string) error {
	if len(n.Pipe.Decl) != 0 {
		return nil
	}

	if len(n.Pipe.Cmds) < 1 {
		return nil
	}

	cmd := n.Pipe.Cmds[len(n.Pipe.Cmds)-1]
	if idNode, ok := cmd.Args[0].(*parse.IdentifierNode); ok {
		if slices.Contains(safe, idNode.Ident) {
			return nil
		}
	}

	n.Pipe.Cmds = append(n.Pipe.Cmds, newIdentCmd(escapeFuncName, n.Pipe.Position()))
	return nil
}

func newIdentCmd(identifier string, pos parse.Pos) *parse.CommandNode {
	return &parse.CommandNode{
		NodeType: parse.NodeCommand,
		Pos:      pos,
		Args: []parse.Node{
			parse.NewIdentifier(identifier).SetTree(nil).SetPos(pos),
		},
	}
}

func escapeBranch(t *template.Template, n *parse.BranchNode, safe []string) error {
	err := escapeList(t, n.List, safe)
	if err != nil {
		return err
	}
	return escapeList(t, n.ElseList, safe)
}

func escapeList(t *template.Template, n *parse.ListNode, safe []string) error {
	if n == nil {
		return nil
	}

	for _, v := range n.Nodes {
		err := escapeNode(t, v, safe)
		if err != nil {
			return err
		}
	}
	return nil
}

func escapeTemplate(t *template.Template, n *parse.TemplateNode, safe []string) error {
	tpl := t.Lookup(n.Name)
	if tpl == nil {
		return nil
	}

	return escapeNode(tpl, tpl.Root, safe)
}
