package functions

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	ToleranceExact    = "exact"
	ToleranceRelative = "rel:1e-9"
	RefusedText       = "refused"
	manifestFile      = "manifest.json"
)

type InputValue struct {
	Text  string
	Items []string
	List  bool
}

func (v *InputValue) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '[' {
		v.List = true
		v.Items = []string{}
		return json.Unmarshal(data, &v.Items)
	}
	return json.Unmarshal(data, &v.Text)
}

func (v InputValue) MarshalJSON() ([]byte, error) {
	if v.List {
		items := v.Items
		if items == nil {
			items = []string{}
		}
		return marshalPlain(items)
	}
	return marshalPlain(v.Text)
}

type Case struct {
	ID        string                `json:"id"`
	Kind      string                `json:"kind"`
	Inputs    map[string]InputValue `json:"inputs"`
	Expected  string                `json:"expected"`
	Tolerance string                `json:"tolerance"`
	Source    string                `json:"source"`
}

func marshalPlain(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buffer.Bytes(), "\n"), nil
}

func quoted(text string) string {
	encoded, _ := marshalPlain(text)
	return string(encoded)
}

func Canonical(c Case, order []string) string {
	var line strings.Builder
	line.WriteString(`{"id":` + quoted(c.ID))
	line.WriteString(`,"kind":` + quoted(c.Kind))
	line.WriteString(`,"inputs":{`)
	first := true
	for _, name := range order {
		value, given := c.Inputs[name]
		if !given {
			continue
		}
		if !first {
			line.WriteString(",")
		}
		first = false
		line.WriteString(quoted(name) + ":")
		if value.List {
			line.WriteString("[")
			for i, item := range value.Items {
				if i > 0 {
					line.WriteString(",")
				}
				line.WriteString(quoted(item))
			}
			line.WriteString("]")
		} else {
			line.WriteString(quoted(value.Text))
		}
	}
	line.WriteString(`},"expected":` + quoted(c.Expected))
	line.WriteString(`,"tolerance":` + quoted(c.Tolerance))
	line.WriteString(`,"source":` + quoted(c.Source) + "}")
	return line.String()
}

func Hash(c Case, order []string) string {
	sum := sha256.Sum256([]byte(Canonical(c, order)))
	return hex.EncodeToString(sum[:])
}

func Short(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}

func LoadCases(dir string) (map[string][]Case, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	versions := map[string][]Case{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".json") || name == manifestFile {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		var cases []Case
		if err := json.Unmarshal(raw, &cases); err != nil {
			return nil, err
		}
		versions[strings.TrimSuffix(name, ".json")] = cases
	}
	return versions, nil
}

func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
