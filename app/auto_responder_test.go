// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"
)

func TestSetAutoResponderStatus(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(th.Context, user)

	th.App.SetStatusOnline(user.Id, true)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_responder_active"] = "true"
	patch.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	userUpdated1, _ := th.App.PatchUser(user.Id, patch, true)

	// autoResponder is enabled, status should be OOO
	th.App.SetAutoResponderStatus(userUpdated1, user.NotifyProps)

	status, err := th.App.GetStatus(userUpdated1.Id)
	require.Nil(t, err)
	assert.Equal(t, model.StatusOutOfOffice, status.Status)

	patch2 := &model.UserPatch{}
	patch2.NotifyProps = make(map[string]string)
	patch2.NotifyProps["auto_responder_active"] = "false"
	patch2.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	userUpdated2, _ := th.App.PatchUser(user.Id, patch2, true)

	// autoResponder is disabled, status should be ONLINE
	th.App.SetAutoResponderStatus(userUpdated2, userUpdated1.NotifyProps)

	status, err = th.App.GetStatus(userUpdated2.Id)
	require.Nil(t, err)
	assert.Equal(t, model.StatusOnline, status.Status)

}

func TestDisableAutoResponder(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(th.Context, user)

	th.App.SetStatusOnline(user.Id, true)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_responder_active"] = "true"
	patch.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	th.App.PatchUser(user.Id, patch, true)

	th.App.DisableAutoResponder(user.Id, true)

	userUpdated1, err := th.App.GetUser(user.Id)
	require.Nil(t, err)
	assert.Equal(t, userUpdated1.NotifyProps["auto_responder_active"], "false")

	th.App.DisableAutoResponder(user.Id, true)

	userUpdated2, err := th.App.GetUser(user.Id)
	require.Nil(t, err)
	assert.Equal(t, userUpdated2.NotifyProps["auto_responder_active"], "false")
}

func TestSendAutoResponseIfNecessary(t *testing.T) {
	t.Run("should send auto response when enabled", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		receiver := th.CreateUser()

		patch := &model.UserPatch{
			NotifyProps: map[string]string{
				"auto_responder_active":  "true",
				"auto_responder_message": "Hello, I'm unavailable today.",
			},
		}
		receiver, err := th.App.PatchUser(receiver.Id, patch, true)
		require.Nil(t, err)

		channel := th.CreateDmChannel(receiver)

		savedPost, _ := th.App.CreatePost(th.Context, &model.Post{
			ChannelId: channel.Id,
			Message:   "zz" + model.NewId() + "a",
			UserId:    th.BasicUser.Id},
			th.BasicChannel,
			false, true)

		sent, err := th.App.SendAutoResponseIfNecessary(th.Context, channel, th.BasicUser, savedPost)

		assert.Nil(t, err)
		assert.True(t, sent)
	})

	t.Run("should not send auto response when disabled", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		receiver := th.CreateUser()

		patch := &model.UserPatch{
			NotifyProps: map[string]string{
				"auto_responder_active":  "false",
				"auto_responder_message": "Hello, I'm unavailable today.",
			},
		}
		receiver, err := th.App.PatchUser(receiver.Id, patch, true)
		require.Nil(t, err)

		channel := th.CreateDmChannel(receiver)

		savedPost, _ := th.App.CreatePost(th.Context, &model.Post{
			ChannelId: channel.Id,
			Message:   "zz" + model.NewId() + "a",
			UserId:    th.BasicUser.Id},
			th.BasicChannel,
			false, true)

		sent, err := th.App.SendAutoResponseIfNecessary(th.Context, channel, th.BasicUser, savedPost)

		assert.Nil(t, err)
		assert.False(t, sent)
	})

	t.Run("should not send auto response for non-DM channel", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		savedPost, _ := th.App.CreatePost(th.Context, &model.Post{
			ChannelId: th.BasicChannel.Id,
			Message:   "zz" + model.NewId() + "a",
			UserId:    th.BasicUser.Id},
			th.BasicChannel,
			false, true)

		sent, err := th.App.SendAutoResponseIfNecessary(th.Context, th.BasicChannel, th.BasicUser, savedPost)

		assert.Nil(t, err)
		assert.False(t, sent)
	})

	t.Run("should not send auto response for bot", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		receiver := th.CreateUser()

		patch := &model.UserPatch{
			NotifyProps: map[string]string{
				"auto_responder_active":  "true",
				"auto_responder_message": "Hello, I'm unavailable today.",
			},
		}
		receiver, err := th.App.PatchUser(receiver.Id, patch, true)
		require.Nil(t, err)

		channel := th.CreateDmChannel(receiver)

		bot, err := th.App.CreateBot(th.Context, &model.Bot{
			Username:    "botusername",
			Description: "bot",
			OwnerId:     th.BasicUser.Id,
		})
		assert.Nil(t, err)

		botUser, err := th.App.GetUser(bot.UserId)
		assert.Nil(t, err)

		savedPost, _ := th.App.CreatePost(th.Context, &model.Post{
			ChannelId: channel.Id,
			Message:   "zz" + model.NewId() + "a",
			UserId:    botUser.Id},
			th.BasicChannel,
			false, true)

		sent, err := th.App.SendAutoResponseIfNecessary(th.Context, channel, botUser, savedPost)

		assert.Nil(t, err)
		assert.False(t, sent)
	})

	t.Run("should send auto response in dm channel if not already sent today", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		receiver := th.CreateUser()

		patch := &model.UserPatch{
			NotifyProps: map[string]string{
				"auto_responder_active":  "true",
				"auto_responder_message": "Hello, I'm unavailable today.",
			},
		}
		receiver, err := th.App.PatchUser(receiver.Id, patch, true)
		require.Nil(t, err)

		channel := th.CreateDmChannel(receiver)

		savedPost, err := th.App.CreatePost(th.Context, &model.Post{
			ChannelId: channel.Id,
			Message:   NewTestId(),
			UserId:    th.BasicUser.Id},
			th.BasicChannel,
			false, true)

		assert.Nil(t, err)

		sent, err := th.App.SendAutoResponseIfNecessary(th.Context, channel, th.BasicUser, savedPost)

		require.Nil(t, err)
		assert.True(t, sent)

		sent, err = th.App.SendAutoResponseIfNecessary(th.Context, channel, th.BasicUser, savedPost)

		require.Nil(t, err)
		assert.False(t, sent)
	})
}

func TestSendAutoResponseSuccess(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(th.Context, user)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_responder_active"] = "true"
	patch.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	userUpdated1, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)

	savedPost, _ := th.App.CreatePost(th.Context, &model.Post{
		ChannelId: th.BasicChannel.Id,
		Message:   "zz" + model.NewId() + "a",
		UserId:    th.BasicUser.Id},
		th.BasicChannel,
		false, true)

	sent, err := th.App.SendAutoResponse(th.Context, th.BasicChannel, userUpdated1, savedPost)

	assert.Nil(t, err)
	assert.True(t, sent)

	list, err := th.App.GetPosts(th.BasicChannel.Id, 0, 1)
	require.Nil(t, err)

	autoResponderPostFound := false
	for _, post := range list.Posts {
		if post.Type == model.PostTypeAutoResponder {
			autoResponderPostFound = true
			assert.Equal(t, savedPost.Id, post.RootId)
		}
	}
	assert.True(t, autoResponderPostFound)
}

func TestSendAutoResponseSuccessOnThread(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(th.Context, user)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_responder_active"] = "true"
	patch.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	userUpdated1, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)

	parentPost, _ := th.App.CreatePost(th.Context, &model.Post{
		ChannelId: th.BasicChannel.Id,
		Message:   "zz" + model.NewId() + "a",
		UserId:    th.BasicUser.Id},
		th.BasicChannel,
		false, true)

	savedPost, _ := th.App.CreatePost(th.Context, &model.Post{
		ChannelId: th.BasicChannel.Id,
		Message:   "zz" + model.NewId() + "a",
		UserId:    th.BasicUser.Id,
		RootId:    parentPost.Id,
	},
		th.BasicChannel,
		false, true)

	sent, err := th.App.SendAutoResponse(th.Context, th.BasicChannel, userUpdated1, savedPost)

	assert.Nil(t, err)
	assert.True(t, sent)

	list, err := th.App.GetPosts(th.BasicChannel.Id, 0, 1)
	require.Nil(t, err)

	autoResponderPostFound := false
	for _, post := range list.Posts {
		if post.Type == model.PostTypeAutoResponder {
			autoResponderPostFound = true
			assert.Equal(t, savedPost.RootId, post.RootId)
		}
	}
	assert.True(t, autoResponderPostFound)
}

func TestSendAutoResponseFailure(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user := th.CreateUser()
	defer th.App.PermanentDeleteUser(th.Context, user)

	patch := &model.UserPatch{}
	patch.NotifyProps = make(map[string]string)
	patch.NotifyProps["auto_responder_active"] = "false"
	patch.NotifyProps["auto_responder_message"] = "Hello, I'm unavailable today."

	userUpdated1, err := th.App.PatchUser(user.Id, patch, true)
	require.Nil(t, err)

	savedPost, _ := th.App.CreatePost(th.Context, &model.Post{
		ChannelId: th.BasicChannel.Id,
		Message:   "zz" + model.NewId() + "a",
		UserId:    th.BasicUser.Id},
		th.BasicChannel,
		false, true)

	sent, err := th.App.SendAutoResponse(th.Context, th.BasicChannel, userUpdated1, savedPost)

	assert.Nil(t, err)
	assert.False(t, sent)

	if list, err := th.App.GetPosts(th.BasicChannel.Id, 0, 1); err != nil {
		require.Nil(t, err)
	} else {
		autoResponderPostFound := false
		for _, post := range list.Posts {
			if post.Type == model.PostTypeAutoResponder {
				autoResponderPostFound = true
			}
		}
		assert.False(t, autoResponderPostFound)
	}
}
