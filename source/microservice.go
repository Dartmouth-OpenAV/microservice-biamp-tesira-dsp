package main

import (
	"errors"
	"github.com/Dartmouth-OpenAV/microservice-framework/framework"
)

func setFrameworkGlobals() {
	// globals that change modes in the microservice framework:
	framework.DefaultSocketPort = 23 // Biamp's default socket port
	framework.CheckFunctionAppendBehavior = "Remove older instance"
	framework.MicroserviceName = "OpenAV Biamp Microservice"
	framework.UseTelnet = true
	framework.KeepAlive = true

	framework.RegisterMainGetFunc(doDeviceSpecificGet)
	framework.RegisterMainSetFunc(doDeviceSpecificSet)
}

// Every microservice using this golang microservice framework needs to provide this function to invoke functions to do sets.
// socketKey is the network connection for the framework to use to communicate with the device.
// setting is the first parameter in the URI.
// arg1 are the second and third parameters in the URI.
//   Example PUT URIs that will result in this function being invoked:
// 	 ":address/:setting/"
//   ":address/:setting/:arg1"
//   ":address/:setting/:arg1/:arg2"
func doDeviceSpecificSet(socketKey string, setting string, arg1 string, arg2 string, arg3 string) (string, error) {
	function := "doDeviceSpecificSet"

	// Add a case statement for each set function your microservice implements.  These calls can use 0, 1, or 2 arguments.
	switch setting {
		case "volume":
			return setVolume(socketKey, arg1, arg2, arg3)
		case "audiomute":
			return setAudioMute(socketKey, arg1, arg2, arg3)
		case "preset":
			return setPreset(socketKey, arg1)
		case "voicelift":
			return setVoiceLift(socketKey, arg1, arg2, arg3)
		case "logicselector":
			return setLogicSelector(socketKey, arg1, arg2, arg3)
		case "audiomode":
			return setAudioMode(socketKey, arg1, arg2)
	}

	// If we get here, we didn't recognize the setting.  Send an error back to the config writer who had a bad URL.
	errMsg := function + " - unrecognized setting in URI: " + setting
	framework.AddToErrors(socketKey, errMsg)
	err := errors.New(errMsg)
	return setting, err
}

// Every microservice using this golang microservice framework needs to provide this function to invoke functions to do gets.
// socketKey is the network connection for the framework to use to communicate with the device.
// setting is the first parameter in the URI.
// arg1 are the second and third parameters in the URI.
//   Example GET URIs that will result in this function being invoked:
// 	 ":address/:setting/"
//   ":address/:setting/:arg1"
//   ":address/:setting/:arg1/:arg2"
func doDeviceSpecificGet(socketKey string, setting string, arg1 string, arg2 string) (string, error) {
	function := "doDeviceSpecificGet"

	switch setting {
		case "volume":
			value, err := getVolume(socketKey, arg1, arg2)
			return value, err
		case "audiomute":
			value, err := getAudioMute(socketKey, arg1, arg2)
			return value, err
		case "voicelift":
			value, err := getVoiceLift(socketKey, arg1, arg2)
			return value, err
		case "logicselector":
			value, err := getLogicSelector(socketKey, arg1, arg2)
			return value, err
		case "audiomode":
			value, err := getAudioMode(socketKey, arg1)
			return value, err
	}

	// If we get here, we didn't recognize the setting.  Send an error back to the config writer who had a bad URL.
	errMsg := function + " - unrecognized setting in URI: " + setting
	framework.AddToErrors(socketKey, errMsg)
	err := errors.New(errMsg)
	return setting, err
}

func main() {
	setFrameworkGlobals()
	framework.Startup()
}