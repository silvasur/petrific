package objects

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type Snapshot struct {
	Tree      ObjectId
	Date      time.Time
	Container string
	Comment   string
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
	out = appendKVPair(out, "container", s.Container)
	out = appendKVPair(out, "date", s.Date.Format(time.RFC3339))
	out = appendKVPair(out, "tree", s.Tree.String())

	if s.Comment != "" {
		out = append(out, '\n')
		out = append(out, []byte(s.Comment)...)
	}

	return out
}

func (s *Snapshot) FromPayload(payload []byte) error {
	r := bytes.NewBuffer(payload)

	seenContainer := false
	seenDate := false
	seenTree := false

	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		line = strings.TrimSpace(line)

		if line == "" {
			break // Header read successfully
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Invalid header: %s", line)
		}

		headerval := strings.TrimSpace(parts[1])
		switch parts[0] {
		case "container":
			s.Container = headerval
			seenContainer = true
		case "date":
			d, err := time.Parse(time.RFC3339, headerval)
			if err != nil {
				return err
			}
			s.Date = d
			seenDate = true
		case "tree":
			oid, err := ParseObjectId(headerval)
			if err != nil {
				return err
			}
			s.Tree = oid
			seenTree = true
		}

		if err == io.EOF {
			break
		}
	}

	if !seenContainer || !seenDate || !seenTree {
		return errors.New("Missing container, date or tree header")
	}

	b := new(bytes.Buffer)
	if _, err := io.Copy(b, r); err != nil {
		return err
	}

	s.Comment = strings.TrimSpace(string(b.Bytes()))
	return nil
}

func (a Snapshot) Equals(b Snapshot) bool {
	return a.Tree.Equals(b.Tree) &&
		a.Container == b.Container &&
		a.Date.Equal(b.Date) &&
		a.Comment == b.Comment
}
