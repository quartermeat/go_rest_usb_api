package usbAPI

//state of api connection should be internal
type connection struct {
}

type usbAPI interface {
	NewUSBAPI() connection
}

func NewUSBAPI() connection {

}
