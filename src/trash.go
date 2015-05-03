// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package drive

import (
	"fmt"
)

func (g *Commands) Trash() (err error) {
	return g.reduceForTrash(g.opts.Sources, true, false)
}

func (g *Commands) Delete() (err error) {
	return g.reduceForTrash(g.opts.Sources, true, true)
}

func (g *Commands) Untrash() (err error) {
	return g.reduceForTrash(g.opts.Sources, false, false)
}

func (g *Commands) EmptyTrash() error {
	rootFile, err := g.rem.FindByPath("/")
	if err != nil {
		return err
	}

	spin := g.playabler()
	spin.play()
	defer spin.stop()

	travSt := traversalSt{
		depth:    -1,
		file:     rootFile,
		headPath: "/",
		inTrash:  true,
		mask:     g.opts.TypeMask,
	}

	if !g.breadthFirst(travSt, spin) {
		return nil
	}

	if g.opts.canPrompt() {
		g.log.Logln("This oepration is irreversible. Empty trash! ")

		if !promptForChanges() {
			g.log.Logln("Aborted emptying trash")
			return nil
		}
	}

	err = g.rem.EmptyTrash()
	if err == nil {
		g.log.Logln("Successfully emptied trash")
	}
	return err
}

func (g *Commands) trasher(relToRoot string, toTrash bool) (change *Change, err error) {
	var file *File
	if relToRoot == "/" && toTrash {
		return nil, fmt.Errorf("Will not try to trash root.")
	}
	if toTrash {
		file, err = g.rem.FindByPath(relToRoot)
	} else {
		file, err = g.rem.FindByPathTrashed(relToRoot)
	}

	if err != nil {
		return
	}

	change = &Change{Path: relToRoot}
	if toTrash {
		change.Dest = file
	} else {
		change.Src = file
	}
	return
}

func (g *Commands) trashByMatch(inTrash, permanent bool) error {
	matches, err := g.rem.FindMatches(g.opts.Path, g.opts.Sources, inTrash)
	if err != nil {
		return err
	}
	var cl []*Change
	p := g.opts.Path
	if p == "/" {
		p = ""
	}
	for match := range matches {
		if match == nil {
			continue
		}
		ch := &Change{Path: p + "/" + match.Name}
		if inTrash {
			ch.Src = match
		} else {
			ch.Dest = match
		}
		cl = append(cl, ch)
	}

	if len(cl) < 1 {
		return fmt.Errorf("no matches found!")
	}

	toTrash := !inTrash
	ok, _ := printChangeList(g.log, cl, !g.opts.canPrompt(), false)
	if !ok {
		return nil
	}

	return g.playTrashChangeList(cl, toTrash, permanent)
}

func (g *Commands) TrashByMatch() error {
	return g.trashByMatch(false, false)
}

func (g *Commands) UntrashByMatch() error {
	return g.trashByMatch(true, false)
}

func (g *Commands) DeleteByMatch() error {
	return g.trashByMatch(false, true)
}

func (g *Commands) reduceForTrash(args []string, toTrash, permanent bool) error {
	var cl []*Change
	for _, relToRoot := range args {
		c, cErr := g.trasher(relToRoot, toTrash)
		if cErr != nil {
			g.log.LogErrf("\033[91m'%s': %v\033[00m\n", relToRoot, cErr)
		} else if c != nil {
			cl = append(cl, c)
		}
	}

	ok, _ := printChangeList(g.log, cl, !g.opts.canPrompt(), false)
	if !ok {
		return nil
	}
	if permanent && g.opts.canPrompt() {
		if !promptForChanges("This operation is irreversible. Continue [Y/N] ") {
			return nil
		}
	}
	return g.playTrashChangeList(cl, toTrash, permanent)
}

func (g *Commands) playTrashChangeList(cl []*Change, toTrash, permanent bool) (err error) {
	trashSize, unTrashSize := reduceToSize(cl, SelectDest|SelectSrc)
	g.taskStart(trashSize + unTrashSize)

	var fn func(*Change) error
	if permanent {
		fn = g.remoteDelete
	} else {
		fn = g.remoteUntrash
		if toTrash {
			fn = g.remoteTrash
		}
	}

	for _, c := range cl {
		if c.Op() == OpNone {
			continue
		}

		cErr := fn(c)
		if cErr != nil {
			g.log.LogErrln(cErr)
		}
	}

	g.taskFinish()
	return err
}
