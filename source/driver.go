package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

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
func readAndConvert(socketKey string) (string, error) {
	function := "readAndConvert"
	resp := framework.ReadLineFromSocket(socketKey)

	// Normally, there is an acknowledgement response or error message.
	if resp == "" {
		errMsg := function + " - k3kxlpo - response was blank"
		framework.AddToErrors(socketKey, errMsg)
		return "unknown", errors.New(errMsg)
	}

	return resp, nil
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
		negotiationResp, _ := readAndConvert(socketKey)
		respHex := fmt.Sprintf("%x", negotiationResp)

		if negotiationResp == "Welcome to the Tesira Text Protocol Server..." {
			negotiationMsg = ""
			welcomeMsg = true
			// Sometimes, the biamp sends more negotiation messages after welcome so not returning here
		} else if negotiationResp == "unknown" && welcomeMsg {
			framework.Log("Negotiations are over")
			return true
		} else if negotiationResp == "unknown" && !welcomeMsg {
			framework.Log("Ending negotiations. No response from the DSP.")
			return false
		} else {
			framework.Log("Printing Negotiation from Biamp: " + respHex)
			// Rejecting all negotiations
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

// Sends a command and checks that the response is valid. Otherwise, tries reading again.
func sendAndValidateResponse(socketKey string, cmdStr string, cmdType string, respType string) (string, error) {
	// Send the command. Return if there is an error.
	sent := convertAndSend(socketKey, cmdStr)
	if !sent {
		errMsg := "in34kf - unable to send command"
		return "unknown", errors.New(errMsg)
	}

	// Try to read at most 5 times if the response is not what is expected.
	// The DSP might respond with an echo or a response for a different command.
	maxRetries := 5
	validResponse := false
	var resp string
	var err error

	for maxRetries > 0 {
		resp, err = readAndConvert(socketKey)
		if err != nil {
			return resp, err
		}

		// Checking if the response is an echo of the sent command or an error
		if strings.TrimSpace(resp) == strings.TrimSpace(cmdStr) {
			framework.Log("Got an echo. Reading again")
			maxRetries--
			continue
		}
		if strings.HasPrefix(resp, "-ERR") {
			errMsg := fmt.Sprintf("gkr5jdi - Read error: " + resp)
			err = errors.New(errMsg)
			return resp, err
		}

		// Checking that the response matches what is expected for the cmdType and respType
		// For example, a volume query should return a number and a command should return +OK.
		if cmdType == "query" {
			_, value, found := strings.Cut(resp, "\"value\":")
			if found {
				if respType == "number" {
					// Valid if the response can be converted to a number
					_, err := strconv.ParseFloat(value, 32)
					if err == nil {
						resp = value
						validResponse = true
						break
					}
				} else if respType == "state" {
					// Valid if the response is true or false
					if value == "true" || value == "false" {
						resp = value
						validResponse = true
						break
					}
				}
			}
		} else if cmdType == "command" {
			// The response to a command should just be an acknowledgement of +OK
			if strings.HasPrefix(resp, "+OK") {
				validResponse = true
				break
			}
		}
		framework.Log("Resp did not match what was expected. Reading again")
		maxRetries--
	}

	// Check if the for loop broke successfully or unsuccessfully
	if validResponse {
		return resp, err
	} else {
		errMsg := "tried to read 5 times. no valid response from the biamp"
		err = errors.New(errMsg)
		return "unknown", err
	}
}

// Takes value from the range 0-100 and transforms it to the range the Biamp uses (-100 - +12).
func transformVolume(vol string) string {
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
	floatVol = 20.0*(math.Log(floatVol/100.0)) + 12.0
	// converts the float to a string
	stringVol := strconv.FormatFloat(floatVol, 'f', 1, 32)
	framework.Log(stringVol)

	return stringVol
}

// Takes value from the Biamp and transforms it to the range 0-100 for the GUI.
func unTransformVolume(vol string) string {
	function := "unTransformVolume"
	floatVol, err := strconv.ParseFloat(vol, 32)

	if err != nil {
		errMsg := fmt.Sprintf(function+" - Error converting volume 345rds %v", err.Error())
		return errMsg
	}
	// convert decibels to loudness (volume)
	floatVol = math.Exp((floatVol-12.0)/20.0) * 100.0
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
	if !connected {
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - h3okxu3 - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	cmdString := instanceTag + " get level " + channel + "\r"

	value, err := sendAndValidateResponse(socketKey, cmdString, "query", "number")

	if err != nil {
		return value, err
	}

	normalizedVolume := unTransformVolume(value)

	framework.Log(function + " - Decoded Response: " + normalizedVolume)

	// If we got here, the response was good, so successful return with the state indication
	return `"` + normalizedVolume + `"`, nil
}

func getGain(socketKey string, instanceTag string) (string, error) {
	function := "getGain"

	value := `"unknown"`
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = getGainDo(socketKey, instanceTag)
		if value == `"unknown"` { // Something went wrong - perhaps try again
			framework.Log(function + " - fq3sdvc - retrying gain operation")
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

// Gets the gain level of the specified instance tag. Returns a value between 0 and 100.
func getGainDo(socketKey string, instanceTag string) (string, error) {
	function := "getGainDo"

	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected {
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - h3okxu3 - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	cmdString := instanceTag + " get gain" + "\r"

	value, err := sendAndValidateResponse(socketKey, cmdString, "query", "number")

	if err != nil {
		return value, err
	}

	normalizedGain := unTransformVolume(value)

	framework.Log(function + " - Decoded Response: " + normalizedGain)

	// If we got here, the response was good, so successful return with the state indication
	return `"` + normalizedGain + `"`, nil
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

	value, err := sendAndValidateResponse(socketKey, cmdString, "query", "state")

	if err != nil {
		return value, err
	}

	framework.Log(function + " - Decoded Response: " + value)

	// If we got here, the response was good, so successful return with the state indication
	return `"` + value + `"`, nil
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
	if value == "\"true\"" {
		return `"` + strconv.Itoa(channel) + `"`, err
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

	value, err := sendAndValidateResponse(socketKey, cmdString, "query", "state")

	if err != nil {
		return value, err
	}

	framework.Log(function + " - Decoded Response: " + value)

	// If we got here, the response was good, so successful return with the state indication
	return `"` + value + `"`, nil
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

	value, err := sendAndValidateResponse(socketKey, cmdString, "command", "none")

	if err != nil {
		return value, err
	}

	framework.Log(function + " - Decoded Response: " + value)

	// If we got here, the response was good, so successful return with the state indication
	return "ok", nil
}

func setGain(socketKey string, instanceTag string, gain string) (string, error) {
	function := "setGain"

	value := "notok"
	err := error(nil)
	maxRetries := 2
	for maxRetries > 0 {
		value, err = setGainDo(socketKey, instanceTag, gain)
		if value != "ok" { // Something went wrong - perhaps try again
			framework.Log(function + " - fq3sdvc - retrying gain operation")
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

// Sets the gain for the specified instance tag. Takes a value from 0-100.
func setGainDo(socketKey string, instanceTag string, gain string) (string, error) {
	function := "setGainDo"
	gain = strings.Trim(gain, "\"")

	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected {
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - h3okxu3 - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	transformedGain := transformVolume(gain)
	framework.Log("Transformed Gain: " + transformedGain)

	cmdString := instanceTag + " set gain " + transformedGain + "\r"

	value, err := sendAndValidateResponse(socketKey, cmdString, "command", "none")

	if err != nil {
		return value, err
	}

	framework.Log(function + " - Decoded Response: " + value)

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

	value, err := sendAndValidateResponse(socketKey, cmdString, "command", "none")

	if err != nil {
		return value, err
	}

	framework.Log(function + " - Decoded Response: " + value)

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

	value, err := sendAndValidateResponse(socketKey, cmdString, "command", "none")

	if err != nil {
		return value, err
	}

	framework.Log(function + " - Decoded Response: " + value)

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

	value, err := sendAndValidateResponse(socketKey, cmdString, "command", "none")

	if err != nil {
		return value, err
	}

	framework.Log(function + " - Decoded Response: " + value)

	// If we got here, the response was good, so successful return with the state indication
	return "ok", nil
}

// Reports the health of the device.
func getHostname(socketKey string) (string, error) {
	function := "getHostname"

	connected := framework.CheckConnectionsMapExists(socketKey)
	if !connected {
		negotiation := loginNegotiation(socketKey)
		if !negotiation {
			errMsg := fmt.Sprintf(function + " - h3okxu35 - error connecting")
			framework.AddToErrors(socketKey, errMsg)
			return errMsg, errors.New(errMsg)
		}
	}

	cmdString := "DEVICE get hostname\r"

	value, err := sendAndValidateResponse(socketKey, cmdString, "command", "none")

	if err != nil {
		return value, err
	}

	framework.Log(function + " - Decoded Response: " + value)

	// If we got here, the response was good, so successful return with the state indication
	return `"` + value + `"`, nil
}

func healthCheck(socketKey string) (string, error) {
	returnStr := "true"
	_, err := getHostname(socketKey)
	if err != nil && (strings.Contains(err.Error(), "unable to send command") || strings.Contains(err.Error(), "error connecting")) {
		returnStr = "false"
	}
	return `"` + returnStr + `"`, nil
}
