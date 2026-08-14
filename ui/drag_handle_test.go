package ui

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestDragHandleMovesByRows(t *testing.T) {
	var gotID, gotFrom, gotTo int
	handle := newDragHandle(func(itemID, from, to int) {
		gotID, gotFrom, gotTo = itemID, from, to
	})
	handle.itemID = 7
	handle.index = 1
	handle.Dragged(&fyne.DragEvent{Dragged: fyne.Delta{DY: 2 * replayQueueRowHeight}})
	handle.DragEnd()

	if gotID != 7 || gotFrom != 1 || gotTo != 3 {
		t.Errorf("drop = (%d, %d, %d), want (7, 1, 3)", gotID, gotFrom, gotTo)
	}
}

func TestDragHandleResetsAfterDrop(t *testing.T) {
	handle := newDragHandle(func(int, int, int) {})
	handle.text.Move(fyne.NewPos(4, 8))
	handle.Dragged(&fyne.DragEvent{Dragged: fyne.Delta{DY: replayQueueRowHeight}})
	if got, want := handle.text.Position(), fyne.NewPos(4, 8+replayQueueRowHeight); got != want {
		t.Errorf("dragged position = %v, want %v", got, want)
	}

	handle.DragEnd()
	if got, want := handle.text.Position(), fyne.NewPos(4, 8); got != want {
		t.Errorf("reset position = %v, want %v", got, want)
	}
}
