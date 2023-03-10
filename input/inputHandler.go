package input

import (
	"fmt"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/quartermeat/go_rest_usb_api/assets"
	"github.com/quartermeat/go_rest_usb_api/console"
	"github.com/quartermeat/go_rest_usb_api/debuglog"
	"github.com/quartermeat/go_rest_usb_api/objects"
	"github.com/quartermeat/go_rest_usb_api/objects/domainObjects/card"
	"golang.org/x/exp/slices"
)

// InputHandler is a monolithic struct to handle user interactions with the app
type InputHandler struct {
	initialized  bool
	Cursor       *pixel.Sprite
	CursorAssets assets.ObjectAnimationAsset
	win          *pixelgl.Window
	cam          *pixel.Matrix
	consoleInput <-chan console.ITxTopic
	oldCamZoom   float64
}

func (input *InputHandler) setCursor(pressed bool) {

	if !pressed {
		input.Cursor = pixel.NewSprite(input.CursorAssets.Sheet, input.CursorAssets.Anims[input.CursorAssets.Description][0])
	} else {
		input.Cursor = pixel.NewSprite(input.CursorAssets.Sheet, input.CursorAssets.Anims[input.CursorAssets.Description][1])
	}

	input.initialized = true
}

func (input *InputHandler) handleConsole(someFlag bool, debugLog debuglog.Entries) debuglog.Entries {
	select {
	case consoleCommand := <-input.consoleInput:
		{
			if consoleCommand.GetTopicId() == console.Poke {
				someFlag = !someFlag
				input.setCursor(someFlag)
			}
			if consoleCommand.GetTopicId() == console.Stop {
				stopCommand := debuglog.Entry{
					Message: console.Stop,
				}
				debugLog = append(debugLog, stopCommand)
				return debugLog
			}
		}
	default:
		{
			//don't do anything
		}
	}
	return debugLog
}

// HandleInput is a super method ran from main
// atm: handles input from the keyboard, mouse and console
func (input *InputHandler) HandleInput(
	win *pixelgl.Window,
	cam *pixel.Matrix,
	gameCommands Commands,
	gameObjs *objects.GameObjects,
	objectAssets assets.ObjectAssets,
	dt float64,
	camSpeed float64,
	camZoom *float64,
	camZoomSpeed float64,
	camPos *pixel.Vec,
	drawHitBox *bool,
	readConsole <-chan console.ITxTopic,
) (debugLog debuglog.Entries) {
	//defaults
	var (
		cursorToggle bool
	)

	//do initialization of input handler
	if !input.initialized {
		//set window and cam
		input.win = win
		input.cam = cam
		*camZoom = 0.578704

		//set cursor
		cursorToggle = false
		var idx int = 0
		idx = slices.IndexFunc(objectAssets, func(c assets.IObjectAsset) bool {
			return objectAssets.IsDescriptionAvailable(CursorDesription)
		})
		if idx != -1 {
			input.CursorAssets = objectAssets[idx].(assets.ObjectAnimationAsset)
			input.setCursor(cursorToggle)
		} else {
			indexError := debuglog.Entry{
				Message: fmt.Sprintf("%s is not in assests", input.CursorAssets.Description),
			}
			fmt.Printf("%s is not in assests", input.CursorAssets.Description)
			debugLog = append(debugLog, indexError)
		}
	}

	input.consoleInput = readConsole
	debugLog = input.handleConsole(cursorToggle, debugLog)

	if win.MouseInsideWindow() {
		if !win.Pressed(pixelgl.KeyLeftControl) {
			win.SetCursorVisible(false)
			//setup and object to place
			input.Cursor.Draw(win, pixel.IM.Moved(cam.Unproject(win.MousePosition())))
		}
	} else {
		win.SetCursorVisible(true)
	}

	if win.JustReleased(pixelgl.MouseButtonLeft) && !win.Pressed(pixelgl.KeyLeftControl) {
		input.setCursor(false)
	}

	//handle ctrl functions
	if win.Pressed(pixelgl.KeyLeftControl) {
		input.setCursor(true)
		win.SetCursorVisible(true)
		if win.JustPressed(pixelgl.MouseButtonLeft) { //ctrl + left click
			mouse := cam.Unproject(win.MousePosition())
			gameCommands[fmt.Sprintf("SelectObjectAtPosition x:%f, y:%f", mouse.X, mouse.Y)] = SelectObjectAtPosition(gameObjs, mouse)
		}
	}

	//place the selected object
	if win.Pressed(pixelgl.MouseButtonLeft) && !win.Pressed(pixelgl.KeyLeftControl) {
		input.setCursor(true)
	}

	if win.JustPressed(pixelgl.Key0) {
		mouse := cam.Unproject(win.MousePosition())
		objectToPlace := card.NewCardObject(objectAssets.GetImage(card.CARD_BACK), mouse)
		gameCommands[fmt.Sprintf("AddObjectAtPosition: x:%f, y:%f, ObjectType:%s", mouse.X, mouse.Y, objectToPlace.ObjectName())] = AddObjectAtPosition(gameObjs, &objectToPlace, mouse)
	}

	//toggle global hit box draw for debugging
	if win.JustPressed(pixelgl.KeyH) {
		*drawHitBox = !*drawHitBox
	}

	//move camera
	if win.Pressed(pixelgl.KeyA) {
		camPos.X -= camSpeed * dt
	}
	if win.Pressed(pixelgl.KeyD) {
		camPos.X += camSpeed * dt
	}
	if win.Pressed(pixelgl.KeyS) {
		camPos.Y -= camSpeed * dt
	}
	if win.Pressed(pixelgl.KeyW) {
		camPos.Y += camSpeed * dt
	}

	// allow zoom on mouse scroll
	// newZoomFactor := math.Pow(camZoomSpeed, win.MouseScroll().Y)
	// //zoom camera
	// if newZoomFactor != input.oldCamZoom {
	// 	fmt.Printf("Old Cam zoom: %f\n", *camZoom)
	// 	*camZoom *= newZoomFactor
	// 	input.oldCamZoom = newZoomFactor
	// 	fmt.Printf("New Cam zoom: %f\n", *camZoom)
	// }

	return debugLog
}
