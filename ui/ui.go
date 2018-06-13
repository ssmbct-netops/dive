package ui

import (
	"log"

	"github.com/jroimartin/gocui"
	"github.com/wagoodman/docker-image-explorer/filetree"
	"github.com/wagoodman/docker-image-explorer/image"
)

const debug = true

var Views struct {
	Tree   *FileTreeView
	Layer  *LayerView
	Status *StatusView
}

type View interface {
	Setup(*gocui.View) error
	CursorDown() error
	CursorUp() error
	Render() error
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == Views.Layer.Name {
		_, err := g.SetCurrentView(Views.Tree.Name)
		return err
	}
	_, err := g.SetCurrentView(Views.Layer.Name)
	return err
}

func CursorDown(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()

	// if there isn't a next line
	line, err := v.Line(cy + 1)
	if err != nil {
		// todo: handle error
	}
	if len(line) == 0 {
		return nil
	}
	if err := v.SetCursor(cx, cy+1); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+1); err != nil {
			return err
		}
	}
	return nil
}

func CursorUp(g *gocui.Gui, v *gocui.View) error {
	ox, oy := v.Origin()
	cx, cy := v.Cursor()
	if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
		if err := v.SetOrigin(ox, oy-1); err != nil {
			return err
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	//if err := g.SetKeybinding("main", gocui.MouseLeft, gocui.ModNone, toggleCollapse); err != nil {
	//	return err
	//}
	if err := g.SetKeybinding("side", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	splitCols := maxX / 2
	debugWidth := 0
	if debug {
		debugWidth = maxX / 4
	}
	debugCols := maxX - debugWidth
	bottomRows := 1
	if view, err := g.SetView(Views.Layer.Name, -1, -1, splitCols, maxY-bottomRows); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		Views.Layer.Setup(view)

	}
	if view, err := g.SetView(Views.Tree.Name, splitCols, -1, debugCols, maxY-bottomRows); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		Views.Tree.Setup(view)

		if _, err := g.SetCurrentView(Views.Tree.Name); err != nil {
			return err
		}
	}
	if debug {
		if _, err := g.SetView("debug", debugCols, -1, maxX, maxY-bottomRows); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
		}
	}
	if view, err := g.SetView(Views.Status.Name, -1, maxY-bottomRows-1, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		Views.Status.Setup(view)

	}

	return nil
}

func Run(layers []*image.Layer, refTrees []*filetree.FileTree) {

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	Views.Layer = NewLayerView("side", g, layers)
	Views.Tree = NewFileTreeView("main", g, filetree.StackRange(refTrees, 0), refTrees)
	Views.Status = NewStatusView("status", g)

	g.Cursor = false
	//g.Mouse = true
	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
