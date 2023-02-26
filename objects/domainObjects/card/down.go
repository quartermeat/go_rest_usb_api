package card

import (
	"fmt"

	"github.com/quartermeat/go_rest_usb_api/objects"
)

// OnAction represents the action executed on entering the On state.
type FlipUpAction struct{}

// Execute Flip Up action
func (a *FlipUpAction) Execute(gameObj objects.IGameObject) objects.EventType {
	gameObj.Sprite().Set(gameObj.GetAssets().GetSheet(), gameObj.GetAssets().GetImages()[TRASH])
	fmt.Println("Flip up action")
	return objects.NoOp
}
