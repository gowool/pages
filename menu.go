package pages

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type Menu interface {
	Get(ctx context.Context, handle string) (model.Menu, error)
}

type DefaultMenu struct {
	menuRepo repository.Menu
	nodeRepo repository.Node
}

func NewDefaultMenu(menuRepo repository.Menu, nodeRepo repository.Node) *DefaultMenu {
	return &DefaultMenu{
		menuRepo: menuRepo,
		nodeRepo: nodeRepo,
	}
}

func (m *DefaultMenu) Get(ctx context.Context, handle string) (model.Menu, error) {
	menu, err := m.menuRepo.FindByHandle(ctx, handle)
	if err != nil {
		return model.Menu{}, err
	}

	if menu.NodeID != nil {
		data, err := m.nodeRepo.FindWithChildren(ctx, *menu.NodeID)
		if err != nil {
			return model.Menu{}, err
		}
		menu.Node = BuildTree(data, *menu.NodeID)
	}

	if menu.NodeID == nil {
		// not found root node
		return model.Menu{}, errors.Join(sql.ErrNoRows, ErrMenuNotFound)
	}
	return menu, nil
}

func BuildTree(nodes []model.Node, id int64) *model.Node {
	nodeMap := make(map[int64]*model.Node)
	var rootNode *model.Node

	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	for i := range nodes {
		node := &nodes[i]
		if node.ID == id {
			rootNode = node
		} else {
			if parent, ok := nodeMap[node.ParentID]; ok {
				parent.Children = append(parent.Children, node)
				node.Parent = parent
			}
		}
	}

	if rootNode == nil {
		return nil
	}

	sortChildrenByPosition(rootNode.Children)

	return rootNode
}

func sortChildrenByPosition(nodes []*model.Node) {
	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].Position < nodes[j].Position
	})

	for i := range nodes {
		if len(nodes[i].Children) > 0 {
			sortChildrenByPosition(nodes[i].Children)
		}
	}
}
