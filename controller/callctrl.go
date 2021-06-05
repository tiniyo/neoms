package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/neoms/logger"
	"github.com/neoms/managers"
	"github.com/neoms/managers/callstats"
	"github.com/neoms/models"
)

type CallController struct {
	callManage *managers.CallManager
}

func (u *CallController) InitializeCallController() {
	u.callManage = new(managers.CallManager)
	u.callManage.InitCallManager()
	new(managers.CallBackManager).InitCallBackManager()
	new(callstats.CallStatManager).InitCallStatManager()
}

/*
	Create Call Request [POST]
 */
func (u CallController) CreateCall(c *gin.Context) {
	authSid := c.Param("account_id")
	callSid := c.Param("call_id")
	logger.UuidLog("Info", callSid, "call create request")
	cr := models.CallRequest{}
	var err error
	if err = c.BindJSON(&cr); err == nil {
		cr.AccountSid = authSid
		cr.CallSid = callSid
		cr.Sid = callSid
		callResp, err := u.callManage.CreateCall(&cr)
		//we need to get callResponse here
		if err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("JSON Parsing Failed :", err.Error()))
			c.JSON(http.StatusBadGateway, gin.H{"status": "failed", "request_uuid": cr.CallSid, "api_id": cr.CallSid})
			return
		}
		logger.UuidLog("Info", callSid, fmt.Sprint("call created success :"))
		c.JSON(http.StatusOK, callResp)
		return
	}
	logger.UuidLog("Err", callSid, fmt.Sprint("JSON Parsing Failed :", err.Error()))
	c.JSON(http.StatusBadGateway, gin.H{"status": "failed", "request_uuid": cr.CallSid, "api_id": cr.CallSid})
}

/*
	Update Call Request [PUT]
*/
func (u CallController) UpdateCall(c *gin.Context) {
	callSid := c.Param("call_id")
	logger.UuidLog("Info", callSid, "call update request")
	cr := models.CallUpdateRequest{}
	if err := c.BindJSON(&cr); err == nil {
		callResponse, err := u.callManage.UpdateCall(cr)
		if err != nil || callResponse == nil || callResponse.Sid == "" {
			logger.UuidLog("Err", callSid, fmt.Sprint("call update failed, call is not active :", err.Error()))
			c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "failed", "request_uuid": cr.Sid, "api_id": cr.Sid})
			return
		}
		logger.UuidLog("Info", callSid, fmt.Sprint("call updated success :"))
		c.JSON(http.StatusOK, callResponse)
	}
	c.JSON(http.StatusBadRequest, "Bad Request")
}

/*
	GET Call Request [GET]
*/
func (u CallController) GetCall(c *gin.Context) {
	accountID := c.Param("account_id")
	callID := c.Param("call_id")
	logger.Logger.Debug("Account ID :", accountID, " CallID :", callID)
	callResponse, err := u.callManage.GetCall(callID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "request_uuid": callID, "api_id": callID})
		return
	}
	c.JSON(http.StatusOK, callResponse)
	return
}

/*
	Delete Call Request [DELETE]
*/
func (u CallController) DeleteCall(c *gin.Context) {
	accountID := c.Param("account_id")
	callID := c.Param("call_id")
	logger.Logger.Debug("Account ID :", accountID, " CallID :", callID)
	u.callManage.DeleteCallWithReason(callID, "DELETE_API_HANGUP")
}
