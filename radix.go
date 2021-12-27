package main

import (
	"fmt"
	"html/template"
	"path"
	"sort"
	"strconv"
	"strings"
)

// A very specialized radix, nodes are doubly linked, bondaries are file paths.
// The main goal of this structure is to simulate the file structure and provide hierarchical coverage information.
type radixNode struct {
	Parent *radixNode
	Sub    map[string]*radixNode

	Pkg     bool
	File    bool
	Body    template.HTML
	SetMode bool

	Total   int64
	Covered int64
}

func (r *radixNode) Simplify() {
	r.simplify()
}

func (r *radixNode) simplify() {
	for k, v := range r.Sub {
		v.simplify()

		if len((*v).Sub) == 1 {
			for kk, vv := range v.Sub {
				delete(r.Sub, k)
				r.Sub[path.Join(k, kk)] = vv
				vv.File = vv.File || v.File
				vv.Pkg = vv.Pkg || v.Pkg
				vv.Parent = r
			}
		}
	}
}

func (r *radixNode) Make(path string) *radixNode {
	tree := r
	for _, part := range strings.Split(path, "/") {
		if tree.Sub == nil {
			tree.Sub = make(map[string]*radixNode)
		}

		subTree := tree.Sub[part]
		if subTree == nil {
			subTree = &radixNode{
				Parent: tree,
			}
			tree.Sub[part] = subTree
		}
		tree = subTree
	}

	return tree
}

func (r radixNode) Keys() []string {
	keys := make([]string, 0, len(r.Sub))
	for k := range r.Sub {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (r radixNode) Packages() []string {
	return r.packages("")
}

func (r radixNode) packages(prefix string) []string {
	var pkg []string
	for k, v := range r.Sub {
		if v.Pkg {
			pkg = append(pkg, path.Join(prefix, k))
		}

		pkg = append(pkg, v.packages(path.Join(prefix, k))...)
	}

	return pkg
}

func (r *radixNode) Files(pkg string) []string {
	tree := r
	modPkg := pkg
	for {
		stop := true
		match := false
		for k, v := range tree.Sub {
			if k == modPkg {
				tree = v
				match = true
				break
			}

			if strings.HasPrefix(modPkg, k) {
				modPkg = strings.TrimPrefix(strings.TrimPrefix(modPkg, k), "/")
				stop = false
				match = true
				tree = v
				break
			}
		}
		if !match {
			return nil
		}

		if stop {
			break
		}
	}

	files := make([]string, 0)
	for k, v := range tree.Sub {
		if v.File {
			files = append(files, path.Join(pkg, k))
		}
	}
	sort.Strings(files)

	return files
}

func (r *radixNode) Path() string {
	path := make([]string, 0)
	var prev *radixNode
	cur := r
	for cur.Parent == nil {
		prev = cur
		cur = cur.Parent
		for k, v := range cur.Sub {
			if v == prev {
				path = append(path, k)
				break
			}
		}
	}

	// Reverse slice
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return strings.Join(path, "/")
}

func (r radixNode) String() string {
	var sb strings.Builder
	r.string(&sb, 0, make(map[int]bool))
	return sb.String()
}

func (r radixNode) string(sb *strings.Builder, level int, lvlDone map[int]bool) {
	n := 0
	for _, k := range r.Keys() {
		v := r.Sub[k]

		childLast := n+1 == len(r.Sub)
		lvlDone[level] = childLast

		for i := 1; i < level; i++ {
			if lvlDone[i] {
				sb.WriteString("  ")
			} else {
				sb.WriteString("│ ")
			}
		}

		if level != 0 {
			if childLast {
				sb.WriteRune('└')
			} else {
				sb.WriteRune('├')
			}
			sb.WriteString("─")
		}

		sb.WriteString(k)
		if v.Pkg {
			sb.WriteString(" (P)")
		}
		if v.File {
			sb.WriteString(" (F)")
		}
		sb.WriteString(" (")
		sb.WriteString(strconv.FormatFloat(v.CoveragePercent(), 'f', 2, 64))
		sb.WriteString("%)")
		sb.WriteRune('\n')

		v.string(sb, level+1, lvlDone)

		n++
	}
}

func (r radixNode) CoveragePercent() float64 {
	return percent(r.Covered, r.Total)
}

func (r radixNode) CoverageStr() string {
	return strconv.FormatFloat(r.CoveragePercent(), 'f', 2, 64)
}

func (r radixNode) CovClass() string {
	return fmt.Sprintf("cov%d", int(r.CoveragePercent()/10))
}
