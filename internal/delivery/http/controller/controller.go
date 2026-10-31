package controller

type Controller struct {
	user UserUsecaseItf
	auth AuthUsecaseItf
}

func NewController(user UserUsecaseItf, auth AuthUsecaseItf) *Controller {
	return &Controller{user: user, auth: auth}
}
