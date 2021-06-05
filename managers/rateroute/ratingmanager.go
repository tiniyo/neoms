package rateroute

import (
	"encoding/json"
	"fmt"
	"net"
	"github.com/neoms/config"
	"github.com/neoms/helper"
	"github.com/neoms/logger"
	"github.com/neoms/managers/callstats"
	"github.com/neoms/models"
	"strings"
)

func GetOutboundRateRoutes(callSid, vendorAuthId, authId, phoneNumber string) (string, *models.RatingRoutingResponse) {
	baseUrl := config.Config.Rating.BaseUrl
	region := config.Config.Rating.Region
	url := fmt.Sprintf("%s/%s/Tenants/%s/%s/%s", baseUrl, vendorAuthId, authId, region, phoneNumber)
	if vendorAuthId == "" || vendorAuthId == "TINIYO1SECRET1AUTHID" {
		vendorAuthId = "TINIYO1SECRET1AUTHID"
		url = fmt.Sprintf("%s/%s/Tenants/%s/%s/%s", baseUrl, vendorAuthId, "DEFAULT", region, phoneNumber)
	}

	logger.UuidLog("Info", callSid, fmt.Sprint("fetching rates with url - ", url))

	var data models.RatingRoutingResponse
	if statusCode, respBody, err := helper.Get(callSid, nil, url); statusCode != 200 || err != nil || respBody == nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("Error while fetching rates - failed", err))
		return "failed", nil
	} else if err = json.Unmarshal(respBody, &data); err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("Error while unmarshal rates - failed", err))
		return "failed", nil
	}

	logger.UuidLog("Info", callSid, fmt.Sprint("termination route response is - ", data.Term))
	logger.UuidLog("Info", callSid, fmt.Sprint("Origination rate response is - ", data.Orig))

	return "success", &data
}

func GetSetOutboundRateRoutes(data *models.CallRequest, destNumber string) *models.InternalRateRoute {
	routingString := ""
	var rateRouteRes = new(models.InternalRateRoute)
	data.CallResponse.Direction = "outbound-call"
	if net.ParseIP(destNumber) != nil ||
		strings.HasPrefix(destNumber, "sip:") || strings.Contains(destNumber, "@") {
		if !strings.HasPrefix(data.To, "sip:") {
			destNumber = fmt.Sprintf("sip:%s", data.To)
		}
	} else {
		destNumber = helper.NumberSanity(destNumber)
	}
	callSid := data.CallSid
	status, rateRoutes := GetOutboundRateRoutes(callSid, data.VendorAuthID, data.AccountSid, destNumber)
	if status == "failed" || rateRoutes == nil {
		logger.Logger.Error("Failed to get rates, rate not found")
		return nil
	}
	pulse := rateRoutes.Orig.InitialPulse
	data.Pulse = pulse
	perSecondRate := float64(rateRoutes.Orig.Rate / 60)
	rateInPulse := perSecondRate * float64(pulse)
	data.Rate = rateInPulse
	data.To = destNumber

	var routingTokenArray = helper.JwtTokenInfos{}

	switch data.DestType {
	case "sip", "Sip":
		logger.Logger.Info("SIP Destination, Skipping termination route processing")
	case "number", "Number":
		data.DestType = "number"
		logger.Logger.Info("Number Destination, termination route processing")
		if rateRoutes.Term == nil {
			logger.Logger.Error("No routes found, exit the call")
			return nil
		}
		for _, rt := range rateRoutes.Term {
			rateRouteRes.RemovePrefix = rt.RemovePrefix
			rateRouteRes.TrunkPrefix = rt.TrunkPrefix
			rateRouteRes.SipPilotNumber = rt.SipPilotNumber
			var routingToken = helper.JwtTokenInfo{}
			if rt.RemovePrefix != "" {
				destNumber = strings.TrimPrefix(destNumber, "+")
				destNumber = strings.TrimPrefix(destNumber, rt.RemovePrefix)
			}
			if rt.FromRemovePrefix != "" {
				rateRouteRes.FromRemovePrefix = rt.FromRemovePrefix
			}
			if rt.TrunkPrefix != "" {
				destNumber = fmt.Sprintf("%s%s", rt.TrunkPrefix, destNumber)
			}
			if routingString == "" {
				routingString = fmt.Sprintf("sip:%s@%s", destNumber, rt.PrimaryIP)
			} else {
				routingString = fmt.Sprintf("%s^sip:%s@%s", routingString, destNumber, rt.PrimaryIP)
			}
			if rt.Username != "" {
				routingToken.Ip = rt.PrimaryIP
				routingToken.Username = rt.Username
				routingToken.Password = rt.Password
				routingTokenArray = append(routingTokenArray, routingToken)
			}
			if rt.FailoverIP != "" {
				routingString = fmt.Sprintf("%s^sip:%s@%s", routingString, destNumber, rt.FailoverIP)
				if rt.Username != "" {
					routingToken.Ip = rt.FailoverIP
					routingToken.Username = rt.Username
					routingToken.Password = rt.Password
					routingTokenArray = append(routingTokenArray, routingToken)
				}
			}
		}
	default:
	}

	logger.UuidLog("Info", data.CallSid, fmt.Sprintf("Routing string for call is - %s", routingString))

	if rateRouteRes.FromRemovePrefix != "" {
		data.FromRemovePrefix = rateRouteRes.FromRemovePrefix
		logger.UuidLog("Info", data.CallSid, fmt.Sprintf("From Remove Prefix is  - %s", data.FromRemovePrefix))
	}

	/* Store Call State */
	err := callstats.SetCallDetailByUUID(data)
	if err != nil {
		logger.Logger.Error("SetCallState Failed", err)
	}

	/* Get JWT token for username and password based trunk */
	if routingTokenArray != nil && len(routingTokenArray) > 0 {
		if jwtRouteToken, err := helper.CreateToken(routingTokenArray); err == nil {
			rateRouteRes.RoutingUserAuthToken = jwtRouteToken
		}
	}

	rateRouteRes.Pulse = pulse
	rateRouteRes.PulseRate = rateInPulse
	rateRouteRes.RoutingGatewayString = routingString
	return rateRouteRes
}
