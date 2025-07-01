package services

import (
	"github.com/greboid/tithon/irc"
	"slices"
	"sync"
)

type ServerList struct {
	Parents     []*ServerListItem
	OrderedList []*ServerListItem
}

type ServerListItem struct {
	Window   *irc.Window
	Link     string
	Name     string
	Children []*ServerListItem
}

type WindowService struct {
	activeWindow  *irc.Window
	activeLock    sync.RWMutex
	pendingUpdate UpdateTrigger
	windowChanged WindowChangedTrigger
}

type ServerListService struct {
	listlock sync.RWMutex
}

func NewServerListService() *ServerListService {
	return &ServerListService{}
}

func (sls *ServerListService) GetServerList(connectionManager *irc.ServerManager) *ServerList {
	sls.listlock.RLock()
	defer sls.listlock.RUnlock()

	serverList := &ServerList{}
	connections := connectionManager.GetConnections()

	for i := range connections {
		serverIndex := slices.IndexFunc(serverList.Parents, func(item *ServerListItem) bool {
			return item.Window == connections[i].GetWindow()
		})

		var server *ServerListItem
		if serverIndex == -1 {
			server = &ServerListItem{
				Window:   connections[i].GetWindow(),
				Link:     connections[i].GetID(),
				Name:     connections[i].GetName(),
				Children: nil,
			}
			serverList.Parents = append(serverList.Parents, server)
			serverList.OrderedList = append(serverList.OrderedList, server)
		} else {
			server = serverList.Parents[serverIndex]
		}

		channels := connections[i].GetChannels()
		for j := range channels {
			windowIndex := slices.IndexFunc(server.Children, func(item *ServerListItem) bool {
				return item.Window == channels[j].Window
			})
			if windowIndex == -1 {
				child := &ServerListItem{
					Window:   channels[j].Window,
					Link:     connections[i].GetID() + "/" + channels[j].GetID(),
					Name:     channels[j].GetName(),
					Children: nil,
				}
				server.Children = append(server.Children, child)
				serverList.OrderedList = append(serverList.OrderedList, child)
			}
		}

		queries := connections[i].GetQueries()
		for j := range queries {
			windowIndex := slices.IndexFunc(server.Children, func(item *ServerListItem) bool {
				return item.Window == queries[j].Window
			})
			if windowIndex == -1 {
				child := &ServerListItem{
					Window:   queries[j].Window,
					Link:     connections[i].GetID() + "/" + queries[j].GetID(),
					Name:     queries[j].GetName(),
					Children: nil,
				}
				server.Children = append(server.Children, child)
				serverList.OrderedList = append(serverList.OrderedList, child)
			}
		}
	}

	return serverList
}
