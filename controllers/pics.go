package controllers

import "supermentor/models"

// 首页广告列表
func GetPics(route *SmRoute) {
	position := route.Request.Form.Get("position")
	if position == "" {
		route.ReturnJson(CODE_OK, "", "")
		return
	}
	ads := models.GetAds(position)
	route.ReturnJson(CODE_OK, "", ads)
	return
}
