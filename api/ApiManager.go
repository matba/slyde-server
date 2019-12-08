package api

import "sync"

var controller Controller
var mutex sync.Mutex

// GetController Get an instance of cache
func GetController() Controller {
	mutex.Lock()
	if controller == nil {
		controller = newControllerInstance()
	}
	mutex.Unlock()
	return controller
}
