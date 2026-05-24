package core

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type tableCell struct {
	Text             string
	RowspanCarryover bool
}

func parseHTMLTable(tableHTML string) ([][]string, error) {
	cellRows, err := parseHTMLTableCells(tableHTML)
	if err != nil {
		return nil, err
	}
	rows := make([][]string, len(cellRows))
	for i, cellRow := range cellRows {
		rows[i] = tableCellTexts(cellRow)
	}
	return rows, nil
}

func parseHTMLTableCells(tableHTML string) ([][]tableCell, error) {
	root, err := html.Parse(strings.NewReader(tableHTML))
	if err != nil {
		return nil, err
	}
	trs := findNodes(root, "tr")
	rows := make([][]tableCell, 0, len(trs))
	spans := map[int]map[int]string{}

	for rowIndex, tr := range trs {
		row := []tableCell{}
		occupied := map[int]bool{}
		if pending := spans[rowIndex]; len(pending) > 0 {
			for col, value := range pending {
				for len(row) <= col {
					row = append(row, tableCell{})
				}
				row[col] = tableCell{Text: value, RowspanCarryover: true}
				occupied[col] = true
			}
		}

		col := 0
		for _, cell := range tableCells(tr) {
			for occupied[col] {
				col++
			}
			text := normalizeText(nodeText(cell))
			colspan := positiveAttrInt(cell, "colspan", 1)
			rowspan := positiveAttrInt(cell, "rowspan", 1)
			for offset := 0; offset < colspan; offset++ {
				targetCol := col + offset
				for len(row) <= targetCol {
					row = append(row, tableCell{})
				}
				row[targetCol] = tableCell{Text: text}
				occupied[targetCol] = true
				for r := 1; r < rowspan; r++ {
					targetRow := rowIndex + r
					if spans[targetRow] == nil {
						spans[targetRow] = map[int]string{}
					}
					spans[targetRow][targetCol] = text
				}
			}
			col += colspan
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no table rows found")
	}
	return rows, nil
}

func tableCellTexts(row []tableCell) []string {
	values := make([]string, len(row))
	for i, cell := range row {
		values[i] = cell.Text
	}
	return values
}

func findNodes(root *html.Node, tag string) []*html.Node {
	var nodes []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == tag {
			nodes = append(nodes, n)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return nodes
}

func tableCells(row *html.Node) []*html.Node {
	var cells []*html.Node
	for child := row.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && (child.Data == "td" || child.Data == "th") {
			cells = append(cells, child)
		}
	}
	return cells
}

func positiveAttrInt(n *html.Node, key string, fallback int) int {
	for _, attr := range n.Attr {
		if attr.Key != key {
			continue
		}
		value, err := strconv.Atoi(strings.TrimSpace(attr.Val))
		if err != nil || value <= 0 {
			return fallback
		}
		return value
	}
	return fallback
}

func nodeText(n *html.Node) string {
	var parts []string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			parts = append(parts, node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)
	return strings.Join(parts, " ")
}
