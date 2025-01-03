package controller

import "github.com/sychonet/vdash-be/service"

type ServerController struct {
	scalewayService  *service.ScalewayService
	dbService        *service.DatabaseService
	libvirtService   *service.LibvirtService
	schedulerService *service.SchedulerService
}

func NewServerController(scalewayService *service.ScalewayService, dbService *service.DatabaseService, libvirtService *service.LibvirtService, schedulerService *service.SchedulerService) *ServerController {
	return &ServerController{
		scalewayService:  scalewayService,
		dbService:        dbService,
		libvirtService:   libvirtService,
		schedulerService: schedulerService,
	}
}
