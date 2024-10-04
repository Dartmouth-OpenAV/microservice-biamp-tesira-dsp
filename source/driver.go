package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"strconv"
	"math"
	"github.com/Dartmouth-OpenAV/microservice-framework/framework"
)

// GLOBAL VARIABLES

// Sends the command to the DSP. 
func convertAndSend(socketKey string, cmdStr string) bool {
	function := "convertAndSend"

	framework.Log(fmt.Sprint("Command sent: ", cmdStr))

	sent := framework.WriteLineToSocket(socketKey, cmdStr)

	if !sent {
		errMsg := fmt.Sprintf(function + " - h3okxu3 - error sending command")
		framework.AddToErrors(socketKey, errMsg)
	}

	return sent
}
// Reads a response from the DSP
func readAndConvert(socketKey string) (string, string, error) {
	function := "readAndConvert"
	resp := framework.ReadLineFromSocket(socketKey)

	// Normally, there is an acknowledgement response or error message.
	if resp == "" {
		errMsg := function + " - k3kxlpo - Response was blank."
		framework.AddToErrors(socketKey, errMsg)
		return "", errMsg, errors.New(errMsg)
	}

	return resp, "", nil
}
// Handles the Telnet negotiation for the Tesira connection
func loginNegotiation(socketKey string) bool {
	function := "loginNegotiation"
	count := 0
	welcomeMsg := false
	// Breaks if the negotiations go over 7 rounds to avoid an infinite loop.
	// Normal negotiations so far are 3-4 rounds.
	for count < 7 {
		count += 1
		negotiationMsg := ""
		negotiationResp, errMsg, err := readAndConvert(socketKey)
		if err != nil {
			framework.AddToErrors(socketKey, errMsg)
		}
		respHex := fmt.Sprintf("%x", negotiationResp)
		framework.Log("Printing Negotiation from Biamp: " + respHex)
		if negotiationResp == "Welcome to the Tesira Text Protocol Server..." {
			negotiationMsg = ""
			welcomeMsg = true
		} else if negotiationResp == "" && welcomeMsg == true {
			framework.Log("Negotiations are over")
			return true
		} else {
			negotiationMsg = strings.Replace(respHex, "fd", "fc", -1)
			negotiationMsg = strings.Replace(negotiationMsg, "fb", "fe", -1)
		}
		negotiationHex, err := hex.DecodeString(negotiationMsg)
		if err != nil {
			framework.AddToErrors(socketKey, err.Error())
		}
		framework.Log("Printing Response to Biamp: " + fmt.Sprintf("%x", negotiationHex))
		convertAndSend(socketKey, string(negotiationHex))
	}
	errMsg := function + " - mrk42 - Stopped negotiation loop after 7 rounds to avoid infinite loop."
	framework.AddToErrors(socketKey, errMsg)

	return false
}
// Takes value from the range 0-100 and transforms it to the range the Biamp uses (-100 - +12).
func transformVolume (vol string) string {
	function := "transformVolume"
	floatVol, err := strconv.ParseFloat(vol, 32)
	if err != nil {
		errMsg := fmt.Sprintf(function+" - Error converting volume r23dfs %v", err.Error())
		return errMsg
	}
	// take care of min case
	if floatVol < 0.369786371648 {
		floatVol = 0.369786371648
	}
	// convert loudness (volume) to decibels
	floatVol = 20.0*(math.Log(floatVol/100.0))+12.0
	// converts the float to a string
	stringVol := strconv.FormatFloat(floatVol, 'f', 1, 32)
	framework.Log(stringVol)

	return stringVol
}
// Takes value from the Biamp and transforms it to the range 0-100 for the GUI.
func unTransformVolume (vol string) string {
	function := "unTransformVolume"
	floatVol, err := strconv.ParseFloat(vol, 32)

	if err != nil {
		errMsg := fmt.Sprintf(function+" - Error converting volume 345rds %v", err.Error())
		return errMsg
	}
	// convert decibels to loudness (volume)
	floatVol = math.Exp((floatVol-12.0)/20.0)*100.0
	// converts the float to a string
	stringVol := strconv.FormatFloat(floatVol, 'f', 0, 32)
	framework.Log(stringVol)

	return stringVol
}

//MAIN FUNCTIONS

// GET Functions

func getVolume(socketKey string, instanceTag string, channel string) (string, error) {
	function := "getVolume"

	value := `"unknown"`
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = getVolumeDo(socketKey, instanceTag, channel)
		if value == `"unknown"` { // Something went wrong - perhaps try again
			framework.Log(function + " - fq3sdvc - retrying volume operation")
			maxRetries--
			time.Sleep(1 * time.Second)
			if maxRetries == 0 {
				errMsg := fmt.Sprintf(function + "f839dk4 - max retries reached")
				framework.AddToErrors(socketKey, errMsg)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	return value, err
}
// Gets the volume level of the specified instance tag and channel. Returns a value between 0 and 100.
func getVolumeDo(socketKey string, instanceTag string, channel string) (string, error) {
	function := "getVolumeDo"

	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected{
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - h3okxu3 - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	cmdString := instanceTag + " get level " + channel + "\r"

	sent := convertAndSend(socketKey, cmdString)

	if !sent {
		errMsg := fmt.Sprintf(function + " - dj3dke - error sending command")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	respArr, errMsg, err := readAndConvert(socketKey)

	if err != nil{
		return errMsg, err
	}

	if strings.HasPrefix(respArr, "+OK"){
		framework.Log("No error")
	} else if strings.HasPrefix(respArr, "-ERR"){
		errMsg := fmt.Sprintf(function + " - gkr5jdi - Error: " + respArr)
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	} else if strings.HasPrefix(respArr, instanceTag + " get level " + channel){
		framework.Log("GOT AN ECHO")
		respArr, errMsg, err = readAndConvert(socketKey)
	} else {
		errMsg := fmt.Sprintf(function + " - kcj3j - error sending command")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	_, value, found := strings.Cut(respArr, "\"value\":")
	if !found {
		errMsg := fmt.Sprintf(function + " - kcj3j - error reading response")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	normalizedVol := unTransformVolume(value)
	value = normalizedVol

	framework.Log(function + " - Decoded Response: "+ value)

	// If we got here, the response was good, so successful return with the state indication
	return `"` + value + `"`, nil
}
// Returns true if the channel is muted. False if it is not muted.
func getAudioMute(socketKey string, instanceTag string, channel string) (string, error) {
	function := "getAudioMute"

	value := `"unknown"`
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = getMuteToggleDo(socketKey, instanceTag, channel)
		if value == `"unknown"` { // Something went wrong - perhaps try again
			framework.Log(function + " - sdf09nd - retrying audiomute operation")
			maxRetries--
			time.Sleep(1 * time.Second)
			if maxRetries == 0 {
				errMsg := fmt.Sprintf(function + "f4fk5n3 - max retries reached")
				framework.AddToErrors(socketKey, errMsg)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	return value, err
}
// Returns true if Voice Lift is on, false if Voice Lift is off
func getVoiceLift(socketKey string, instanceTag string, channel string) (string, error) {
	function := "getVoiceLift"

	value := `"unknown"`
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = getMuteToggleDo(socketKey, instanceTag, channel)
		if value == `"unknown"` { // Something went wrong - perhaps try again
			framework.Log(function + " - sdf09nd - retrying voice lift operation")
			maxRetries--
			time.Sleep(1 * time.Second)
			if maxRetries == 0 {
				errMsg := fmt.Sprintf(function + "f4fk5n3 - max retries reached")
				framework.AddToErrors(socketKey, errMsg)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	//If Voice Lift mute is true, Voice Lift is "off". If mute is false, Voice Lift is "on"
	if value == "\"true\"" {
		value = "\"off\""
	} else if value == "\"false\"" {
		value = "\"on\""
	}

	return value, err
}
// Gets the mute status of the specified instance tag and channel. Returns true or false.
func getMuteToggleDo(socketKey string, instanceTag string, channel string) (string, error) {
	function := "getMuteToggleDo"

	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected {
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - h3okxu3 - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	cmdString := instanceTag + " get mute " + channel + "\r"

	sent := convertAndSend(socketKey, cmdString)

	if !sent {
		errMsg := fmt.Sprintf(function + " - i5kcfoe - error sending command")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	respArr, errMsg, err := readAndConvert(socketKey)

	if err != nil{
		return errMsg, err
	}

	if strings.HasPrefix(respArr, "+OK"){
		framework.Log("No error")
	} else if strings.HasPrefix(respArr, "-ERR"){
		errMsg := fmt.Sprintf(function + " - estg74 - Error: " + respArr)
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	} else if strings.HasPrefix(respArr, instanceTag + " get mute " + channel){
		framework.Log("GOT AN ECHO")
		respArr, errMsg, err = readAndConvert(socketKey)
	} else {
		errMsg := fmt.Sprintf(function + " - j5jcu - Not a standard response")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	_, value, found := strings.Cut(respArr, "\"value\":")
	if !found {
		errMsg := fmt.Sprintf(function + " - 2jxi4 - Could not get value from response")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}
	
	framework.Log(function + " - Decoded Response: "+ value)

	// If we got here, the response was good, so successful return with the state indication
	return `"`+ value + `"`, nil
}
// Returns true if Logic Selector is true, false if Logic Selector is false
func getLogicSelector(socketKey string, instanceTag string, channel string) (string, error) {
	function := "getLogicSelector"

	value := `"unknown"`
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = getStateToggleDo(socketKey, instanceTag, channel)
		if value == `"unknown"` { // Something went wrong - perhaps try again
			framework.Log(function + " - 94ndk3l - retrying getLogicSelector operation")
			maxRetries--
			time.Sleep(1 * time.Second)
			if maxRetries == 0 {
				errMsg := fmt.Sprintf(function + "aoi5pj2 - max retries reached")
				framework.AddToErrors(socketKey, errMsg)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	return value, err
}
// Returns the channel number that is set to true
func getAudioMode(socketKey string, instanceTag string) (string, error) {
	function := "getAudioMode"

	value := `"unknown"`
	err := error(nil)
	maxRetries := 2
	channel := 1
	for maxRetries > 0 {
		channel = 1
		// Loop through 5 channels to find which is set to true
		for channel <= 5 {
			value, err = getStateToggleDo(socketKey, instanceTag, strconv.Itoa(channel))

			if err != nil { // Something went wrong - perhaps try again
				framework.Log(function + " - 54ijxl - retrying getAudioMode operation")
				maxRetries--
				time.Sleep(1 * time.Second)

				if maxRetries == 0 {
					errMsg := fmt.Sprintf(function + "2jj3hx - max retries reached")
					framework.AddToErrors(socketKey, errMsg)
					break
				}
			} else { // Got a response
				// If one channel is true, return 
				if value == "\"true\"" {
					maxRetries = 0
					break
				} else if value == "\"false\"" {
					channel++
					if channel >= 6 {
						maxRetries--
					}
				}
			}
		}
	}

	// If the for loop broke with a channel returning true, return the channel number
	if value == "\"true\""{
		return `"`+strconv.Itoa(channel)+`"`, err
	} else {
		return `"unknown"`, err
	}
}
// Gets the state of the specified instance tag and channel. Returns true or false.
func getStateToggleDo(socketKey string, instanceTag string, channel string) (string, error) {
	function := "getStateToggleDo"

	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected {
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - jl3kldj - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	cmdString := instanceTag + " get state " + channel + "\r"

	sent := convertAndSend(socketKey, cmdString)

	if !sent {
		errMsg := fmt.Sprintf(function + " - y2hgdh - error sending command")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	respArr, errMsg, err := readAndConvert(socketKey)

	if err != nil{
		return errMsg, err
	}

	if strings.HasPrefix(respArr, "+OK"){
		framework.Log("No error")
	} else if strings.HasPrefix(respArr, "-ERR"){
		errMsg := fmt.Sprintf(function + " - i3jdieo - Error: " + respArr)
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	} else if strings.HasPrefix(respArr, instanceTag + " get state " + channel){
		framework.Log("GOT AN ECHO")
		respArr, errMsg, err = readAndConvert(socketKey)
	} else {
		errMsg := fmt.Sprintf(function + " - h4hxi - Not a standard response")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	_, value, found := strings.Cut(respArr, "\"value\":")
	if !found {
		errMsg := fmt.Sprintf(function + " - i3usho4 - Could not get value from response")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}
	
	framework.Log(function + " - Decoded Response: "+ value)

	// If we got here, the response was good, so successful return with the state indication
	return `"`+ value + `"`, nil
}

//SET Functions

func setVolume(socketKey string, instanceTag string, channel string, volume string) (string, error) {
	function := "setVolume"

	value := "notok"
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = setVolumeDo(socketKey, instanceTag, channel, volume)
		if value != "ok" { // Something went wrong - perhaps try again
			framework.Log(function + " - fq3sdvc - retrying volume operation")
			maxRetries--
			time.Sleep(1 * time.Second)
			if maxRetries == 0 {
				errMsg := fmt.Sprintf(function + " - fds3nf3 - max retries reached")
				framework.AddToErrors(socketKey, errMsg)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	return value, err
}
// Sets the volume for the specified instance tag and channel. Takes a value from 0-100.
func setVolumeDo(socketKey string, instanceTag string, channel string, volume string) (string, error) {
	function := "setVolumeDo"
	volume = strings.Trim(volume, "\"")

	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected {
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - h3okxu3 - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	transformedVol := transformVolume(volume)
	framework.Log("Transformed Volume: " + transformedVol)

	cmdString := instanceTag + " set level " + channel + " " + transformedVol + "\r"

	sent := convertAndSend(socketKey, cmdString)

	if !sent {
		errMsg := fmt.Sprintf(function + " - i5kcfoe - error sending command")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	respArr, errMsg, err := readAndConvert(socketKey)

	if err != nil{
		return errMsg, err
	}

	if strings.HasPrefix(respArr, "+OK"){
		framework.Log("No error")
	} else if strings.HasPrefix(respArr, "-ERR"){
		errMsg := fmt.Sprintf(function + " - uygv8 - Error: " + respArr)
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	} else if strings.HasPrefix(respArr, instanceTag + " set level " + channel + " " + transformedVol) {
		framework.Log("GOT AN ECHO")
		readAndConvert(socketKey)
	} else {
		errMsg := fmt.Sprintf(function + " - j5jcu - Not a standard response")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}
	
	framework.Log(function + " - Decoded Response: "+ respArr)

	// If we got here, the response was good, so successful return with the state indication
	return "ok", nil
}
// Sets mute to true or false for the specified instance tag and channel.
func setAudioMute(socketKey string, instanceTag string, channel string, state string) (string, error) {
	function := "setAudioMute"

	value := "notok"
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = setMuteToggleDo(socketKey, instanceTag, channel, state)
		if value != "ok" { // Something went wrong - perhaps try again
			framework.Log(function + " - fq5dhs - retrying audiomute operation")
			maxRetries--
			time.Sleep(1 * time.Second)
			if maxRetries == 0 {
				errMsg := fmt.Sprintf(function + " - 03kfl4d - max retries reached")
				framework.AddToErrors(socketKey, errMsg)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	return value, err
}
// Turns Voice Lift on or off
func setVoiceLift(socketKey string, instanceTag string, channel string, state string) (string, error) {
	function := "setVoiceLift"
	state = strings.Trim(state, "\"")

	// Flipping to make the GUI button make sense.
	// Button on - Voice Lift unmuted. Button off - Voice Lift muted.
	if state == "on" {
		state = "false"
	} else if state == "off" {
		state = "true"
	}

	value := "notok"
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = setMuteToggleDo(socketKey, instanceTag, channel, state)
		if value != "ok" { // Something went wrong - perhaps try again
			framework.Log(function + " - 3j3md3 - retrying voicelift operation")
			maxRetries--
			time.Sleep(1 * time.Second)
			if maxRetries == 0 {
				errMsg := fmt.Sprintf(function + " - 5h4ne3 - max retries reached")
				framework.AddToErrors(socketKey, errMsg)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	return value, err
}
// Sets mute to true or false for the specified instance tag and channel.
func setMuteToggleDo(socketKey string, instanceTag string, channel string, state string) (string, error) {
	function := "setMuteToggleDo"
	state = strings.Trim(state, "\"")
	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected {
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - h3okxu3 - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	cmdString := instanceTag + " set mute " + channel + " " + state + "\r"

	sent := convertAndSend(socketKey, cmdString)

	if !sent {
		errMsg := fmt.Sprintf(function + " - i5kcfoe - error sending command")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	respArr, errMsg, err := readAndConvert(socketKey)

	if err != nil{
		return errMsg, err
	}

	if strings.HasPrefix(respArr, "+OK"){
		framework.Log("No error")
	} else if strings.HasPrefix(respArr, "-ERR"){
		errMsg := fmt.Sprintf(function + " - uiub8 - Error: " + respArr)
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	} else if strings.HasPrefix(respArr, instanceTag + " set mute " + channel + " " + state) {
		framework.Log("GOT AN ECHO")
		readAndConvert(socketKey)
	} else {
		errMsg := fmt.Sprintf(function + " - j5jcu - Not a standard response")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}
	
	framework.Log(function + " - Decoded Response: "+ respArr)

	// If we got here, the response was good, so successful return with the state indication
	return "ok", nil
}
func setPreset(socketKey string, presetID string) (string, error) {
	function := "setPreset"

	value := "notok"
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = setPresetDo(socketKey, presetID)
		if value != "ok" { // Something went wrong - perhaps try again
			framework.Log(function + " - k5kifj - retrying preset operation")
			maxRetries--
			time.Sleep(1 * time.Second)
			if maxRetries == 0 {
				errMsg := fmt.Sprintf(function + " - sj34h - max retries reached")
				framework.AddToErrors(socketKey, errMsg)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	return value, err
}
// Recalls a device preset for the DSP by ID. Preset ID must be greater than 1001.
func setPresetDo(socketKey string, presetID string) (string, error) {
	function := "setPresetDo"

	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected {
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - h3okxu3 - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	cmdString := "DEVICE recallPreset " + presetID + "\r"

	sent := convertAndSend(socketKey, cmdString)

	if !sent {
		errMsg := fmt.Sprintf(function + " - i5kcfoe - error sending command")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	respArr, errMsg, err := readAndConvert(socketKey)

	if err != nil{
		return errMsg, err
	}

	if strings.HasPrefix(respArr, "+OK"){
		framework.Log("No error")
	} else if strings.HasPrefix(respArr, "-ERR"){
		errMsg := fmt.Sprintf(function + " - kgy7bh - Error: " + respArr)
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	} else if strings.HasPrefix(respArr, "DEVICE recallPreset " + presetID) {
		framework.Log("GOT AN ECHO")
		readAndConvert(socketKey)
	} else {
		errMsg := fmt.Sprintf(function + " - j5jcu - Not a standard response")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}
	
	framework.Log(function + " - Decoded Response: "+ respArr)

	// If we got here, the response was good, so successful return with the state indication
	return "ok", nil
}
// Sets state to true or false for the specified instance tag and channel.
func setLogicSelector(socketKey string, instanceTag string, channel string, state string) (string, error) {
	function := "setLogicSelector"

	value := "notok"
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = setStateToggleDo(socketKey, instanceTag, channel, state)
		if value != "ok" { // Something went wrong - perhaps try again
			framework.Log(function + " - ir3jdd - retrying setLogicSelector operation")
			maxRetries--
			time.Sleep(1 * time.Second)
			if maxRetries == 0 {
				errMsg := fmt.Sprintf(function + " - 9jcj8k - max retries reached")
				framework.AddToErrors(socketKey, errMsg)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	return value, err
}
// Sets state to true for the specified instance tag and channel.
func setAudioMode(socketKey string, instanceTag string, channel string) (string, error) {
	function := "setAudioMode"
	channel = strings.Trim(channel, "\"")

	value := "notok"
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = setStateToggleDo(socketKey, instanceTag, channel, "true")
		if value != "ok" { // Something went wrong - perhaps try again
			framework.Log(function + " - oj5jex - retrying setAudioMode operation")
			maxRetries--
			time.Sleep(1 * time.Second)
			if maxRetries == 0 {
				errMsg := fmt.Sprintf(function + " - 5lsm3g - max retries reached")
				framework.AddToErrors(socketKey, errMsg)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	return value, err
}
// Sets state to true or false for the specified instance tag and channel.
func setStateToggleDo(socketKey string, instanceTag string, channel string, state string) (string, error) {
	function := "setStateToggleDo"
	state = strings.Trim(state, "\"")
	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected {
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - u2nj45l - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	cmdString := instanceTag + " set state " + channel + " " + state + "\r"

	sent := convertAndSend(socketKey, cmdString)

	if !sent {
		errMsg := fmt.Sprintf(function + " - 95nckx - error sending command")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}

	respArr, errMsg, err := readAndConvert(socketKey)

	if err != nil{
		return errMsg, err
	}

	if strings.HasPrefix(respArr, "+OK"){
		framework.Log("No error")
	} else if strings.HasPrefix(respArr, "-ERR"){
		errMsg := fmt.Sprintf(function + " - k3k2md - Error: " + respArr)
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	} else if strings.HasPrefix(respArr, instanceTag + " set state " + channel + " " + state) {
		framework.Log("GOT AN ECHO")
		readAndConvert(socketKey)
	} else {
		errMsg := fmt.Sprintf(function + " - kfci4kd - Not a standard response")
		framework.AddToErrors(socketKey, errMsg)
		return errMsg, errors.New(errMsg)
	}
	
	framework.Log(function + " - Decoded Response: "+ respArr)

	// If we got here, the response was good, so successful return with the state indication
	return "ok", nil
}