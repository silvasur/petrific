package objects

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	snapshot_start_marker = "== BEGIN SNAPSHOT =="
	snapshot_end_marker   = "== END SNAPSHOT =="

	snapshot_start_line = snapshot_start_marker + "\n"
	snapshot_end_line   = snapshot_end_marker + "\n"
)

type Snapshot struct {
	Tree    ObjectId
	Date    time.Time
	Archive string
	Comment string
	Signed  bool
}

func (s Snapshot) Type() ObjectType {
	return OTSnapshot
}

func appendKVPair(b []byte, k, v string) []byte {
	b = append(b, []byte(k)...)
	b = append(b, ' ')
	b = append(b, []byte(v)...)
	b = append(b, '\n')
	return b
}

func (s Snapshot) Payload() (out []byte) {
	out = append(out, []byte(snapshot_start_line)...)
	out = appendKVPair(out, "archive", s.Archive)
	out = appendKVPair(out, "date", s.Date.Format(time.RFC3339))
	if s.Signed {
		out = appendKVPair(out, "signed", "yes")
	}
	out = appendKVPair(out, "tree", s.Tree.String())

	if s.Comment != "" {
		// We must prevent that the comment includes an end marker
		comment := strings.Replace(s.Comment, snapshot_end_marker, "~~ END SNAPSHOT ~~", -1)

		if comment[len(comment)-1] != '\n' {
			comment += "\n"
		}

		out = append(out, '\n')
		out = append(out, []byte(comment)...)
	}
	out = append(out, []byte(snapshot_end_line)...)

	return out
}

type Signer interface {
	Sign([]byte) ([]byte, error)
}

func (s Snapshot) SignedPayload(signer Signer) ([]byte, error) {
	return signer.Sign(s.Payload())
}

func (s *Snapshot) FromPayload(payload []byte) error {
	r := bytes.NewBuffer(payload)

	seenArchive := false
	seenDate := false
	seenTree := false

	start := false
	terminated := false
	comment := false

	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if !start {
			if line == snapshot_start_marker {
				start = true
			}
			continue
		}

		if line == snapshot_end_marker {
			terminated = true
			break
		}

		if comment {
			s.Comment += line + "\n"
			continue
		}

		if line == "" {
			comment = true
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Invalid header: %s", line)
		}

		headerval := strings.TrimSpace(parts[1])
		switch parts[0] {
		case "archive":
			s.Archive = headerval
			seenArchive = true
		case "date":
			d, err := time.Parse(time.RFC3339, headerval)
			if err != nil {
				return err
			}
			s.Date = d
			seenDate = true
		case "signed":
			s.Signed = headerval == "yes"
		case "tree":
			oid, err := ParseObjectId(headerval)
			if err != nil {
				return err
			}
			s.Tree = oid
			seenTree = true
		}
	}

	if !terminated {
		return errors.New("The snapshot was not properly terminated")
	}

	if !seenArchive || !seenDate || !seenTree {
		return errors.New("Missing archive, date or tree header")
	}

	s.Comment = strings.TrimSpace(s.Comment)
	return nil
}

func (a Snapshot) Equals(b Snapshot) bool {
	return a.Tree.Equals(b.Tree) &&
		a.Archive == b.Archive &&
		a.Date.Equal(b.Date) &&
		a.Comment == b.Comment
}
