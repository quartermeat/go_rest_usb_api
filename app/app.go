// Package 'app' is the main game function and main loop
package app

import (
	"fmt"
	_ "image/png"
	"math/rand"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"

	"github.com/quartermeat/go_rest_usb_api/assets"
	"github.com/quartermeat/go_rest_usb_api/console"
	"github.com/quartermeat/go_rest_usb_api/debuglog"
	"github.com/quartermeat/go_rest_usb_api/input"
	"github.com/quartermeat/go_rest_usb_api/objects"
	"github.com/quartermeat/go_rest_usb_api/ui"
	"github.com/quartermeat/go_rest_usb_api/usbAPI"
)

// App() is the main game function and main loop for a card game.
// It sets up the window configuration, initializes the GUI, loads assets,
// seeds the random number generator, and starts a command server.
// It then enters a loop that handles delta time, handles input, updates game objects,
// draws game objects, draws the GUI, and draws a cursor based on selected object.
// At the end of each loop it also updates the window title with FPS and number of game objects.
// Finally it checks for any debug log entries with a message of 'console.Stop' and closes the window if found.
func App() {

	cfg := pixelgl.WindowConfig{
		Title:  APP_TITLE,
		Bounds: WINDOW_SIZE,
		VSync:  false,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	var (
		camPos             = pixel.ZV
		camSpeed           = 500.0
		camZoom            = 1.0
		camZoomSpeed       = 1.2
		gameObjs           objects.GameObjects
		gameCommands       = make(input.Commands)
		frames             = 0
		second             = time.NewTicker(time.Second)
		drawHitBox         = false
		inputHandler       input.InputHandler
		objectAssets       assets.ObjectAssets
		debugLog           debuglog.Entries
		sysErrors          []error
		consoleToInputChan chan console.ITxTopic
		gui                ui.GUI
	)

	consoleToInputChan = make(chan console.ITxTopic, 1)
	defer close(consoleToInputChan)

	// Start REST API server
	usbApi, err := usbAPI.NewUSBAPI()
	if err != nil {
		panic(err)
	}
	go func() {
		err := usbApi.Start()
		if err != nil {
			panic(err)
		}
	}()

	// start command server
	go console.StartServer(consoleToInputChan)
	if Test {
		go console.RunConsole()
	}

	//setup gui
	gui.InitGUI()

	//panic level errors
	sysErrors = make([]error, 0)

	//load assets
	objectAssets = loadAssets(sysErrors)

	//seed rng
	rand.Seed(time.Now().UnixNano())

	last := time.Now()
	for !win.Closed() {
		//handle delta
		dt := time.Since(last).Seconds()
		last = time.Now()

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		debugLog = inputHandler.HandleInput(
			win,
			&cam,
			gameCommands,
			&gameObjs,
			objectAssets,
			dt,
			camSpeed,
			&camZoom,
			camZoomSpeed,
			&camPos,
			&drawHitBox,
			consoleToInputChan,
		)

		var waitGroup sync.WaitGroup

		//handle game updates
		gui.UpdateGUI(gameCommands)
		gameCommands.ExecuteCommands(&waitGroup)
		waitGroup.Wait()
		gameObjs.UpdateAllObjects(dt, &waitGroup)
		waitGroup.Wait()

		win.Clear(colornames.Black)
		//draw game objects
		gameObjs.DrawAllObjects(win, drawHitBox, &waitGroup)
		waitGroup.Wait()

		gui.DrawGUI(win, &cam)

		//draw cursor based on selected object
		//must be done outside of inputHandler to be the last thing drawn
		if win.MouseInsideWindow() {
			if !win.Pressed(pixelgl.KeyLeftControl) {
				win.SetCursorVisible(false)
				//setup and object to place
				inputHandler.Cursor.Draw(win, pixel.IM.Moved(cam.Unproject(win.MousePosition())))
			}
		} else {
			win.SetCursorVisible(true)
		}

		win.Update()

		frames++
		select {
		case <-second.C:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d | GameObjects: %d", cfg.Title, frames, len(gameObjs)))
			frames = 0
		default:
		}

		for _, entry := range debugLog {
			fmt.Printf("debugLog: %s", entry.GetMessage())
			if entry.GetMessage() == console.Stop {
				//give time for graphics stuff finish
				time.Sleep(2 * time.Second)
				win.Destroy()
			}
		}
	}
}
