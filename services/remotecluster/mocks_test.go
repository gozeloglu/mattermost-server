// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package remotecluster

import (
	"context"
	"testing"

	"github.com/mattermost/mattermost-server/v6/einterfaces"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest/mock"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"github.com/mattermost/mattermost-server/v6/store"
	"github.com/mattermost/mattermost-server/v6/store/storetest/mocks"
)

type mockServer struct {
	remotes []*model.RemoteCluster
	logger  *mlog.Logger
	user    *model.User
}

func newMockServer(t *testing.T, remotes []*model.RemoteCluster) *mockServer {
	testLogger := mlog.CreateTestLogger(t, nil, mlog.StdAll...)

	return &mockServer{
		remotes: remotes,
		logger:  testLogger,
	}
}

func (ms *mockServer) SetUser(user *model.User) {
	ms.user = user
}

func (ms *mockServer) Config() *model.Config                                  { return nil }
func (ms *mockServer) GetMetrics() einterfaces.MetricsInterface               { return nil }
func (ms *mockServer) IsLeader() bool                                         { return true }
func (ms *mockServer) AddClusterLeaderChangedListener(listener func()) string { return model.NewId() }
func (ms *mockServer) RemoveClusterLeaderChangedListener(id string)           {}
func (ms *mockServer) GetLogger() mlog.LoggerIFace {
	return ms.logger
}
func (ms *mockServer) GetStore() store.Store {
	anyQueryFilter := mock.MatchedBy(func(filter model.RemoteClusterQueryFilter) bool {
		return true
	})
	anyUserId := mock.AnythingOfType("string")

	remoteClusterStoreMock := &mocks.RemoteClusterStore{}
	remoteClusterStoreMock.On("GetByTopic", "share").Return(ms.remotes, nil)
	remoteClusterStoreMock.On("GetAll", anyQueryFilter).Return(ms.remotes, nil)

	userStoreMock := &mocks.UserStore{}
	userStoreMock.On("Get", context.Background(), anyUserId).Return(ms.user, nil)

	storeMock := &mocks.Store{}
	storeMock.On("RemoteCluster").Return(remoteClusterStoreMock)
	storeMock.On("User").Return(userStoreMock)
	return storeMock
}
func (ms *mockServer) Shutdown() { ms.logger.Shutdown() }
