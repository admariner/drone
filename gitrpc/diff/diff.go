// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package diff

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/harness/gitness/gitrpc/enum"
)

// Predefine []byte variables to avoid runtime allocations.
var (
	escapedSlash = []byte(`\\`)
	regularSlash = []byte(`\`)
	escapedTab   = []byte(`\t`)
	regularTab   = []byte("\t")
)

// LineType is the line type in diff.
type LineType uint8

// A list of different line types.
const (
	DiffLinePlain LineType = iota + 1
	DiffLineAdd
	DiffLineDelete
	DiffLineSection
)

// FileType is the file status in diff.
type FileType uint8

// A list of different file statuses.
const (
	FileAdd FileType = iota
	FileChange
	FileDelete
	FileRename
)

// Line represents a line in diff.
type Line struct {
	Type      LineType // The type of the line
	Content   string   // The content of the line
	LeftLine  int      // The left line number
	RightLine int      // The right line number
}

// Section represents a section in diff.
type Section struct {
	Lines []*Line // lines in the section

	numAdditions int
	numDeletions int
}

// NumLines returns the number of lines in the section.
func (s *Section) NumLines() int {
	return len(s.Lines)
}

// Line returns a specific line by given type and line number in a section.
func (s *Section) Line(lineType LineType, line int) *Line {
	var (
		difference      = 0
		addCount        = 0
		delCount        = 0
		matchedDiffLine *Line
	)

loop:
	for _, diffLine := range s.Lines {
		switch diffLine.Type {
		case DiffLineAdd:
			addCount++
		case DiffLineDelete:
			delCount++
		default:
			if matchedDiffLine != nil {
				break loop
			}
			difference = diffLine.RightLine - diffLine.LeftLine
			addCount = 0
			delCount = 0
		}

		switch lineType {
		case DiffLineDelete:
			if diffLine.RightLine == 0 && diffLine.LeftLine == line-difference {
				matchedDiffLine = diffLine
			}
		case DiffLineAdd:
			if diffLine.LeftLine == 0 && diffLine.RightLine == line+difference {
				matchedDiffLine = diffLine
			}
		}
	}

	if addCount == delCount {
		return matchedDiffLine
	}
	return nil
}

// File represents a file in diff.
type File struct {
	// The name and path of the file.
	Path string
	// The old name and path of the file.
	OldPath string
	// The type of the file.
	Type FileType
	// The index (SHA1 hash) of the file. For a changed/new file, it is the new SHA,
	// and for a deleted file it becomes "000000".
	SHA string
	// OldSHA is the old index (SHA1 hash) of the file.
	OldSHA string
	// The sections in the file.
	Sections []*Section

	numAdditions int
	numDeletions int

	mode    enum.EntryMode
	oldMode enum.EntryMode

	IsBinary    bool
	IsSubmodule bool
}

func (f *File) Status() string {
	switch {
	case f.Type == FileAdd:
		return "added"
	case f.Type == FileDelete:
		return "deleted"
	case f.Type == FileRename:
		return "renamed"
	case f.Type == FileChange:
		return "changed"
	default:
		return "unchanged"
	}
}

// NumSections returns the number of sections in the file.
func (f *File) NumSections() int {
	return len(f.Sections)
}

// NumAdditions returns the number of additions in the file.
func (f *File) NumAdditions() int {
	return f.numAdditions
}

// NumChanges returns the number of additions and deletions in the file.
func (f *File) NumChanges() int {
	return f.numAdditions + f.numDeletions
}

// NumDeletions returns the number of deletions in the file.
func (f *File) NumDeletions() int {
	return f.numDeletions
}

// Mode returns the mode of the file.
func (f *File) Mode() enum.EntryMode {
	return f.mode
}

// OldMode returns the old mode of the file if it's changed.
func (f *File) OldMode() enum.EntryMode {
	return f.oldMode
}

func (f *File) IsEmpty() bool {
	return f.Path == "" && f.OldPath == ""
}

type Parser struct {
	*bufio.Reader

	// The next line that hasn't been processed. It is used to determine what kind
	// of process should go in.
	buffer []byte
	isEOF  bool
}

func (p *Parser) readLine() error {
	if p.buffer != nil {
		return nil
	}

	var err error
	p.buffer, err = p.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			return fmt.Errorf("read string: %v", err)
		}

		p.isEOF = true
	}

	// Remove line break
	if len(p.buffer) > 0 && p.buffer[len(p.buffer)-1] == '\n' {
		p.buffer = p.buffer[:len(p.buffer)-1]
	}
	return nil
}

var diffHead = []byte("diff --git ")

func (p *Parser) parseFileHeader() (*File, error) {
	submoduleMode := " 160000"
	line := string(p.buffer)
	p.buffer = nil

	// NOTE: In case file name is surrounded by double quotes (it happens only in
	// git-shell). e.g. diff --git "a/xxx" "b/xxx"
	hasQuote := line[len(diffHead)] == '"'
	middle := strings.Index(line, ` b/`)
	if hasQuote {
		middle = strings.Index(line, ` "b/`)
	}

	beg := len(diffHead)
	a := line[beg+2 : middle]
	b := line[middle+3:]
	if hasQuote {
		a = string(UnescapeChars([]byte(a[1 : len(a)-1])))
		b = string(UnescapeChars([]byte(b[1 : len(b)-1])))
	}

	file := &File{
		Path:    a,
		OldPath: b,
		Type:    FileChange,
	}

	// Check file diff type and submodule
	var err error
checkType:
	for !p.isEOF {
		if err = p.readLine(); err != nil {
			return nil, err
		}

		line := string(p.buffer)
		p.buffer = nil

		if len(line) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, enum.DiffExtHeaderNewFileMode):
			file.Type = FileAdd
			file.IsSubmodule = strings.HasSuffix(line, submoduleMode)
			fields := strings.Fields(line)
			if len(fields) > 0 {
				mode, _ := strconv.ParseUint(fields[len(fields)-1], 8, 64)
				file.mode = enum.EntryMode(mode)
				if file.oldMode == 0 {
					file.oldMode = file.mode
				}
			}
		case strings.HasPrefix(line, enum.DiffExtHeaderDeletedFileMode):
			file.Type = FileDelete
			file.IsSubmodule = strings.HasSuffix(line, submoduleMode)
			fields := strings.Fields(line)
			if len(fields) > 0 {
				mode, _ := strconv.ParseUint(fields[len(fields)-1], 8, 64)
				file.mode = enum.EntryMode(mode)
				if file.oldMode == 0 {
					file.oldMode = file.mode
				}
			}
		case strings.HasPrefix(line, enum.DiffExtHeaderIndex): // e.g. index ee791be..9997571 100644
			fields := strings.Fields(line[6:])
			shas := strings.Split(fields[0], "..")
			if len(shas) != 2 {
				return nil, errors.New("malformed index: expect two SHAs in the form of <old>..<new>")
			}

			file.OldSHA = shas[0]
			file.SHA = shas[1]
			if len(fields) > 1 {
				mode, _ := strconv.ParseUint(fields[1], 8, 64)
				file.mode = enum.EntryMode(mode)
				file.oldMode = enum.EntryMode(mode)
			}
			break checkType
		case strings.HasPrefix(line, enum.DiffExtHeaderSimilarity):
			file.Type = FileRename
			file.OldPath = a
			file.Path = b

			// No need to look for index if it's a pure rename
			if strings.HasSuffix(line, "100%") {
				break checkType
			}
		case strings.HasPrefix(line, enum.DiffExtHeaderNewMode):
			fields := strings.Fields(line)
			if len(fields) > 0 {
				mode, _ := strconv.ParseUint(fields[len(fields)-1], 8, 64)
				file.mode = enum.EntryMode(mode)
			}
		case strings.HasPrefix(line, enum.DiffExtHeaderOldMode):
			fields := strings.Fields(line)
			if len(fields) > 0 {
				mode, _ := strconv.ParseUint(fields[len(fields)-1], 8, 64)
				file.oldMode = enum.EntryMode(mode)
			}
		}
	}

	return file, nil
}

func (p *Parser) parseSection() (*Section, error) {
	line := string(p.buffer)
	p.buffer = nil

	section := &Section{
		Lines: []*Line{
			{
				Type:    DiffLineSection,
				Content: line,
			},
		},
	}

	// Parse line number, e.g. @@ -0,0 +1,3 @@
	var leftLine, rightLine int
	ss := strings.Split(line, "@@")
	ranges := strings.Split(ss[1][1:], " ")
	leftLine, _ = strconv.Atoi(strings.Split(ranges[0], ",")[0][1:])
	if len(ranges) > 1 {
		rightLine, _ = strconv.Atoi(strings.Split(ranges[1], ",")[0])
	} else {
		rightLine = leftLine
	}

	var err error
	for !p.isEOF {
		if err = p.readLine(); err != nil {
			return nil, err
		}

		if len(p.buffer) == 0 {
			p.buffer = nil
			continue
		}

		// Make sure we're still in the section. If not, we're done with this section.
		if p.buffer[0] != ' ' &&
			p.buffer[0] != '+' &&
			p.buffer[0] != '-' {

			// No new line indicator
			if p.buffer[0] == '\\' &&
				bytes.HasPrefix(p.buffer, []byte(`\ No newline at end of file`)) {
				p.buffer = nil
				continue
			}
			return section, nil
		}

		line := string(p.buffer)
		p.buffer = nil

		switch line[0] {
		case ' ':
			section.Lines = append(section.Lines, &Line{
				Type:      DiffLinePlain,
				Content:   line,
				LeftLine:  leftLine,
				RightLine: rightLine,
			})
			leftLine++
			rightLine++
		case '+':
			section.Lines = append(section.Lines, &Line{
				Type:      DiffLineAdd,
				Content:   line,
				RightLine: rightLine,
			})
			section.numAdditions++
			rightLine++
		case '-':
			section.Lines = append(section.Lines, &Line{
				Type:     DiffLineDelete,
				Content:  line,
				LeftLine: leftLine,
			})
			section.numDeletions++
			if leftLine > 0 {
				leftLine++
			}
		}
	}

	return section, nil
}

func (p *Parser) Parse(f func(f *File)) error {
	file := new(File)
	currentFileLines := 0
	additions := 0
	deletions := 0

	var (
		err error
	)
	for !p.isEOF {
		if err = p.readLine(); err != nil {
			return err
		}

		if len(p.buffer) == 0 ||
			bytes.HasPrefix(p.buffer, []byte("+++ ")) ||
			bytes.HasPrefix(p.buffer, []byte("--- ")) {
			p.buffer = nil
			continue
		}

		// Found new file
		if bytes.HasPrefix(p.buffer, diffHead) {
			// stream previous file
			if !file.IsEmpty() && f != nil {
				f(file)
			}
			file, err = p.parseFileHeader()
			if err != nil {
				return err
			}

			currentFileLines = 0
			continue
		}

		if file == nil {
			p.buffer = nil
			continue
		}

		if bytes.HasPrefix(p.buffer, []byte("Binary")) {
			p.buffer = nil
			file.IsBinary = true
			continue
		}

		// Loop until we found section header
		if p.buffer[0] != '@' {
			p.buffer = nil
			continue
		}

		section, err := p.parseSection()
		if err != nil {
			return err
		}
		file.Sections = append(file.Sections, section)
		file.numAdditions += section.numAdditions
		file.numDeletions += section.numDeletions
		additions += section.numAdditions
		deletions += section.numDeletions
		currentFileLines += section.NumLines()
	}

	// stream last file
	if !file.IsEmpty() && f != nil {
		f(file)
	}

	return nil
}

// UnescapeChars reverses escaped characters.
func UnescapeChars(in []byte) []byte {
	if bytes.ContainsAny(in, "\\\t") {
		return in
	}

	out := bytes.Replace(in, escapedSlash, regularSlash, -1)
	out = bytes.Replace(out, escapedTab, regularTab, -1)
	return out
}
