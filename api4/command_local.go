// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/audit"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

func (api *API) InitCommandLocal() {
	api.BaseRoutes.Commands.Handle("", api.APILocal(localCreateCommand)).Methods("POST")
	api.BaseRoutes.Commands.Handle("", api.APILocal(listCommands)).Methods("GET")

	api.BaseRoutes.Command.Handle("", api.APILocal(getCommand)).Methods("GET")
	api.BaseRoutes.Command.Handle("", api.APILocal(updateCommand)).Methods("PUT")
	api.BaseRoutes.Command.Handle("/move", api.APILocal(moveCommand)).Methods("PUT")
	api.BaseRoutes.Command.Handle("", api.APILocal(deleteCommand)).Methods("DELETE")
}

func localCreateCommand(c *Context, w http.ResponseWriter, r *http.Request) {
	cmd := model.CommandFromJson(r.Body)
	if cmd == nil {
		c.SetInvalidParam("command")
		return
	}

	auditRec := c.MakeAuditRecord("localCreateCommand", audit.Fail)
	defer c.LogAuditRec(auditRec)
	c.LogAudit("attempt")

	rcmd, err := c.App.CreateCommand(cmd)
	if err != nil {
		c.Err = err
		return
	}

	auditRec.Success()
	c.LogAudit("success")
	auditRec.AddMeta("command", rcmd)

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(rcmd); err != nil {
		mlog.Warn("Error while writing response", mlog.Err(err))
	}
}
